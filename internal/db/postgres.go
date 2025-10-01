package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// PostgresConnection represents a PostgreSQL database connection
type PostgresConnection struct {
	config ConnectionConfig
	conn   *pgx.Conn
	status ConnectionStatus
}

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection(config ConnectionConfig) *PostgresConnection {
	return &PostgresConnection{
		config: config,
		status: StatusDisconnected,
	}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresConnection) Connect(ctx context.Context) error {
	p.status = StatusConnecting

	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.Username,
		p.config.Password,
		p.config.Database,
		p.config.SSLMode,
	)

	// Attempt connection
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		p.status = StatusError
		return fmt.Errorf("failed to connect: %w", err)
	}

	p.conn = conn
	p.status = StatusConnected
	return nil
}

// Disconnect closes the database connection
func (p *PostgresConnection) Disconnect(ctx context.Context) error {
	if p.conn == nil {
		return nil
	}

	err := p.conn.Close(ctx)
	p.conn = nil
	p.status = StatusDisconnected
	return err
}

// Ping checks if the connection is alive
func (p *PostgresConnection) Ping(ctx context.Context) error {
	if p.conn == nil {
		return fmt.Errorf("not connected")
	}
	return p.conn.Ping(ctx)
}

// Status returns the current connection status
func (p *PostgresConnection) Status() ConnectionStatus {
	return p.status
}

// Config returns the connection configuration
func (p *PostgresConnection) Config() ConnectionConfig {
	return p.config
}

// Conn returns the underlying pgx connection
func (p *PostgresConnection) Conn() *pgx.Conn {
	return p.conn
}
