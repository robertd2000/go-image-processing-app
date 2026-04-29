package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Kafka    KafkaConfig
	Storage  StorageConfig
	Log      LogConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port    string `mapstructure:"port"`
	RunMode string `mapstructure:"run_mode"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type KafkaConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Brokers []string `mapstructure:"brokers"`
	GroupID string   `mapstructure:"group_id"`
	Topics  Topics   `mapstructure:"topics"`
}

type Topics struct {
	ImageProcessingRequested string `mapstructure:"image_processing_requested"`
	ImageProcessed           string `mapstructure:"image_processed"` // future
}

type StorageConfig struct {
	Type      string `mapstructure:"type"` // s3 | local
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User,
		p.Password,
		p.Host,
		p.Port,
		p.DBName,
		p.SSLMode,
	)
}

func Load() (*Config, error) {
	v := viper.New()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	v.SetConfigFile("./config/config-" + env + ".yml")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.Kafka.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers is empty")
	}

	cfg.Server.Port = normalizePort(cfg.Server.Port)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func normalizePort(p string) string {
	if p == "" {
		return ":8081"
	}
	if strings.HasPrefix(p, ":") {
		return p
	}
	return ":" + p
}

func (c *Config) Validate() error {
	if c.Postgres.Host == "" {
		return fmt.Errorf("postgres host is required")
	}
	if c.Postgres.User == "" {
		return fmt.Errorf("postgres user is required")
	}
	if c.Postgres.DBName == "" {
		return fmt.Errorf("postgres dbname is required")
	}

	if c.Storage.Type == "" {
		return fmt.Errorf("storage type is required")
	}
	if c.Storage.Type == "s3" {
		if c.Storage.Endpoint == "" {
			return fmt.Errorf("storage endpoint is required")
		}
		if c.Storage.Bucket == "" {
			return fmt.Errorf("storage bucket is required")
		}
	}

	if c.Kafka.Enabled {
		if len(c.Kafka.Brokers) == 0 {
			return fmt.Errorf("kafka brokers are required")
		}
		if c.Kafka.Topics.ImageProcessingRequested == "" {
			return fmt.Errorf("processing topic is required")
		}
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	return nil
}
