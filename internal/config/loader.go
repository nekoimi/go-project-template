package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)

	// 环境变量绑定
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 数据库环境变量绑定
	_ = v.BindEnv("database.host", "DATABASE_HOST")
	_ = v.BindEnv("database.port", "DATABASE_PORT")
	_ = v.BindEnv("database.user", "DATABASE_USER")
	_ = v.BindEnv("database.password", "DATABASE_PASSWORD")
	_ = v.BindEnv("database.dbname", "DATABASE_NAME")
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("server.timezone", "TZ")
	_ = v.BindEnv("snowflake.node_id", "SNOWFLAKE_NODE_ID")
	_ = v.BindEnv("task_queue.enabled", "TASK_QUEUE_ENABLED")
	_ = v.BindEnv("task_queue.redis.addr", "REDIS_ADDR")
	_ = v.BindEnv("task_queue.redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("task_queue.redis.db", "REDIS_DB")
	_ = v.BindEnv("task_queue.concurrency", "TASK_QUEUE_CONCURRENCY")

	_ = v.BindEnv("storage.minio.access_key", "MINIO_ACCESS_KEY")
	_ = v.BindEnv("storage.minio.secret_key", "MINIO_SECRET_KEY")
	_ = v.BindEnv("storage.minio.endpoint", "MINIO_ENDPOINT")
	_ = v.BindEnv("storage.minio.public_url", "MINIO_PUBLIC_URL")
	_ = v.BindEnv("storage.minio.bucket", "MINIO_BUCKET")

	// Generic S3-compatible object storage. MINIO_* remains a fallback so
	// existing deployments can migrate without rotating environment names.
	_ = v.BindEnv("storage.s3.access_key", "S3_ACCESS_KEY", "MINIO_ACCESS_KEY")
	_ = v.BindEnv("storage.s3.secret_key", "S3_SECRET_KEY", "MINIO_SECRET_KEY")
	_ = v.BindEnv("storage.s3.endpoint", "S3_ENDPOINT", "MINIO_ENDPOINT")
	_ = v.BindEnv("storage.s3.public_url", "S3_PUBLIC_URL", "MINIO_PUBLIC_URL")
	_ = v.BindEnv("storage.s3.bucket", "S3_BUCKET", "MINIO_BUCKET")
	_ = v.BindEnv("storage.s3.region", "S3_REGION")
	_ = v.BindEnv("storage.s3.use_ssl", "S3_USE_SSL")
	_ = v.BindEnv("storage.s3.force_path_style", "S3_FORCE_PATH_STYLE")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	cfg.Storage.Normalize()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Enabled:         true,
			Port:            "8080",
			Mode:            "debug",
			Timezone:        "Asia/Shanghai",
			ShutdownTimeout: 10,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			User:            "postgres",
			Password:        "postgres",
			DBName:          "go_template",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 30,
		},
		JWT: JWTConfig{
			Secret:      "change-me-in-production",
			ExpireHours: 72,
		},
		Scheduler: SchedulerConfig{
			Enabled:  true,
			Timezone: "Asia/Shanghai",
		},
		TaskQueue: TaskQueueConfig{
			Enabled:         false,
			Concurrency:     10,
			ShutdownTimeout: 30 * time.Second,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"ai":       1,
			},
			Redis: RedisConfig{
				Addr: "localhost:6379",
				DB:   0,
			},
		},
		Snowflake: SnowflakeConfig{
			NodeID: 1,
		},
		RateLimit: RateLimitConfig{
			Enabled: false,
			RPS:     100,
			Burst:   200,
		},
		Websocket: WebsocketConfig{
			Enabled:         false,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			PingPeriod:      60 * 1e9, // 60s
			WriteWait:       10 * 1e9, // 10s
			ReadWait:        60 * 1e9, // 60s
			MaxMessageSize:  5120,
		},
		Storage: StorageConfig{
			Driver:  "local",
			BaseURL: "http://localhost:8080/uploads",
			Upload: UploadConfig{
				MaxFileSize: 10,
			},
			Local: LocalConfig{
				UploadDir: "./uploads",
			},
			S3: S3Config{
				Provider:       "minio",
				Endpoint:       "localhost:9000",
				AccessKey:      "minioadmin",
				SecretKey:      "minioadmin",
				Bucket:         "go-template",
				Region:         "us-east-1",
				UseSSL:         false,
				ForcePathStyle: true,
				PublicURL:      "http://localhost:9000",
			},
		},
		Modules: ModulesConfig{
			"auth": {
				Enabled: true,
			},
			"user": {
				Enabled: true,
			},
			"upload": {
				Enabled: true,
			},
			"websocket": {
				Enabled: true,
			},
			"example_job": {
				Enabled: true,
			},
		},
	}
}
