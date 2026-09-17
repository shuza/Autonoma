package config

import "testing"

func TestLoadFromEnvUsesDefaults(t *testing.T) {
	t.Setenv("AUTONOMA_API_HOST", "")
	t.Setenv("AUTONOMA_API_PORT", "")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Errorf("Error loading config from env: %v", err)
	}

	if cfg.Host != defaultHost {
		t.Errorf("Expected default host %q, got %q", defaultHost, cfg.Host)
	}

	if cfg.Port != defaultPort {
		t.Errorf("Expected default port %q, got %q", defaultPort, cfg.Port)
	}
}

func TestLoadFromEnvRejectsZeroPort(t *testing.T) {
	t.Setenv("AUTONOMA_API_PORT", "0")
	_, err := LoadFromEnv()

	if err == nil {
		t.Fatalf("expected error for zero port, got nil")
	}
}
