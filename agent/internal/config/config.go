package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the agent's persistent configuration.
type Config struct {
	AgentID         string `json:"agentId"`
	Port            int    `json:"port"`
	SigningSecret   string `json:"signingSecret"`
	PairingToken    string `json:"pairingToken,omitempty"`
	PairedTokenHash string `json:"pairedTokenHash,omitempty"`

	path string `json:"-"`
}

// Load reads a config from the given JSON file path.
// If the file doesn't exist, it returns a new config with generated defaults.
func Load(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}

	cfg := &Config{path: absPath}

	data, err := os.ReadFile(absPath)
	if os.IsNotExist(err) {
		if err := cfg.generateDefaults(); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.path = absPath

	if cfg.AgentID == "" {
		cfg.AgentID, err = generateHex(16)
		if err != nil {
			return nil, fmt.Errorf("generate agent ID: %w", err)
		}
	}
	if cfg.SigningSecret == "" {
		cfg.SigningSecret, err = generateHex(32)
		if err != nil {
			return nil, fmt.Errorf("generate signing secret: %w", err)
		}
	}

	return cfg, nil
}

// Save writes the config back to its file path.
func (c *Config) Save() error {
	if c.path == "" {
		return fmt.Errorf("config has no file path")
	}

	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// IsPaired returns true if a desktop has been paired.
func (c *Config) IsPaired() bool {
	return c.PairedTokenHash != ""
}

// SetPath sets the file path for Save operations.
func (c *Config) SetPath(path string) {
	c.path = path
}

// GeneratePairingToken creates a new one-time pairing token.
func (c *Config) GeneratePairingToken() error {
	token, err := generateHex(12)
	if err != nil {
		return err
	}
	c.PairingToken = token
	return nil
}

func (c *Config) generateDefaults() error {
	var err error
	c.AgentID, err = generateHex(16)
	if err != nil {
		return fmt.Errorf("generate agent ID: %w", err)
	}
	c.SigningSecret, err = generateHex(32)
	if err != nil {
		return fmt.Errorf("generate signing secret: %w", err)
	}
	return nil
}

func generateHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
