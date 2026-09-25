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
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_HOST", "127.0.0.1")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "1s")
	t.Setenv("HTTP_READ_TIMEOUT", "2s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "3s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "4s")

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

func TestHTTPConfigAddr(t *testing.T) {
	cfg := HTTPConfig{Host: "0.0.0.0", Port: "8080"}

	if got, want := cfg.Addr(), "0.0.0.0:8080"; got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
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
