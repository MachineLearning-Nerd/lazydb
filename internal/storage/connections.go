package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// ConnectionsConfig stores all connection configurations and state
type ConnectionsConfig struct {
	Connections    []db.ConnectionConfig `json:"connections"`
	ActiveConnection string              `json:"active_connection"`
}

// GetConnectionsDir returns the directory for LazyDB configuration
func GetConnectionsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".lazydb")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// GetConnectionsFile returns the full path to connections.json
func GetConnectionsFile() (string, error) {
	configDir, err := GetConnectionsDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "connections.json"), nil
}

// SaveConnections saves all connections to file
func SaveConnections(connections []db.ConnectionConfig, activeConnection string) error {
	filePath, err := GetConnectionsFile()
	if err != nil {
		return err
	}

	config := ConnectionsConfig{
		Connections:      connections,
		ActiveConnection: activeConnection,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600) // 0600 for security (user read/write only)
}

// LoadConnections loads all connections from file
func LoadConnections() (*ConnectionsConfig, error) {
	filePath, err := GetConnectionsFile()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &ConnectionsConfig{
			Connections:      []db.ConnectionConfig{},
			ActiveConnection: "",
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config ConnectionsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
