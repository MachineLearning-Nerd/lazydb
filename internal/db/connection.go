package db

import (
	"context"
	"fmt"
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

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Name     string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

// Connection represents a database connection
type Connection interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	Status() ConnectionStatus
	Config() ConnectionConfig
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

// ListConnections returns all connection names
func (cm *ConnectionManager) ListConnections() []string {
	names := make([]string, 0, len(cm.connections))
	for name := range cm.connections {
		names = append(names, name)
	}
	return names
}

// ActiveName returns the name of the active connection
func (cm *ConnectionManager) ActiveName() string {
	return cm.active
}
