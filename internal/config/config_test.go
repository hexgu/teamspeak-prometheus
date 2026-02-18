package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := []byte(`
servers:
  - name: "Test Server"
    host: "127.0.0.1"
    port: 10011
    username: "admin"
    password: "password"
metrics_port: 9090
read_interval: 30
`)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(config.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(config.Servers))
	}
	if config.Servers[0].Name != "Test Server" {
		t.Errorf("expected server name 'Test Server', got '%s'", config.Servers[0].Name)
	}
	if config.MetricsPort != 9090 {
		t.Errorf("expected metrics port 9090, got %d", config.MetricsPort)
	}
	if config.ReadInterval != 30 {
		t.Errorf("expected read interval 30, got %d", config.ReadInterval)
	}
}
