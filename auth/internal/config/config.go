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
	JWT      JWTConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port    string `mapstructure:"port"`
	RunMode string `mapstructure:"run_mode"`
	Domain  string `mapstructure:"domain"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	AccessTTLMin  int    `mapstructure:"access_ttl_min"`
	RefreshTTLMin int    `mapstructure:"refresh_ttl_min"`
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

	v.SetConfigName("config-" + env)
	v.SetConfigType("yaml")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// ENV override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.Server.Port = normalizePort(cfg.Server.Port)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func normalizePort(p string) string {
	if p == "" {
		return ":8080"
	}
	if strings.HasPrefix(p, ":") {
		return p
	}
	return ":" + p
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	if c.Postgres.Host == "" {
		return fmt.Errorf("postgres host is required")
	}
	if c.Postgres.User == "" {
		return fmt.Errorf("postgres user is required")
	}
	if c.Postgres.DBName == "" {
		return fmt.Errorf("postgres dbname is required")
	}
	return nil
}
