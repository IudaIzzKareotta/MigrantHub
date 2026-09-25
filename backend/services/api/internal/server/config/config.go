package config

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App  AppConfig
	HTTP HTTPConfig
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

var validEnvironments = map[string]bool{
	"dev":        true,
	"staging":    true,
	"production": true,
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

	return nil
}
