package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_LegacyMinIOConfigNormalizesToS3(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	content := `
server:
  port: "8080"
  mode: debug
database:
  host: from-yaml
  port: "5432"
  user: postgres
  password: postgres
  dbname: go_template
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 30
jwt:
  secret: from-yaml-secret
  expire_hours: 72
scheduler:
  enabled: false
  timezone: "Asia/Shanghai"
snowflake:
  node_id: 1
rate_limit:
  enabled: false
  rps: 100
  burst: 200
websocket:
  enabled: false
storage:
  driver: minio
  minio:
    endpoint: "localhost:9000"
    access_key: "${MINIO_ACCESS_KEY}"
    secret_key: "${MINIO_SECRET_KEY}"
    bucket: go-template
    use_ssl: false
    public_url: "http://localhost:9000"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MINIO_ACCESS_KEY", "env-access")
	t.Setenv("MINIO_SECRET_KEY", "env-secret")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Storage.Minio.AccessKey != "env-access" {
		t.Fatalf("AccessKey = %q, want env-access", cfg.Storage.Minio.AccessKey)
	}
	if cfg.Storage.Minio.SecretKey != "env-secret" {
		t.Fatalf("SecretKey = %q, want env-secret", cfg.Storage.Minio.SecretKey)
	}
	if cfg.Storage.Driver != "s3" {
		t.Fatalf("Driver = %q, want s3", cfg.Storage.Driver)
	}
	if cfg.Storage.S3.Provider != "minio" {
		t.Fatalf("Provider = %q, want minio", cfg.Storage.S3.Provider)
	}
	if cfg.Storage.S3.AccessKey != "env-access" || cfg.Storage.S3.SecretKey != "env-secret" {
		t.Fatalf("S3 credentials = %q/%q, want env overrides", cfg.Storage.S3.AccessKey, cfg.Storage.S3.SecretKey)
	}
}

func TestLoad_S3BindEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	content := `
storage:
  driver: s3
  s3:
    provider: rustfs
    endpoint: "localhost:9000"
    access_key: from-yaml
    secret_key: from-yaml
    bucket: go-template
    region: us-east-1
    use_ssl: false
    force_path_style: true
    public_url: "http://localhost:9000"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("S3_ACCESS_KEY", "rustfs-access")
	t.Setenv("S3_SECRET_KEY", "rustfs-secret")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Storage.S3.Provider != "rustfs" {
		t.Fatalf("Provider = %q, want rustfs", cfg.Storage.S3.Provider)
	}
	if cfg.Storage.S3.AccessKey != "rustfs-access" || cfg.Storage.S3.SecretKey != "rustfs-secret" {
		t.Fatalf("S3 credentials = %q/%q, want env overrides", cfg.Storage.S3.AccessKey, cfg.Storage.S3.SecretKey)
	}
}

func TestLoad_databaseBindEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	content := `
server:
  port: "8080"
  mode: debug
database:
  host: placeholder
  port: "5432"
  user: postgres
  password: postgres
  dbname: go_template
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 30
jwt:
  secret: x
  expire_hours: 72
scheduler:
  enabled: false
snowflake:
  node_id: 1
rate_limit:
  enabled: false
websocket:
  enabled: false
storage:
  driver: local
  local:
    upload_dir: ./uploads
    max_file_size: 10
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DATABASE_HOST", "db.example")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Database.Host != "db.example" {
		t.Fatalf("Host = %q, want db.example", cfg.Database.Host)
	}
	if cfg.Storage.Upload.MaxFileSize != 10 {
		t.Fatalf("Upload.MaxFileSize = %d, want legacy local value 10", cfg.Storage.Upload.MaxFileSize)
	}
}

func TestLoad_ProjectConfigs(t *testing.T) {
	t.Setenv("DATABASE_HOST", "postgres")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_USER", "postgres")
	t.Setenv("DATABASE_PASSWORD", "postgres")
	t.Setenv("DATABASE_NAME", "go_template")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("S3_ACCESS_KEY", "access-key")
	t.Setenv("S3_SECRET_KEY", "secret-key")
	t.Setenv("REDIS_ADDR", "redis:6379")

	paths := []string{
		filepath.Join("..", "..", "config", "config.dev.yaml"),
		filepath.Join("..", "..", "config", "config.test.yaml"),
		filepath.Join("..", "..", "config", "config.prod.yaml"),
	}
	for _, path := range paths {
		if _, err := Load(path); err != nil {
			t.Errorf("Load(%q): %v", path, err)
		}
	}
}
