// Package config loads service configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds the service configuration.
type Config struct {
	APIKey       string        `env:"KIMI_API_KEY,required,notEmpty"`
	BaseURL      string        `env:"KIMI_BASE_URL" envDefault:"https://api.kimi.com/coding/v1"`
	PollInterval time.Duration `env:"POLL_INTERVAL" envDefault:"5m"`

	MQTTBroker   string `env:"MQTT_BROKER,required,notEmpty"`
	MQTTClientID string `env:"MQTT_CLIENT_ID" envDefault:"kimi-quota-monitor"`
	MQTTTopic    string `env:"MQTT_TOPIC" envDefault:"quota/llm"`
	MQTTUsername string `env:"MQTT_USERNAME"`
	MQTTPassword string `env:"MQTT_PASSWORD"`

	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

// Load parses environment variables into a Config.
func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if cfg.PollInterval <= 0 {
		return Config{}, errors.New("POLL_INTERVAL must be positive")
	}
	return cfg, nil
}
