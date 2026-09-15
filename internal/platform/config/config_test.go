package config

import "testing"

func TestLoadFromEnvUseDefaults(t *testing.T) {
	t.Setenv("AUTONOMA_API_HOST", "")
	t.Setenv("AUTONOMA_API_PORT", "")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Errorf("Error loading config from env: %v", err)
	}
	if cfg.Host != defaultHost {
		t.Errorf("Expected default host %q, got %q", defaultHost, cfg.Host)
	}
	if cfg.Port != defaultHost {
		t.Errorf("Expected default port %q, got %q", defaultPort, cfg.Port)
	}
}
