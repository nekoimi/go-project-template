package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/nekoimi/go-project-template/internal/config"
)

const usageText = `Usage:
  go run ./cmd/migrate [flags] up [steps]
  go run ./cmd/migrate [flags] down [steps|all]
  go run ./cmd/migrate [flags] version
  go run ./cmd/migrate [flags] goto <version>
  go run ./cmd/migrate [flags] force <version>

Flags:
`

type options struct {
	configPath  string
	migrations  string
	databaseURL string
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) (resultErr error) {
	opts, command, commandArgs, err := parseArgs(args, stderr)
	if err != nil {
		return err
	}

	databaseURL, err := resolveDatabaseURL(opts)
	if err != nil {
		return err
	}

	m, migrationPath, err := newMigrator(databaseURL, opts.migrations)
	if err != nil {
		return err
	}
	defer func() {
		sourceErr, databaseErr := m.Close()
		resultErr = errors.Join(resultErr, sourceErr, databaseErr)
	}()

	if _, err := fmt.Fprintf(stdout, "migration path: %s\n", migrationPath); err != nil {
		return fmt.Errorf("write migration path: %w", err)
	}
	if err := execute(m, command, commandArgs, stdout); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			_, writeErr := fmt.Fprintln(stdout, "no migration changes")
			return writeErr
		}
		return err
	}
	return nil
}

func parseArgs(args []string, stderr io.Writer) (options, string, []string, error) {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := options{}
	fs.StringVar(&opts.configPath, "config", "config/config.dev.yaml", "application config file")
	fs.StringVar(&opts.migrations, "path", "migrations", "migration SQL directory")
	fs.StringVar(&opts.databaseURL, "database-url", "", "database URL (overrides config and MIGRATE_DATABASE_URL)")
	fs.Usage = func() {
		_, _ = fmt.Fprint(stderr, usageText)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return options{}, "", nil, err
	}
	positionals := fs.Args()
	if len(positionals) == 0 {
		fs.Usage()
		return options{}, "", nil, errors.New("command is required")
	}
	return opts, strings.ToLower(positionals[0]), positionals[1:], nil
}

func resolveDatabaseURL(opts options) (string, error) {
	if value := strings.TrimSpace(opts.databaseURL); value != "" {
		return value, nil
	}
	if value := strings.TrimSpace(os.Getenv("MIGRATE_DATABASE_URL")); value != "" {
		return value, nil
	}
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return "", fmt.Errorf("load config %q: %w", opts.configPath, err)
	}
	return databaseURL(cfg.Database), nil
}

func databaseURL(cfg config.DatabaseConfig) string {
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     "/" + cfg.DBName,
		RawQuery: query.Encode(),
	}
	return u.String()
}

func newMigrator(databaseURL, migrationsPath string) (*migrate.Migrate, string, error) {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, "", fmt.Errorf("resolve migration path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, "", fmt.Errorf("open migration path %q: %w", absPath, err)
	}
	if !info.IsDir() {
		return nil, "", fmt.Errorf("migration path %q is not a directory", absPath)
	}
	source, err := iofs.New(os.DirFS(absPath), ".")
	if err != nil {
		return nil, "", fmt.Errorf("create migration source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		_ = source.Close()
		return nil, "", fmt.Errorf("create migrator: %w", err)
	}
	return m, absPath, nil
}

func execute(m *migrate.Migrate, command string, args []string, stdout io.Writer) error {
	switch command {
	case "up":
		if len(args) == 0 {
			return m.Up()
		}
		steps, err := positiveIntArg("up", args)
		if err != nil {
			return err
		}
		return m.Steps(steps)
	case "down":
		if len(args) == 0 {
			return m.Steps(-1)
		}
		if len(args) == 1 && strings.EqualFold(args[0], "all") {
			return m.Down()
		}
		steps, err := positiveIntArg("down", args)
		if err != nil {
			return err
		}
		return m.Steps(-steps)
	case "version":
		if len(args) != 0 {
			return errors.New("version does not accept arguments")
		}
		version, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			_, writeErr := fmt.Fprintln(stdout, "version: none (database is clean)")
			return writeErr
		}
		if err != nil {
			return err
		}
		_, writeErr := fmt.Fprintf(stdout, "version: %d, dirty: %t\n", version, dirty)
		return writeErr
	case "goto":
		version, err := uintArg("goto", args)
		if err != nil {
			return err
		}
		return m.Migrate(version)
	case "force":
		version, err := intArg("force", args)
		if err != nil {
			return err
		}
		if version < -1 {
			return errors.New("force version must be -1 or greater")
		}
		return m.Force(version)
	default:
		return fmt.Errorf("unsupported command %q", command)
	}
}

func positiveIntArg(command string, args []string) (int, error) {
	value, err := intArg(command, args)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s steps must be greater than zero", command)
	}
	return value, nil
}

func intArg(command string, args []string) (int, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("%s requires exactly one integer argument", command)
	}
	value, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q: %w", command, args[0], err)
	}
	return value, nil
}

func uintArg(command string, args []string) (uint, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("%s requires exactly one version argument", command)
	}
	value, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s version %q: %w", command, args[0], err)
	}
	if uint64(uint(value)) != value {
		return 0, fmt.Errorf("%s version %q overflows uint", command, args[0])
	}
	return uint(value), nil
}
