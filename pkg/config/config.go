package config

import (
	"errors"
	"os"
	"path/filepath"
)

// Config represents the minimal runtime configuration for the application.
// Only external secrets and PVC mount paths are loaded from environment variables.
// All internal tuning values (port 8080, timeouts, limits) are immutable constants.
type Config struct {
	DataDir            string
	DBPath             string
	AssetDir           string
	SessionSecret      string
	GoogleClientID     string
	GoogleClientSecret string
}

// Immutable internal constants
const (
	DefaultPort        = 8080
	DefaultDBTimeoutMs = 5000
	MaxUploadBytes     = 10 * 1024 * 1024 // 10MB
)

// Load reads configuration from environment variables with zero-config defaults.
func Load() (*Config, error) {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "dev-insecure-secret-key-32bytes-min!!"
	}

	cfg := &Config{
		DataDir:            dataDir,
		DBPath:             filepath.Join(dataDir, "penlight.db"),
		AssetDir:           filepath.Join(dataDir, "assets"),
		SessionSecret:      sessionSecret,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}

	// Fail-fast validation when running in production (indicated by PVC mount path /data)
	if cfg.DataDir == "/data" {
		if cfg.SessionSecret == "dev-insecure-secret-key-32bytes-min!!" || len(cfg.SessionSecret) < 32 {
			return nil, errors.New("SESSION_SECRET must be set to a secure key (>= 32 bytes) in production")
		}
	}

	return cfg, nil
}
