package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	DB      PostgresConfig
	Kafka   KafkaConfig
	Storage StorageConfig
	Image   ImageConfig
	JWT     JWTConfig
	Log     LogConfig
}

type ServerConfig struct {
	InternalPort string
	ExternalPort string
	RunMode      string
	Domain       string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type KafkaConfig struct {
	Brokers         []string
	TaskTopic       string
	ResultTopic     string
	ConsumerGroup   string
	RetryTopic      string
	DeadLetterTopic string
}

type StorageConfig struct {
	Type      string
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type ImageConfig struct {
	MaxUploadSizeMB int
	AllowedFormats  []string
	DefaultQuality  int
	Workers         int
}

type JWTConfig struct {
	Secret        string
	AccessTTLMin  int
	RefreshTTLMin int
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {

	v := viper.New()

	v.SetConfigName(getConfigPath(os.Getenv("APP_ENV")))
	v.SetConfigType("yml")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./internal/config")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.ExternalPort = port
	} else {
		cfg.Server.ExternalPort = cfg.Server.InternalPort
	}

	return &cfg, nil
}

func getConfigPath(env string) string {

	switch env {

	case "docker":
		return "config-docker"

	case "production":
		return "config-production"

	default:
		return "config-development"
	}
}
