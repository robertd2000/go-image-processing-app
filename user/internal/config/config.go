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
	Log      LogConfig
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
	Brokers []string `mapstructure:"brokers"`
	Topics  Topics   `mapstructure:"topics"`
	GroupID string   `mapstructure:"group_id"`
}

type Topics struct {
	UserCreated string `mapstructure:"user_created"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
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
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}
	if c.Kafka.Topics.UserCreated == "" {
		return fmt.Errorf("kafka user_created topic is required")
	}
	if c.Kafka.GroupID == "" {
		return fmt.Errorf("kafka group_id is required")
	}
	return nil
}
