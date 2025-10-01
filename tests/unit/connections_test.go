package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	"github.com/MachineLearning-Nerd/lazydb/internal/storage"
)

// TestHelper allows us to override GetConnectionsFile for testing
type TestConnectionsHelper struct {
	testFile string
}

func (h *TestConnectionsHelper) Setup(t *testing.T) {
	tempDir := t.TempDir()
	h.testFile = filepath.Join(tempDir, "test_connections.json")
}

func TestSaveLoadConnections(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_connections.json")

	// Test data
	connections := []db.ConnectionConfig{
		{
			Name:        "test-dev",
			Host:        "localhost",
			Port:        5432,
			Database:    "testdb",
			Username:    "testuser",
			Password:    "secret-password-123",
			SSLMode:     "disable",
			Environment: db.EnvDevelopment,
		},
		{
			Name:        "test-prod",
			Host:        "prod.example.com",
			Port:        5432,
			Database:    "proddb",
			Username:    "produser",
			Password:    "prod-secret-456",
			SSLMode:     "require",
			Environment: db.EnvProduction,
		},
	}
	activeConnection := "test-dev"

	// Encrypt passwords manually
	encryptedConns := make([]db.ConnectionConfig, len(connections))
	for i, conn := range connections {
		encryptedConns[i] = conn
		if conn.Password != "" {
			encrypted, err := storage.Encrypt(conn.Password)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}
			encryptedConns[i].Password = encrypted
		}
	}

	// Save to file
	err := saveConnectionsToFile(testFile, encryptedConns, activeConnection)
	if err != nil {
		t.Fatalf("saveConnectionsToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Connections file was not created")
	}

	// Read file and verify passwords are encrypted
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	fileContent := string(data)
	if strings.Contains(fileContent, "secret-password-123") || strings.Contains(fileContent, "prod-secret-456") {
		t.Error("Passwords are not encrypted in the file")
	}

	// Load connections
	loadedConfig, err := loadConnectionsFromFile(testFile)
	if err != nil {
		t.Fatalf("loadConnectionsFromFile failed: %v", err)
	}

	// Verify active connection
	if loadedConfig.ActiveConnection != activeConnection {
		t.Errorf("Active connection mismatch. Got %q, want %q", loadedConfig.ActiveConnection, activeConnection)
	}

	// Verify connection count
	if len(loadedConfig.Connections) != len(connections) {
		t.Errorf("Connection count mismatch. Got %d, want %d", len(loadedConfig.Connections), len(connections))
	}

	// Verify each connection (passwords should be decrypted)
	for i, conn := range loadedConfig.Connections {
		if conn.Name != connections[i].Name {
			t.Errorf("Connection %d: Name mismatch. Got %q, want %q", i, conn.Name, connections[i].Name)
		}
		if conn.Password != connections[i].Password {
			t.Errorf("Connection %d: Password decryption failed. Got %q, want %q", i, conn.Password, connections[i].Password)
		}
		if conn.Environment != connections[i].Environment {
			t.Errorf("Connection %d: Environment mismatch. Got %q, want %q", i, conn.Environment, connections[i].Environment)
		}
	}
}

func TestPasswordEncryptionRoundTrip(t *testing.T) {
	testPasswords := []string{
		"simple",
		"",
		"with-special-chars!@#$%",
		"unicode-密码",
	}

	for _, password := range testPasswords {
		t.Run(password, func(t *testing.T) {
			// Encrypt
			encrypted, err := storage.Encrypt(password)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Decrypt
			decrypted, err := storage.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Verify
			if decrypted != password {
				t.Errorf("Password mismatch. Got %q, want %q", decrypted, password)
			}
		})
	}
}

// Helper functions to test storage logic without relying on global state

func saveConnectionsToFile(filePath string, connections []db.ConnectionConfig, activeConnection string) error {
	// Similar to storage.SaveConnections but with explicit file path
	config := storage.ConnectionsConfig{
		Connections:      connections,
		ActiveConnection: activeConnection,
	}

	data, err := storage.MarshalConfig(&config)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600)
}

func loadConnectionsFromFile(filePath string) (*storage.ConnectionsConfig, error) {
	// Similar to storage.LoadConnections but with explicit file path
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &storage.ConnectionsConfig{
			Connections:      []db.ConnectionConfig{},
			ActiveConnection: "",
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return storage.UnmarshalConfig(data)
}
