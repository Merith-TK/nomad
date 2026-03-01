package main

import (
	"os"
	"path/filepath"
)

// Config holds application configuration.
type Config struct {
	ConfigPath string
}

// getConfigPath determines the configuration path.
// In development, it uses a relative path; in production, it would use user home.
func getConfigPath() string {
	// Check for .nomad directory in current path
	if info, err := os.Stat(".nomad"); err == nil && info.IsDir() {
		path := filepath.Join(".nomad", "interface", "streamdeck", "config")
		if err := os.MkdirAll(path, 0755); err != nil {
			// Log error or handle appropriately
			return path // Still return path even if creation fails
		}
		return path
	}

	// Fall back to ~/.nomad
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Consider returning error or using a temp directory
		return filepath.Join(".nomad", "interface", "streamdeck", "config")
	}

	path := filepath.Join(homeDir, ".nomad", "interface", "streamdeck", "config")
	if err := os.MkdirAll(path, 0755); err != nil {
		// Log error or handle appropriately
	}
	return path
}

// ensureConfigDir creates the configuration directory if it doesn't exist.
func ensureConfigDir(configPath string) (string, error) {
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return "", err
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", err
	}

	return absConfigPath, nil
}
