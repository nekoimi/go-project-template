package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func Load(configPath string) (*Config, error) {
	return load(configPath, nil)
}

func LoadFor(configPath string, requirements ValidationRequirements) (*Config, error) {
	return load(configPath, &requirements)
}

func load(configPath string, requirements *ValidationRequirements) (*Config, error) {
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
	_ = v.BindEnv("database.connect_timeout", "DATABASE_CONNECT_TIMEOUT")
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("server.timezone", "TZ")
	_ = v.BindEnv("snowflake.node_id", "SNOWFLAKE_NODE_ID")
	_ = v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	_ = v.BindEnv("server.read_header_timeout", "SERVER_READ_HEADER_TIMEOUT")
	_ = v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	_ = v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")
	_ = v.BindEnv("task_queue.enabled", "TASK_QUEUE_ENABLED")
	_ = v.BindEnv("task_queue.redis.addr", "REDIS_ADDR")
	_ = v.BindEnv("task_queue.redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("task_queue.redis.db", "REDIS_DB")
	_ = v.BindEnv("task_queue.concurrency", "TASK_QUEUE_CONCURRENCY")
	_ = v.BindEnv("scheduler.leader_election", "SCHEDULER_LEADER_ELECTION")
	_ = v.BindEnv("scheduler.leader_key", "SCHEDULER_LEADER_KEY")
	_ = v.BindEnv("scheduler.leader_ttl", "SCHEDULER_LEADER_TTL")

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
	_ = v.BindEnv("storage.s3.create_bucket", "S3_CREATE_BUCKET")
	_ = v.BindEnv("storage.s3.startup_timeout", "S3_STARTUP_TIMEOUT")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	applyOperationalDefaults(cfg)
	cfg.Storage.Normalize()
	var validateErr error
	if requirements == nil {
		validateErr = cfg.Validate()
	} else {
		validateErr = cfg.ValidateFor(*requirements)
	}
	if validateErr != nil {
		return nil, fmt.Errorf("validate config: %w", validateErr)
	}

	return cfg, nil
}

func applyOperationalDefaults(cfg *Config) {
	defaults := DefaultConfig()
	if cfg.Server.ReadTimeout <= 0 {
		cfg.Server.ReadTimeout = defaults.Server.ReadTimeout
	}
	if cfg.Server.ReadHeaderTimeout <= 0 {
		cfg.Server.ReadHeaderTimeout = defaults.Server.ReadHeaderTimeout
	}
	if cfg.Server.WriteTimeout <= 0 {
		cfg.Server.WriteTimeout = defaults.Server.WriteTimeout
	}
	if cfg.Server.IdleTimeout <= 0 {
		cfg.Server.IdleTimeout = defaults.Server.IdleTimeout
	}
	if cfg.Database.ConnectTimeout <= 0 {
		cfg.Database.ConnectTimeout = defaults.Database.ConnectTimeout
	}
	if cfg.Scheduler.LeaderKey == "" {
		cfg.Scheduler.LeaderKey = defaults.Scheduler.LeaderKey
	}
	if cfg.Scheduler.LeaderTTL <= 0 {
		cfg.Scheduler.LeaderTTL = defaults.Scheduler.LeaderTTL
	}
	if cfg.Storage.Upload.MaxRequestSize <= 0 {
		cfg.Storage.Upload.MaxRequestSize = defaults.Storage.Upload.MaxRequestSize
	}
	if cfg.Storage.Upload.MaxFiles <= 0 {
		cfg.Storage.Upload.MaxFiles = defaults.Storage.Upload.MaxFiles
	}
	if cfg.Storage.S3.StartupTimeout <= 0 {
		cfg.Storage.S3.StartupTimeout = defaults.Storage.S3.StartupTimeout
	}
	if cfg.Storage.Minio.StartupTimeout <= 0 {
		cfg.Storage.Minio.StartupTimeout = defaults.Storage.S3.StartupTimeout
	}
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Enabled:           true,
			Port:              "8080",
			Mode:              "debug",
			Timezone:          "Asia/Shanghai",
			ShutdownTimeout:   10,
			ReadTimeout:       2 * time.Minute,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      2 * time.Minute,
			IdleTimeout:       60 * time.Second,
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
			ConnectTimeout:  10 * time.Second,
		},
		JWT: JWTConfig{
			Secret:      "change-me-in-production",
			ExpireHours: 72,
		},
		Scheduler: SchedulerConfig{
			Enabled:        true,
			Timezone:       "Asia/Shanghai",
			LeaderElection: true,
			LeaderKey:      "go-project-template:scheduler:leader",
			LeaderTTL:      15 * time.Second,
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
				MaxFileSize:    10,
				MaxRequestSize: 50,
				MaxFiles:       10,
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
				CreateBucket:   true,
				StartupTimeout: 10 * time.Second,
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
