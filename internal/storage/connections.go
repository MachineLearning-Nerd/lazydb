package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// ConnectionsConfig stores all connection configurations and state
type ConnectionsConfig struct {
	Connections      []db.ConnectionConfig `json:"connections"`
	ActiveConnection string                `json:"active_connection"`
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

// SaveConnections saves all connections to file with encrypted passwords
func SaveConnections(connections []db.ConnectionConfig, activeConnection string) error {
	filePath, err := GetConnectionsFile()
	if err != nil {
		return err
	}

	// Create a copy of connections with encrypted passwords
	encryptedConns := make([]db.ConnectionConfig, len(connections))
	for i, conn := range connections {
		encryptedConns[i] = conn
		// Encrypt password if not empty
		if conn.Password != "" {
			encrypted, err := Encrypt(conn.Password)
			if err != nil {
				return err
			}
			encryptedConns[i].Password = encrypted
		}
	}

	config := ConnectionsConfig{
		Connections:      encryptedConns,
		ActiveConnection: activeConnection,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600) // 0600 for security (user read/write only)
}

// LoadConnections loads all connections from file and decrypts passwords
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

	// Decrypt passwords
	for i := range config.Connections {
		if config.Connections[i].Password != "" {
			decrypted, err := Decrypt(config.Connections[i].Password)
			if err != nil {
				// If decryption fails, it might be a plaintext password from old version
				// Keep the original value (backward compatibility)
				continue
			}
			config.Connections[i].Password = decrypted
		}
	}

	return &config, nil
}

// MarshalConfig marshals config to JSON (exported for testing)
func MarshalConfig(config *ConnectionsConfig) ([]byte, error) {
	return json.MarshalIndent(config, "", "  ")
}

// UnmarshalConfig unmarshals JSON to config and decrypts passwords (exported for testing)
func UnmarshalConfig(data []byte) (*ConnectionsConfig, error) {
	var config ConnectionsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Decrypt passwords
	for i := range config.Connections {
		if config.Connections[i].Password != "" {
			decrypted, err := Decrypt(config.Connections[i].Password)
			if err != nil {
				// Backward compatibility: keep original if decryption fails
				continue
			}
			config.Connections[i].Password = decrypted
		}
	}

	return &config, nil
}

