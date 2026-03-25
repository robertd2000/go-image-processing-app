package config

import (
	"errors"
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
	Port    string `mapstructure:"SERVER_PORT"`
	RunMode string `mapstructure:"SERVER_RUN_MODE"`
	Domain  string `mapstructure:"SERVER_DOMAIN"`
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
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

type JWTConfig struct {
	Secret        string `mapstructure:"JWT_SECRET"`
	AccessTTLMin  int    `mapstructure:"JWT_ACCESS_TTL_MIN"`
	RefreshTTLMin int    `mapstructure:"JWT_REFRESH_TTL_MIN"`
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	v := viper.New()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	v.SetConfigName("config-" + env)
	v.SetConfigType("yml")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// config файл теперь опциональный
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// 🔥 ЖЁСТКО прокидываем env (убираем магию Viper)
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	cfg.Postgres = PostgresConfig{
		Host:     os.Getenv("AUTH_DB_HOST"),
		Port:     os.Getenv("AUTH_DB_PORT"),
		User:     os.Getenv("AUTH_DB_USER"),
		Password: os.Getenv("AUTH_DB_PASS"),
		DBName:   os.Getenv("AUTH_DB_NAME"),
		SSLMode:  os.Getenv("AUTH_DB_SSL_MODE"),
	}

	cfg.Server.Port = os.Getenv("SERVER_PORT")

	// PORT override (docker/platform)
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
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
