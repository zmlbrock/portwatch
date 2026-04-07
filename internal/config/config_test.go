package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	if cfg.Interval != 30*time.Second {
		t.Errorf("expected default interval 30s, got %v", cfg.Interval)
	}
	if cfg.Protocols == nil || len(cfg.Protocols) == 0 {
		t.Error("expected default protocols to be non-empty")
	}
	if cfg.AlertThreshold < 1 {
		t.Errorf("expected alert threshold >= 1, got %d", cfg.AlertThreshold)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/portwatch.yaml")
	if err == nil {
		t.Error("expected error for missing config file, got nil")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
interval: 10s
protocols:
  - tcp
  - udp
alert_threshold: 2
ignored_ports:
  - 8080
  - 9090
notifiers:
  console:
    enabled: true
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "portwatch.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Interval != 10*time.Second {
		t.Errorf("expected interval 10s, got %v", cfg.Interval)
	}
	if len(cfg.Protocols) != 2 {
		t.Errorf("expected 2 protocols, got %d", len(cfg.Protocols))
	}
	if cfg.AlertThreshold != 2 {
		t.Errorf("expected alert threshold 2, got %d", cfg.AlertThreshold)
	}
	if len(cfg.IgnoredPorts) != 2 {
		t.Errorf("expected 2 ignored ports, got %d", len(cfg.IgnoredPorts))
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := `interval: [invalid yaml`

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "portwatch.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	// Empty config should fall back to defaults
	content := `{}`

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "portwatch.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def := defaults()
	if cfg.Interval != def.Interval {
		t.Errorf("expected default interval %v, got %v", def.Interval, cfg.Interval)
	}
	if len(cfg.Protocols) != len(def.Protocols) {
		t.Errorf("expected default protocols length %d, got %d", len(def.Protocols), len(cfg.Protocols))
	}
}
