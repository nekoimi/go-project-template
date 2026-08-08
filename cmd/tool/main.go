package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	userdomain "github.com/nekoimi/go-project-template/internal/domain/user"
	"github.com/nekoimi/go-project-template/internal/pkg/database"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/pkg/timeutil"
	"github.com/nekoimi/go-project-template/internal/repository"
)

func main() {
	app := &cli.App{
		Name:  "tool",
		Usage: "project maintenance and administration tools",
		Commands: []*cli.Command{
			{
				Name:  "user",
				Usage: "user account operations",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "create a user account",
						Action: createUser,
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "config", Value: "config/config.dev.yaml", Usage: "application config file", EnvVars: []string{"APP_CONFIG"}},
							&cli.StringFlag{Name: "username", Usage: "username", EnvVars: []string{"APP_USER_USERNAME"}, Required: true},
							&cli.StringFlag{Name: "email", Usage: "email", EnvVars: []string{"APP_USER_EMAIL"}, Required: true},
							&cli.StringFlag{Name: "password", Usage: "password", EnvVars: []string{"APP_USER_PASSWORD"}, Required: true},
						},
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("tool failed: %v", err)
	}
}

func createUser(c *cli.Context) error {
	username := strings.TrimSpace(c.String("username"))
	email := strings.TrimSpace(strings.ToLower(c.String("email")))
	password := c.String("password")
	if len(username) < 3 || len(username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if email == "" || !strings.Contains(email, "@") {
		return errors.New("email must be a valid email address")
	}
	if len(password) < 6 || len(password) > 50 {
		return errors.New("password must be between 6 and 50 characters")
	}

	cfg, err := config.Load(c.String("config"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if err := timeutil.SetGlobalLocation(cfg.Server.Timezone); err != nil {
		return fmt.Errorf("set timezone: %w", err)
	}
	if err := idgen.Init(cfg.Snowflake.NodeID); err != nil {
		return fmt.Errorf("initialize id generator: %w", err)
	}

	db, err := database.NewPostgresDB(cfg.Database, zap.NewNop(), cfg.Server.Mode)
	if err != nil {
		return err
	}
	defer func() {
		if sqlDB, closeErr := db.DB(); closeErr == nil {
			_ = sqlDB.Close()
		}
	}()

	user, err := createUserRecord(c.Context, db, username, email, password)
	if err != nil {
		return err
	}
	fmt.Printf("user created: id=%d username=%s email=%s\n", user.ID, user.Username, user.Email)
	return nil
}

func createUserRecord(ctx context.Context, db *gorm.DB, username, email, password string) (*userdomain.User, error) {
	repo := repository.NewUserRepository(db)
	if _, err := repo.FindByEmail(ctx, email); err == nil {
		return nil, fmt.Errorf("user with email %q already exists; promote it with SQL instead of creating a duplicate", email)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if _, err := repo.FindByUsername(ctx, username); err == nil {
		return nil, fmt.Errorf("user with username %q already exists; promote it with SQL instead of creating a duplicate", username)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check username: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := &userdomain.User{Username: username, Email: email, Password: string(hashed)}
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return repo.WithTx(tx).Create(ctx, user)
	}); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}
