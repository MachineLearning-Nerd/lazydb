package db

import (
	"context"
	"fmt"
	"sort"
)

// ConnectionStatus represents the state of a database connection
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

func (s ConnectionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "Disconnected"
	case StatusConnecting:
		return "Connecting..."
	case StatusConnected:
		return "Connected"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Environment represents the deployment environment
type Environment string

const (
	EnvDevelopment Environment = "Development"
	EnvStaging     Environment = "Staging"
	EnvProduction  Environment = "Production"
)

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Name        string
	Host        string
	Port        int
	Database    string
	Username    string
	Password    string
	SSLMode     string
	Environment Environment
}

// SchemaObject represents a database schema object
type SchemaObject struct {
	Name   string
	Type   string // "table", "view", "function", "sequence"
	Schema string
}

// TableColumn represents a table column
type TableColumn struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
}

// Connection represents a database connection
type Connection interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	Status() ConnectionStatus
	Config() ConnectionConfig

	// Schema exploration methods
	ListSchemas(ctx context.Context) ([]string, error)
	ListTables(ctx context.Context, schema string) ([]SchemaObject, error)
	ListViews(ctx context.Context, schema string) ([]SchemaObject, error)
	ListFunctions(ctx context.Context, schema string) ([]SchemaObject, error)
	GetTableColumns(ctx context.Context, schema, table string) ([]TableColumn, error)
}

// ConnectionManager manages multiple database connections
type ConnectionManager struct {
	connections map[string]Connection
	active      string
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]Connection),
	}
}

// AddConnection adds a connection to the manager
func (cm *ConnectionManager) AddConnection(name string, conn Connection) {
	cm.connections[name] = conn
}

// GetConnection retrieves a connection by name
func (cm *ConnectionManager) GetConnection(name string) (Connection, error) {
	conn, exists := cm.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection %s not found", name)
	}
	return conn, nil
}

// SetActive sets the active connection
func (cm *ConnectionManager) SetActive(name string) error {
	if _, exists := cm.connections[name]; !exists {
		return fmt.Errorf("connection %s not found", name)
	}
	cm.active = name
	return nil
}

// GetActive returns the active connection
func (cm *ConnectionManager) GetActive() (Connection, error) {
	if cm.active == "" {
		return nil, fmt.Errorf("no active connection")
	}
	return cm.GetConnection(cm.active)
}

// ListConnections returns all connection names in sorted order
func (cm *ConnectionManager) ListConnections() []string {
	names := make([]string, 0, len(cm.connections))
	for name := range cm.connections {
		names = append(names, name)
	}
	// Sort for stable, predictable order
	sort.Strings(names)
	return names
}

// ActiveName returns the name of the active connection
func (cm *ConnectionManager) ActiveName() string {
	return cm.active
}

// RemoveConnection removes a connection from the manager
func (cm *ConnectionManager) RemoveConnection(name string) error {
	if _, exists := cm.connections[name]; !exists {
		return fmt.Errorf("connection %s not found", name)
	}

	delete(cm.connections, name)

	// Clear active if we're deleting the active connection
	if cm.active == name {
		cm.active = ""
	}

	return nil
}

// GetAllConfigs returns all connection configurations (for persistence)
func (cm *ConnectionManager) GetAllConfigs() []ConnectionConfig {
	configs := make([]ConnectionConfig, 0, len(cm.connections))
	for _, conn := range cm.connections {
		configs = append(configs, conn.Config())
	}
	return configs
}
