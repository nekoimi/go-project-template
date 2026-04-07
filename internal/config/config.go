package config

import "time"

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Websocket WebsocketConfig `mapstructure:"websocket"`
	Storage   StorageConfig   `mapstructure:"storage"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug / release
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
	Driver  string          `mapstructure:"driver"` // local / minio
	BaseURL string          `mapstructure:"base_url"`
	Local   LocalConfig     `mapstructure:"local"`
	Minio   MinioConfig     `mapstructure:"minio"`
}

type LocalConfig struct {
	UploadDir   string `mapstructure:"upload_dir"`
	MaxFileSize int    `mapstructure:"max_file_size"` // MB
}

type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	PublicURL string `mapstructure:"public_url"`
}
