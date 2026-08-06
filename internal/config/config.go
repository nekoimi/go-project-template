package config

import (
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Snowflake SnowflakeConfig `mapstructure:"snowflake"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Websocket WebsocketConfig `mapstructure:"websocket"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Modules   ModulesConfig   `mapstructure:"modules"`
}

type ModulesConfig map[string]ModuleConfig

type ModuleConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

func (c *Config) ModuleEnabled(name string) bool {
	if c == nil {
		return false
	}
	if c.Modules == nil {
		return true
	}
	module, ok := c.Modules[name]
	if !ok {
		return true
	}
	return module.Enabled
}

type SnowflakeConfig struct {
	NodeID int64 `mapstructure:"node_id"`
}

type RateLimitConfig struct {
	Enabled bool    `mapstructure:"enabled"`
	RPS     float64 `mapstructure:"rps"`
	Burst   int     `mapstructure:"burst"`
}

type ServerConfig struct {
	Enabled         bool     `mapstructure:"enabled"`
	Port            string   `mapstructure:"port"`
	Mode            string   `mapstructure:"mode"` // debug / release
	Timezone        string   `mapstructure:"timezone"`
	ShutdownTimeout int      `mapstructure:"shutdown_timeout"` // 秒
	AllowedOrigins  []string `mapstructure:"allowed_origins"`  // CORS/WebSocket 允许的来源，空则允许全部
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            string `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 分钟
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type SchedulerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Timezone string `mapstructure:"timezone"`
}

type WebsocketConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	ReadBufferSize  int           `mapstructure:"read_buffer_size"`
	WriteBufferSize int           `mapstructure:"write_buffer_size"`
	PingPeriod      time.Duration `mapstructure:"ping_period"`
	WriteWait       time.Duration `mapstructure:"write_wait"`
	ReadWait        time.Duration `mapstructure:"read_wait"`
	MaxMessageSize  int64         `mapstructure:"max_message_size"`
}

type StorageConfig struct {
	Driver  string       `mapstructure:"driver"` // local / s3
	BaseURL string       `mapstructure:"base_url"`
	Upload  UploadConfig `mapstructure:"upload"`
	Local   LocalConfig  `mapstructure:"local"`
	S3      S3Config     `mapstructure:"s3"`
	// Minio is kept for backward compatibility with existing configuration
	// files. It is normalized into S3 when driver is set to minio.
	Minio S3Config `mapstructure:"minio"`
}

type LocalConfig struct {
	UploadDir string `mapstructure:"upload_dir"`
	// Deprecated: use StorageConfig.Upload. These fields remain readable so
	// older YAML files can be normalized during migration.
	MaxFileSize  int      `mapstructure:"max_file_size"`
	AllowedExts  []string `mapstructure:"allowed_exts"`
	AllowedMIMEs []string `mapstructure:"allowed_mimes"`
}

type UploadConfig struct {
	MaxFileSize  int      `mapstructure:"max_file_size"` // MB
	AllowedExts  []string `mapstructure:"allowed_exts"`
	AllowedMIMEs []string `mapstructure:"allowed_mimes"`
}

type S3Config struct {
	Provider       string `mapstructure:"provider"` // minio / rustfs / aws
	Endpoint       string `mapstructure:"endpoint"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	Bucket         string `mapstructure:"bucket"`
	Region         string `mapstructure:"region"`
	UseSSL         bool   `mapstructure:"use_ssl"`
	ForcePathStyle bool   `mapstructure:"force_path_style"`
	PublicURL      string `mapstructure:"public_url"`
}

// Normalize converts legacy storage settings to the protocol-oriented S3
// configuration and moves upload policy out of the local filesystem driver.
func (c *StorageConfig) Normalize() {
	if c == nil {
		return
	}

	c.Driver = strings.ToLower(strings.TrimSpace(c.Driver))
	if c.Driver == "" {
		c.Driver = "local"
	}

	if c.Driver == "minio" {
		c.Driver = "s3"
		c.S3 = c.Minio
		if c.S3.Provider == "" {
			c.S3.Provider = "minio"
		}
	}

	if c.Driver == "s3" && c.S3.Provider == "" {
		c.S3.Provider = "s3"
	}

	if c.Local.MaxFileSize > 0 {
		c.Upload.MaxFileSize = c.Local.MaxFileSize
	}
	if len(c.Local.AllowedExts) > 0 {
		c.Upload.AllowedExts = append([]string(nil), c.Local.AllowedExts...)
	}
	if len(c.Local.AllowedMIMEs) > 0 {
		c.Upload.AllowedMIMEs = append([]string(nil), c.Local.AllowedMIMEs...)
	}
}
