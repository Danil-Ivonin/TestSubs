package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUsesDefaultsWhenFilesAndEnvAreMissing(t *testing.T) {
	t.Setenv("HTTP_PORT", "")
	t.Setenv("POSTGRES_HOST", "")
	t.Setenv("POSTGRES_PORT", "")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("POSTGRES_DB", "")
	t.Setenv("POSTGRES_SSLMODE", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load(Options{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Port != "8080" {
		t.Fatalf("HTTP.Port = %q, want %q", cfg.HTTP.Port, "8080")
	}
	if cfg.HTTP.ReadTimeout != 5*time.Second {
		t.Fatalf("HTTP.ReadTimeout = %s, want %s", cfg.HTTP.ReadTimeout, 5*time.Second)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "info")
	}
	if cfg.Postgres.DSN != "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" {
		t.Fatalf("Postgres.DSN = %q", cfg.Postgres.DSN)
	}
}

func TestLoadReadsYAMLConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(configPath, []byte(`
http:
  port: "9090"
  read_timeout: 3s
  write_timeout: 4s
  idle_timeout: 15s
postgres:
  host: db
  port: "5433"
  user: user
  password: pass
  db: subs
  sslmode: require
log:
  level: debug
  format: json
`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(Options{ConfigPath: configPath})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Port != "9090" {
		t.Fatalf("HTTP.Port = %q, want %q", cfg.HTTP.Port, "9090")
	}
	if cfg.Postgres.DSN != "postgres://user:pass@db:5433/subs?sslmode=require" {
		t.Fatalf("Postgres.DSN = %q", cfg.Postgres.DSN)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "debug")
	}
	if cfg.Log.Format != "json" {
		t.Fatalf("Log.Format = %q, want %q", cfg.Log.Format, "json")
	}
}

func TestLoadEnvOverridesYAMLConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(configPath, []byte(`
http:
  port: "9090"
postgres:
  host: config-host
  port: "5432"
  user: config-user
  password: config-password
  db: config-db
  sslmode: disable
log:
  level: info
`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("HTTP_PORT", "7070")
	t.Setenv("POSTGRES_HOST", "env-host")
	t.Setenv("POSTGRES_PORT", "15432")
	t.Setenv("POSTGRES_USER", "env-user")
	t.Setenv("POSTGRES_PASSWORD", "env-password")
	t.Setenv("POSTGRES_DB", "env-db")
	t.Setenv("POSTGRES_SSLMODE", "verify-full")
	t.Setenv("LOG_LEVEL", "warn")

	cfg, err := Load(Options{ConfigPath: configPath})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Port != "7070" {
		t.Fatalf("HTTP.Port = %q, want %q", cfg.HTTP.Port, "7070")
	}
	if cfg.Postgres.DSN != "postgres://env-user:env-password@env-host:15432/env-db?sslmode=verify-full" {
		t.Fatalf("Postgres.DSN = %q", cfg.Postgres.DSN)
	}
	if cfg.Log.Level != "warn" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "warn")
	}
}
