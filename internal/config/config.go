// Package config handles loading and validating portwatch configuration.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultScanInterval is the fallback scan interval if none is specified.
const DefaultScanInterval = 30 * time.Second

// Config holds the top-level portwatch configuration.
type Config struct {
	// ScanInterval defines how often the port scanner runs.
	ScanInterval time.Duration `yaml:"scan_interval"`

	// Protocols lists which protocols to monitor ("tcp", "udp").
	Protocols []string `yaml:"protocols"`

	// AllowedPorts defines ports that are expected to be open.
	// Changes outside this list will trigger alerts.
	AllowedPorts []uint16 `yaml:"allowed_ports"`

	// IgnorePorts lists ports to silently ignore regardless of state.
	IgnorePorts []uint16 `yaml:"ignore_ports"`

	// Alerts configures the alerting backends.
	Alerts AlertConfig `yaml:"alerts"`
}

// AlertConfig controls how and where alerts are delivered.
type AlertConfig struct {
	// Console enables printing alerts to stdout.
	Console bool `yaml:"console"`

	// WebhookURL, if set, sends alert payloads via HTTP POST.
	WebhookURL string `yaml:"webhook_url"`

	// WebhookTimeout is the HTTP client timeout for webhook calls.
	WebhookTimeout time.Duration `yaml:"webhook_timeout"`
}

// Load reads a YAML config file from the given path and returns a Config.
// If path is empty, a default Config is returned.
func Load(path string) (*Config, error) {
	cfg := defaults()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: reading file %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parsing YAML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that the configuration values are semantically valid.
func (c *Config) Validate() error {
	if c.ScanInterval < time.Second {
		return fmt.Errorf("config: scan_interval must be at least 1s, got %s", c.ScanInterval)
	}

	validProto := map[string]bool{"tcp": true, "udp": true}
	for _, p := range c.Protocols {
		if !validProto[p] {
			return fmt.Errorf("config: unsupported protocol %q (must be \"tcp\" or \"udp\")", p)
		}
	}

	if len(c.Protocols) == 0 {
		return fmt.Errorf("config: at least one protocol must be specified")
	}

	if c.Alerts.WebhookURL != "" && c.Alerts.WebhookTimeout <= 0 {
		c.Alerts.WebhookTimeout = 5 * time.Second
	}

	return nil
}

// AllowedPortSet returns a set (map) of allowed ports for O(1) lookup.
func (c *Config) AllowedPortSet() map[uint16]struct{} {
	s := make(map[uint16]struct{}, len(c.AllowedPorts))
	for _, p := range c.AllowedPorts {
		s[p] = struct{}{}
	}
	return s
}

// IgnorePortSet returns a set (map) of ignored ports for O(1) lookup.
func (c *Config) IgnorePortSet() map[uint16]struct{} {
	s := make(map[uint16]struct{}, len(c.IgnorePorts))
	for _, p := range c.IgnorePorts {
		s[p] = struct{}{}
	}
	return s
}

// defaults returns a Config populated with sensible defaults.
func defaults() *Config {
	return &Config{
		ScanInterval: DefaultScanInterval,
		Protocols:    []string{"tcp"},
		Alerts: AlertConfig{
			Console:        true,
			WebhookTimeout: 5 * time.Second,
		},
	}
}
