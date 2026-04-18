package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/pkg/config"
)

func TestLoad_defaults(t *testing.T) {
	// Non-existent file should return defaults without error.
	cfg, err := config.Load(filepath.Join(t.TempDir(), "cli.yml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host() != "https://circleci.com" {
		t.Errorf("Host() = %q; want default", cfg.Host())
	}
	if cfg.Token() != "" {
		t.Errorf("Token() = %q; want empty", cfg.Token())
	}
	uc, _ := cfg.Get("update_check")
	if uc != "true" {
		t.Errorf("update_check = %q; want true", uc)
	}
}

func TestLoad_roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.yml")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if err := cfg.Set("host", "https://example.com"); err != nil {
		t.Fatalf("Set host: %v", err)
	}
	if err := cfg.Set("token", "tok-abc123"); err != nil {
		t.Fatalf("Set token: %v", err)
	}
	if err := cfg.Set("telemetry", "false"); err != nil {
		t.Fatalf("Set telemetry: %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Reload and verify.
	cfg2, err := config.Load(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if cfg2.Host() != "https://example.com" {
		t.Errorf("Host after reload = %q; want https://example.com", cfg2.Host())
	}
	if cfg2.Token() != "tok-abc123" {
		t.Errorf("Token after reload = %q; want tok-abc123", cfg2.Token())
	}
	tel, _ := cfg2.Get("telemetry")
	if tel != "false" {
		t.Errorf("telemetry after reload = %q; want false", tel)
	}
}

func TestLoad_envVarPrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.yml")
	cfg, _ := config.Load(path)
	_ = cfg.Set("token", "file-token")
	_ = cfg.Save()

	// Env var should win.
	t.Setenv("CIRCLECI_TOKEN", "env-token")
	cfg2, _ := config.Load(path)
	if cfg2.Token() != "env-token" {
		t.Errorf("Token() = %q; want env-token (env var precedence)", cfg2.Token())
	}
}

func TestSet_unknownKey(t *testing.T) {
	cfg, _ := config.Load(filepath.Join(t.TempDir(), "cli.yml"))
	if err := cfg.Set("nonexistent", "value"); err == nil {
		t.Error("Set with unknown key should return error")
	}
}

func TestSet_invalidBool(t *testing.T) {
	cfg, _ := config.Load(filepath.Join(t.TempDir(), "cli.yml"))
	if err := cfg.Set("telemetry", "maybe"); err == nil {
		t.Error("Set telemetry with non-bool should return error")
	}
}

func TestSave_createsDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	path := filepath.Join(dir, "cli.yml")
	cfg, _ := config.Load(path)
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save in nested dir: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}

func TestKeys(t *testing.T) {
	cfg, _ := config.Load(filepath.Join(t.TempDir(), "cli.yml"))
	keys := cfg.Keys()
	want := []string{"host", "token", "update_check", "telemetry"}
	if len(keys) != len(want) {
		t.Fatalf("Keys() = %v; want %v", keys, want)
	}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("Keys()[%d] = %q; want %q", i, k, want[i])
		}
	}
}
