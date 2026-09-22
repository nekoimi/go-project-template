//go:build integration

package integration_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/config"
	userdomain "github.com/nekoimi/go-project-template/internal/domain/user"
	"github.com/nekoimi/go-project-template/internal/pkg/database"
	"github.com/nekoimi/go-project-template/internal/repository"
	"github.com/nekoimi/go-project-template/internal/taskqueue"
)

func TestPostgresMigrationAndUniqueConstraints(t *testing.T) {
	databaseURL := envOrDefault("INTEGRATION_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/go_template_test?sslmode=disable")
	m := newMigrator(t, databaseURL)
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("reset migrations: %v", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("apply migrations: %v", err)
	}
	t.Cleanup(func() {
		_ = m.Down()
		_, _ = m.Close()
	})

	cfg := config.DefaultConfig().Database
	cfg.Host = "localhost"
	cfg.Port = "5432"
	cfg.User = "postgres"
	cfg.Password = "postgres"
	cfg.DBName = "go_template_test"
	db, err := database.NewPostgresDB(cfg, zap.NewNop(), "test")
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	repo := repository.NewUserRepository(db)
	ctx := context.Background()
	first := &userdomain.User{ID: 1, Username: "integration-one", Email: "integration@example.com", Password: "hash"}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first user: %v", err)
	}
	duplicate := &userdomain.User{ID: 2, Username: "integration-two", Email: first.Email, Password: "hash"}
	if err := repo.Create(ctx, duplicate); err == nil {
		t.Fatal("expected duplicate email constraint error")
	}
}

func TestTaskQueueDeliveryAndShutdown(t *testing.T) {
	cfg := config.DefaultConfig().TaskQueue
	cfg.Enabled = true
	cfg.Redis.Addr = envOrDefault("INTEGRATION_REDIS_ADDR", "localhost:6379")
	cfg.Concurrency = 1
	cfg.Queues = map[string]int{"integration": 1}
	cfg.ShutdownTimeout = 5 * time.Second

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if err := redisClient.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flush redis: %v", err)
	}
	t.Cleanup(func() { _ = redisClient.Close() })

	worker := taskqueue.NewWorker(cfg, zap.NewNop())
	delivered := make(chan struct{}, 1)
	if err := worker.Handle("integration:delivery", func(context.Context, []byte) error {
		delivered <- struct{}{}
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}
	if err := worker.Start(); err != nil {
		t.Fatalf("start worker: %v", err)
	}
	var stopOnce sync.Once
	stopWorker := func() { stopOnce.Do(worker.Shutdown) }
	t.Cleanup(stopWorker)

	client := taskqueue.NewClient(cfg.Redis)
	t.Cleanup(func() { _ = client.Close() })
	_, err := client.Enqueue(context.Background(), taskqueue.Task{
		Type:    "integration:delivery",
		Payload: []byte(`{"ok":true}`),
	}, taskqueue.EnqueueOptions{Queue: "integration", MaxRetry: 0, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("enqueue task: %v", err)
	}

	select {
	case <-delivered:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for task delivery")
	}
	stopWorker()
}

func newMigrator(t *testing.T, databaseURL string) *migrate.Migrate {
	t.Helper()
	migrations, err := fs.Sub(os.DirFS(".."), "migrations")
	if err != nil {
		t.Fatalf("open migrations: %v", err)
	}
	source, err := iofs.New(migrations, ".")
	if err != nil {
		t.Fatalf("create migration source: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		_ = source.Close()
		t.Fatalf("create migrator: %v", err)
	}
	return m
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
