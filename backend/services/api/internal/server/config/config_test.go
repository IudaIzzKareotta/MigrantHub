package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Environment != "dev" {
		t.Errorf("Environment = %q, want %q", cfg.App.Environment, "dev")
	}
	if cfg.HTTP.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want %q", cfg.HTTP.Host, "0.0.0.0")
	}
	if cfg.HTTP.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.HTTP.Port, "8080")
	}
	if cfg.HTTP.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want %v", cfg.HTTP.ReadHeaderTimeout, 5*time.Second)
	}
	if cfg.HTTP.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", cfg.HTTP.ReadTimeout, 10*time.Second)
	}
	if cfg.HTTP.WriteTimeout != 10*time.Second {
		t.Errorf("WriteTimeout = %v, want %v", cfg.HTTP.WriteTimeout, 10*time.Second)
	}
	if cfg.HTTP.IdleTimeout != 60*time.Second {
		t.Errorf("IdleTimeout = %v, want %v", cfg.HTTP.IdleTimeout, 60*time.Second)
	}
	if cfg.DB.Host != "localhost" {
		t.Errorf("DB.Host = %q, want %q", cfg.DB.Host, "localhost")
	}
	if cfg.DB.Port != "5432" {
		t.Errorf("DB.Port = %q, want %q", cfg.DB.Port, "5432")
	}
	if cfg.DB.User != "migranthub" {
		t.Errorf("DB.User = %q, want %q", cfg.DB.User, "migranthub")
	}
	if cfg.DB.Name != "migranthub" {
		t.Errorf("DB.Name = %q, want %q", cfg.DB.Name, "migranthub")
	}
	if cfg.DB.SSLMode != "disable" {
		t.Errorf("DB.SSLMode = %q, want %q", cfg.DB.SSLMode, "disable")
	}
	if cfg.DB.MaxConns != 10 {
		t.Errorf("DB.MaxConns = %d, want %d", cfg.DB.MaxConns, 10)
	}
	if cfg.DB.MinConns != 2 {
		t.Errorf("DB.MinConns = %d, want %d", cfg.DB.MinConns, 2)
	}
	if cfg.DB.MaxConnLifetime != 30*time.Minute {
		t.Errorf("DB.MaxConnLifetime = %v, want %v", cfg.DB.MaxConnLifetime, 30*time.Minute)
	}
	if cfg.DB.MaxConnIdleTime != 5*time.Minute {
		t.Errorf("DB.MaxConnIdleTime = %v, want %v", cfg.DB.MaxConnIdleTime, 5*time.Minute)
	}
	if cfg.DB.ConnectTimeout != 5*time.Second {
		t.Errorf("DB.ConnectTimeout = %v, want %v", cfg.DB.ConnectTimeout, 5*time.Second)
	}
	if cfg.Auth.JWTSecret != devJWTSecret {
		t.Errorf("Auth.JWTSecret = %q, want the default dev secret", cfg.Auth.JWTSecret)
	}
	if cfg.Auth.AccessTokenTTL != 15*time.Minute {
		t.Errorf("Auth.AccessTokenTTL = %v, want %v", cfg.Auth.AccessTokenTTL, 15*time.Minute)
	}
	if cfg.Auth.RefreshTokenTTL != 720*time.Hour {
		t.Errorf("Auth.RefreshTokenTTL = %v, want %v", cfg.Auth.RefreshTokenTTL, 720*time.Hour)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_HOST", "127.0.0.1")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "1s")
	t.Setenv("HTTP_READ_TIMEOUT", "2s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "3s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "4s")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_USER", "app")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "migranthub_test")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("DB_MAX_CONNS", "20")
	t.Setenv("DB_MIN_CONNS", "5")
	t.Setenv("DB_MAX_CONN_LIFETIME", "1h")
	t.Setenv("DB_MAX_CONN_IDLE_TIME", "10m")
	t.Setenv("DB_CONNECT_TIMEOUT", "2s")
	t.Setenv("AUTH_JWT_SECRET", "a-properly-random-32-plus-char-secret-for-tests")
	t.Setenv("AUTH_ACCESS_TOKEN_TTL", "5m")
	t.Setenv("AUTH_REFRESH_TOKEN_TTL", "1h")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Environment != "production" {
		t.Errorf("Environment = %q, want %q", cfg.App.Environment, "production")
	}
	if cfg.HTTP.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want %q", cfg.HTTP.Host, "127.0.0.1")
	}
	if cfg.HTTP.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.HTTP.Port, "9090")
	}
	if cfg.HTTP.ReadHeaderTimeout != 1*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want %v", cfg.HTTP.ReadHeaderTimeout, 1*time.Second)
	}
	if cfg.HTTP.ReadTimeout != 2*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", cfg.HTTP.ReadTimeout, 2*time.Second)
	}
	if cfg.HTTP.WriteTimeout != 3*time.Second {
		t.Errorf("WriteTimeout = %v, want %v", cfg.HTTP.WriteTimeout, 3*time.Second)
	}
	if cfg.HTTP.IdleTimeout != 4*time.Second {
		t.Errorf("IdleTimeout = %v, want %v", cfg.HTTP.IdleTimeout, 4*time.Second)
	}
	if cfg.DB.Host != "db.internal" {
		t.Errorf("DB.Host = %q, want %q", cfg.DB.Host, "db.internal")
	}
	if cfg.DB.Port != "6543" {
		t.Errorf("DB.Port = %q, want %q", cfg.DB.Port, "6543")
	}
	if cfg.DB.User != "app" {
		t.Errorf("DB.User = %q, want %q", cfg.DB.User, "app")
	}
	if cfg.DB.Password != "secret" {
		t.Errorf("DB.Password = %q, want %q", cfg.DB.Password, "secret")
	}
	if cfg.DB.Name != "migranthub_test" {
		t.Errorf("DB.Name = %q, want %q", cfg.DB.Name, "migranthub_test")
	}
	if cfg.DB.SSLMode != "require" {
		t.Errorf("DB.SSLMode = %q, want %q", cfg.DB.SSLMode, "require")
	}
	if cfg.DB.MaxConns != 20 {
		t.Errorf("DB.MaxConns = %d, want %d", cfg.DB.MaxConns, 20)
	}
	if cfg.DB.MinConns != 5 {
		t.Errorf("DB.MinConns = %d, want %d", cfg.DB.MinConns, 5)
	}
	if cfg.DB.MaxConnLifetime != 1*time.Hour {
		t.Errorf("DB.MaxConnLifetime = %v, want %v", cfg.DB.MaxConnLifetime, 1*time.Hour)
	}
	if cfg.DB.MaxConnIdleTime != 10*time.Minute {
		t.Errorf("DB.MaxConnIdleTime = %v, want %v", cfg.DB.MaxConnIdleTime, 10*time.Minute)
	}
	if cfg.DB.ConnectTimeout != 2*time.Second {
		t.Errorf("DB.ConnectTimeout = %v, want %v", cfg.DB.ConnectTimeout, 2*time.Second)
	}
	if cfg.Auth.JWTSecret != "a-properly-random-32-plus-char-secret-for-tests" {
		t.Errorf("Auth.JWTSecret = %q, want the overridden secret", cfg.Auth.JWTSecret)
	}
	if cfg.Auth.AccessTokenTTL != 5*time.Minute {
		t.Errorf("Auth.AccessTokenTTL = %v, want %v", cfg.Auth.AccessTokenTTL, 5*time.Minute)
	}
	if cfg.Auth.RefreshTokenTTL != 1*time.Hour {
		t.Errorf("Auth.RefreshTokenTTL = %v, want %v", cfg.Auth.RefreshTokenTTL, 1*time.Hour)
	}
}

func TestLoadInvalidDuration(t *testing.T) {
	t.Setenv("HTTP_READ_TIMEOUT", "not-a-duration")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "HTTP_READ_TIMEOUT") {
		t.Errorf("Load() error = %v, want mention of HTTP_READ_TIMEOUT", err)
	}
}

func TestLoadInvalidEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "not-a-real-env")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Errorf("Load() error = %v, want mention of APP_ENV", err)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("HTTP_PORT", "not-a-port")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "HTTP_PORT") {
		t.Errorf("Load() error = %v, want mention of HTTP_PORT", err)
	}
}

func TestLoadInvalidMaxConns(t *testing.T) {
	t.Setenv("DB_MAX_CONNS", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "DB_MAX_CONNS") {
		t.Errorf("Load() error = %v, want mention of DB_MAX_CONNS", err)
	}
}

func TestHTTPConfigAddr(t *testing.T) {
	cfg := HTTPConfig{Host: "0.0.0.0", Port: "8080"}

	if got, want := cfg.Addr(), "0.0.0.0:8080"; got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
	}
}

func TestDBConfigDSN(t *testing.T) {
	cfg := DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "app",
		Password: "p@ss/word",
		Name:     "migranthub",
		SSLMode:  "disable",
	}

	want := "postgres://app:p%40ss%2Fword@localhost:5432/migranthub?sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestConfigValidate(t *testing.T) {
	valid := func() Config {
		return Config{
			App: AppConfig{Environment: "dev"},
			HTTP: HTTPConfig{
				Host:              "0.0.0.0",
				Port:              "8080",
				ReadHeaderTimeout: 5 * time.Second,
				ReadTimeout:       10 * time.Second,
				WriteTimeout:      10 * time.Second,
				IdleTimeout:       60 * time.Second,
			},
			DB: DBConfig{
				Host:            "localhost",
				Port:            "5432",
				User:            "migranthub",
				Password:        "migranthub",
				Name:            "migranthub",
				SSLMode:         "disable",
				MaxConns:        10,
				MinConns:        2,
				MaxConnLifetime: 30 * time.Minute,
				MaxConnIdleTime: 5 * time.Minute,
				ConnectTimeout:  5 * time.Second,
			},
			Auth: AuthConfig{
				JWTSecret:       "a-properly-random-32-plus-char-secret-for-tests",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 720 * time.Hour,
			},
		}
	}

	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name:    "valid config",
			mutate:  func(c *Config) {},
			wantErr: "",
		},
		{
			name:    "invalid environment",
			mutate:  func(c *Config) { c.App.Environment = "nope" },
			wantErr: "APP_ENV",
		},
		{
			name:    "empty host",
			mutate:  func(c *Config) { c.HTTP.Host = "" },
			wantErr: "HTTP_HOST",
		},
		{
			name:    "non-numeric port",
			mutate:  func(c *Config) { c.HTTP.Port = "abc" },
			wantErr: "HTTP_PORT",
		},
		{
			name:    "port out of range",
			mutate:  func(c *Config) { c.HTTP.Port = "70000" },
			wantErr: "HTTP_PORT",
		},
		{
			name:    "zero read header timeout",
			mutate:  func(c *Config) { c.HTTP.ReadHeaderTimeout = 0 },
			wantErr: "HTTP_READ_HEADER_TIMEOUT",
		},
		{
			name:    "negative idle timeout",
			mutate:  func(c *Config) { c.HTTP.IdleTimeout = -1 * time.Second },
			wantErr: "HTTP_IDLE_TIMEOUT",
		},
		{
			name:    "empty db host",
			mutate:  func(c *Config) { c.DB.Host = "" },
			wantErr: "DB_HOST",
		},
		{
			name:    "invalid db port",
			mutate:  func(c *Config) { c.DB.Port = "70000" },
			wantErr: "DB_PORT",
		},
		{
			name:    "empty db user",
			mutate:  func(c *Config) { c.DB.User = "" },
			wantErr: "DB_USER",
		},
		{
			name:    "empty db name",
			mutate:  func(c *Config) { c.DB.Name = "" },
			wantErr: "DB_NAME",
		},
		{
			name:    "invalid sslmode",
			mutate:  func(c *Config) { c.DB.SSLMode = "yes-please" },
			wantErr: "DB_SSLMODE",
		},
		{
			name:    "zero max conns",
			mutate:  func(c *Config) { c.DB.MaxConns = 0 },
			wantErr: "DB_MAX_CONNS",
		},
		{
			name:    "min conns greater than max conns",
			mutate:  func(c *Config) { c.DB.MinConns = c.DB.MaxConns + 1 },
			wantErr: "DB_MIN_CONNS",
		},
		{
			name:    "zero max conn lifetime",
			mutate:  func(c *Config) { c.DB.MaxConnLifetime = 0 },
			wantErr: "DB_MAX_CONN_LIFETIME",
		},
		{
			name:    "zero connect timeout",
			mutate:  func(c *Config) { c.DB.ConnectTimeout = 0 },
			wantErr: "DB_CONNECT_TIMEOUT",
		},
		{
			name:    "short jwt secret",
			mutate:  func(c *Config) { c.Auth.JWTSecret = "too-short" },
			wantErr: "AUTH_JWT_SECRET",
		},
		{
			name: "dev jwt secret in production",
			mutate: func(c *Config) {
				c.App.Environment = "production"
				c.Auth.JWTSecret = devJWTSecret
			},
			wantErr: "AUTH_JWT_SECRET",
		},
		{
			name:    "zero access token ttl",
			mutate:  func(c *Config) { c.Auth.AccessTokenTTL = 0 },
			wantErr: "AUTH_ACCESS_TOKEN_TTL",
		},
		{
			name:    "refresh ttl not greater than access ttl",
			mutate:  func(c *Config) { c.Auth.RefreshTokenTTL = c.Auth.AccessTokenTTL },
			wantErr: "AUTH_REFRESH_TOKEN_TTL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid()
			tt.mutate(&cfg)

			err := cfg.Validate()

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() error = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
