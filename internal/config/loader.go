package config

import (
	"strings"

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

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Mode: "debug",
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
			Enabled:  false,
			Timezone: "Asia/Shanghai",
		},
		Websocket: WebsocketConfig{
			Enabled:         false,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			PingPeriod:      60 * 1e9,  // 60s
			WriteWait:       10 * 1e9,  // 10s
			ReadWait:        60 * 1e9,  // 60s
			MaxMessageSize:  5120,
		},
		Storage: StorageConfig{
			Driver:  "local",
			BaseURL: "http://localhost:8080/uploads",
			Local: LocalConfig{
				UploadDir:   "./uploads",
				MaxFileSize: 10,
			},
			Minio: MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				Bucket:    "go-template",
				UseSSL:    false,
				PublicURL: "http://localhost:9000",
			},
		},
	}
}
