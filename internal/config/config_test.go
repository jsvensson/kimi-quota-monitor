package config

import "testing"

// TestLoad verifies that Load reads required variables and applies defaults.
func TestLoad(t *testing.T) {
	t.Setenv("KIMI_API_KEY", "sk-kimi-test")
	t.Setenv("MQTT_BROKER", "tcp://localhost:1883")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.APIKey != "sk-kimi-test" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "sk-kimi-test")
	}
	if cfg.MQTTBroker != "tcp://localhost:1883" {
		t.Errorf("MQTTBroker = %q, want %q", cfg.MQTTBroker, "tcp://localhost:1883")
	}
	if cfg.MQTTTopic != "quota/llm" {
		t.Errorf("MQTTTopic = %q, want default %q", cfg.MQTTTopic, "quota/llm")
	}
	if cfg.PollInterval.String() != "5m0s" {
		t.Errorf("PollInterval = %v, want 5m", cfg.PollInterval)
	}
}

// TestLoadMissingRequired verifies that Load fails when a required
// variable is missing.
func TestLoadMissingRequired(t *testing.T) {
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MQTT_BROKER", "tcp://localhost:1883")

	if _, err := Load(); err == nil {
		t.Error("Load() error = nil, want error for missing KIMI_API_KEY")
	}
}

// TestLoadInvalidInterval verifies that Load rejects a non-positive
// POLL_INTERVAL.
func TestLoadInvalidInterval(t *testing.T) {
	t.Setenv("KIMI_API_KEY", "sk-kimi-test")
	t.Setenv("MQTT_BROKER", "tcp://localhost:1883")
	t.Setenv("POLL_INTERVAL", "0s")

	if _, err := Load(); err == nil {
		t.Error("Load() error = nil, want error for zero POLL_INTERVAL")
	}
}
