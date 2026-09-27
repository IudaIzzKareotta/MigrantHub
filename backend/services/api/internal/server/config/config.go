package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App  AppConfig
	HTTP HTTPConfig
	DB   DBConfig
	Auth AuthConfig
}

type AppConfig struct {
	Environment string
}

type HTTPConfig struct {
	Host              string
	Port              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func (c HTTPConfig) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

func (c DBConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   "/" + c.Name,
	}

	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()

	return u.String()
}

type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

const (
	minJWTSecretLength = 32
	devJWTSecret       = "dev-only-insecure-secret-please-change-in-production"
)

var validEnvironments = map[string]bool{
	"dev":        true,
	"staging":    true,
	"production": true,
}

var validSSLModes = map[string]bool{
	"disable":     true,
	"allow":       true,
	"prefer":      true,
	"require":     true,
	"verify-ca":   true,
	"verify-full": true,
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	v.SetDefault("APP_ENV", "dev")
	v.SetDefault("HTTP_HOST", "0.0.0.0")
	v.SetDefault("HTTP_PORT", "8080")
	v.SetDefault("HTTP_READ_HEADER_TIMEOUT", "5s")
	v.SetDefault("HTTP_READ_TIMEOUT", "10s")
	v.SetDefault("HTTP_WRITE_TIMEOUT", "10s")
	v.SetDefault("HTTP_IDLE_TIMEOUT", "60s")

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "migranthub")
	v.SetDefault("DB_PASSWORD", "migranthub")
	v.SetDefault("DB_NAME", "migranthub")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_CONNS", "10")
	v.SetDefault("DB_MIN_CONNS", "2")
	v.SetDefault("DB_MAX_CONN_LIFETIME", "30m")
	v.SetDefault("DB_MAX_CONN_IDLE_TIME", "5m")
	v.SetDefault("DB_CONNECT_TIMEOUT", "5s")

	v.SetDefault("AUTH_JWT_SECRET", devJWTSecret)
	v.SetDefault("AUTH_ACCESS_TOKEN_TTL", "15m")
	v.SetDefault("AUTH_REFRESH_TOKEN_TTL", "720h")

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{
		App: AppConfig{
			Environment: v.GetString("APP_ENV"),
		},
		HTTP: HTTPConfig{
			Host: v.GetString("HTTP_HOST"),
			Port: v.GetString("HTTP_PORT"),
		},
		DB: DBConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetString("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			Name:     v.GetString("DB_NAME"),
			SSLMode:  v.GetString("DB_SSLMODE"),
		},
		Auth: AuthConfig{
			JWTSecret: v.GetString("AUTH_JWT_SECRET"),
		},
	}

	var err error
	if cfg.HTTP.ReadHeaderTimeout, err = time.ParseDuration(v.GetString("HTTP_READ_HEADER_TIMEOUT")); err != nil {
		return nil, fmt.Errorf("parse HTTP_READ_HEADER_TIMEOUT: %w", err)
	}
	if cfg.HTTP.ReadTimeout, err = time.ParseDuration(v.GetString("HTTP_READ_TIMEOUT")); err != nil {
		return nil, fmt.Errorf("parse HTTP_READ_TIMEOUT: %w", err)
	}
	if cfg.HTTP.WriteTimeout, err = time.ParseDuration(v.GetString("HTTP_WRITE_TIMEOUT")); err != nil {
		return nil, fmt.Errorf("parse HTTP_WRITE_TIMEOUT: %w", err)
	}
	if cfg.HTTP.IdleTimeout, err = time.ParseDuration(v.GetString("HTTP_IDLE_TIMEOUT")); err != nil {
		return nil, fmt.Errorf("parse HTTP_IDLE_TIMEOUT: %w", err)
	}

	maxConns, err := strconv.ParseInt(v.GetString("DB_MAX_CONNS"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse DB_MAX_CONNS: %w", err)
	}
	cfg.DB.MaxConns = int32(maxConns)

	minConns, err := strconv.ParseInt(v.GetString("DB_MIN_CONNS"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse DB_MIN_CONNS: %w", err)
	}
	cfg.DB.MinConns = int32(minConns)

	if cfg.DB.MaxConnLifetime, err = time.ParseDuration(v.GetString("DB_MAX_CONN_LIFETIME")); err != nil {
		return nil, fmt.Errorf("parse DB_MAX_CONN_LIFETIME: %w", err)
	}
	if cfg.DB.MaxConnIdleTime, err = time.ParseDuration(v.GetString("DB_MAX_CONN_IDLE_TIME")); err != nil {
		return nil, fmt.Errorf("parse DB_MAX_CONN_IDLE_TIME: %w", err)
	}
	if cfg.DB.ConnectTimeout, err = time.ParseDuration(v.GetString("DB_CONNECT_TIMEOUT")); err != nil {
		return nil, fmt.Errorf("parse DB_CONNECT_TIMEOUT: %w", err)
	}

	if cfg.Auth.AccessTokenTTL, err = time.ParseDuration(v.GetString("AUTH_ACCESS_TOKEN_TTL")); err != nil {
		return nil, fmt.Errorf("parse AUTH_ACCESS_TOKEN_TTL: %w", err)
	}
	if cfg.Auth.RefreshTokenTTL, err = time.ParseDuration(v.GetString("AUTH_REFRESH_TOKEN_TTL")); err != nil {
		return nil, fmt.Errorf("parse AUTH_REFRESH_TOKEN_TTL: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if !validEnvironments[c.App.Environment] {
		return fmt.Errorf("APP_ENV: invalid value %q", c.App.Environment)
	}

	if c.HTTP.Host == "" {
		return errors.New("HTTP_HOST: must not be empty")
	}

	port, err := strconv.Atoi(c.HTTP.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("HTTP_PORT: invalid port %q", c.HTTP.Port)
	}

	for name, d := range map[string]time.Duration{
		"HTTP_READ_HEADER_TIMEOUT": c.HTTP.ReadHeaderTimeout,
		"HTTP_READ_TIMEOUT":        c.HTTP.ReadTimeout,
		"HTTP_WRITE_TIMEOUT":       c.HTTP.WriteTimeout,
		"HTTP_IDLE_TIMEOUT":        c.HTTP.IdleTimeout,
	} {
		if d <= 0 {
			return fmt.Errorf("%s: must be greater than zero", name)
		}
	}

	if c.DB.Host == "" {
		return errors.New("DB_HOST: must not be empty")
	}

	dbPort, err := strconv.Atoi(c.DB.Port)
	if err != nil || dbPort < 1 || dbPort > 65535 {
		return fmt.Errorf("DB_PORT: invalid port %q", c.DB.Port)
	}

	if c.DB.User == "" {
		return errors.New("DB_USER: must not be empty")
	}
	if c.DB.Name == "" {
		return errors.New("DB_NAME: must not be empty")
	}
	if !validSSLModes[c.DB.SSLMode] {
		return fmt.Errorf("DB_SSLMODE: invalid value %q", c.DB.SSLMode)
	}
	if c.DB.MaxConns < 1 {
		return fmt.Errorf("DB_MAX_CONNS: must be at least 1, got %d", c.DB.MaxConns)
	}
	if c.DB.MinConns < 0 || c.DB.MinConns > c.DB.MaxConns {
		return fmt.Errorf("DB_MIN_CONNS: must be between 0 and DB_MAX_CONNS (%d), got %d", c.DB.MaxConns, c.DB.MinConns)
	}

	for name, d := range map[string]time.Duration{
		"DB_MAX_CONN_LIFETIME":  c.DB.MaxConnLifetime,
		"DB_MAX_CONN_IDLE_TIME": c.DB.MaxConnIdleTime,
		"DB_CONNECT_TIMEOUT":    c.DB.ConnectTimeout,
	} {
		if d <= 0 {
			return fmt.Errorf("%s: must be greater than zero", name)
		}
	}

	if len(c.Auth.JWTSecret) < minJWTSecretLength {
		return fmt.Errorf("AUTH_JWT_SECRET: must be at least %d characters", minJWTSecretLength)
	}
	if c.App.Environment == "production" && c.Auth.JWTSecret == devJWTSecret {
		return errors.New("AUTH_JWT_SECRET: must not use the default development secret in production")
	}
	if c.Auth.AccessTokenTTL <= 0 {
		return errors.New("AUTH_ACCESS_TOKEN_TTL: must be greater than zero")
	}
	if c.Auth.RefreshTokenTTL <= c.Auth.AccessTokenTTL {
		return errors.New("AUTH_REFRESH_TOKEN_TTL: must be greater than AUTH_ACCESS_TOKEN_TTL")
	}

	return nil
}
