package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConnection represents a PostgreSQL database connection pool
type PostgresConnection struct {
	config ConnectionConfig
	pool   *pgxpool.Pool
	status ConnectionStatus
}

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection(config ConnectionConfig) *PostgresConnection {
	return &PostgresConnection{
		config: config,
		status: StatusDisconnected,
	}
}

// Connect establishes a connection pool to the PostgreSQL database
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

	// Create pool configuration
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		p.status = StatusError
		return fmt.Errorf("failed to parse connection config: %w", err)
	}

	// Configure pool settings for optimal performance
	poolConfig.MaxConns = 10                    // Maximum connections in pool
	poolConfig.MinConns = 2                     // Minimum connections to maintain
	poolConfig.HealthCheckPeriod = 30 * time.Second
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Set connection timeout
	timeout := time.Duration(p.config.ConnectionTimeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	poolConfig.ConnConfig.ConnectTimeout = timeout

	// Create context with timeout for initial connection
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Attempt to create connection pool
	pool, err := pgxpool.NewWithConfig(ctxWithTimeout, poolConfig)
	if err != nil {
		p.status = StatusError

		// Check if error is due to timeout
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("connection timeout after %v: could not connect to database at %s:%d", timeout, p.config.Host, p.config.Port)
		}

		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection with a ping
	err = pool.Ping(ctxWithTimeout)
	if err != nil {
		pool.Close()
		p.status = StatusError
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.pool = pool
	p.status = StatusConnected
	return nil
}

// Disconnect closes the database connection pool
func (p *PostgresConnection) Disconnect(ctx context.Context) error {
	if p.pool == nil {
		return nil
	}

	p.pool.Close()
	p.pool = nil
	p.status = StatusDisconnected
	return nil
}

// Ping checks if the connection is alive
func (p *PostgresConnection) Ping(ctx context.Context) error {
	if p.pool == nil {
		return fmt.Errorf("not connected")
	}
	return p.pool.Ping(ctx)
}

// Status returns the current connection status
func (p *PostgresConnection) Status() ConnectionStatus {
	return p.status
}

// Config returns the connection configuration
func (p *PostgresConnection) Config() ConnectionConfig {
	return p.config
}

// Pool returns the underlying connection pool
func (p *PostgresConnection) Pool() *pgxpool.Pool {
	return p.pool
}

// Conn returns the underlying pgx connection for backward compatibility
// Note: This method is deprecated, use Pool() instead
func (p *PostgresConnection) Conn() interface{} {
	// For backward compatibility, return the pool
	// This method should not be used with connection pools
	return p.pool
}

// ListSchemas returns all schemas in the database
func (p *PostgresConnection) ListSchemas(ctx context.Context) ([]string, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT LIKE 'pg_%'
		  AND schema_name != 'information_schema'
		ORDER BY schema_name
	`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	return schemas, rows.Err()
}

// ListTables returns all tables in a schema
func (p *PostgresConnection) ListTables(ctx context.Context, schema string) ([]SchemaObject, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		  AND table_type = 'BASE TABLE'
		  AND table_name NOT LIKE 'pg_%'
		ORDER BY table_name
	`

	rows, err := p.pool.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []SchemaObject
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, SchemaObject{
			Name:   name,
			Type:   "table",
			Schema: schema,
		})
	}

	return tables, rows.Err()
}

// ListViews returns all views in a schema
func (p *PostgresConnection) ListViews(ctx context.Context, schema string) ([]SchemaObject, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT table_name
		FROM information_schema.views
		WHERE table_schema = $1
		ORDER BY table_name
	`

	rows, err := p.pool.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list views: %w", err)
	}
	defer rows.Close()

	var views []SchemaObject
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		views = append(views, SchemaObject{
			Name:   name,
			Type:   "view",
			Schema: schema,
		})
	}

	return views, rows.Err()
}

// ListFunctions returns all functions in a schema
func (p *PostgresConnection) ListFunctions(ctx context.Context, schema string) ([]SchemaObject, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT routine_name
		FROM information_schema.routines
		WHERE routine_schema = $1 AND routine_type = 'FUNCTION'
		ORDER BY routine_name
	`

	rows, err := p.pool.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}
	defer rows.Close()

	var functions []SchemaObject
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan function: %w", err)
		}
		functions = append(functions, SchemaObject{
			Name:   name,
			Type:   "function",
			Schema: schema,
		})
	}

	return functions, rows.Err()
}

// GetTableColumns returns column information for a table
func (p *PostgresConnection) GetTableColumns(ctx context.Context, schema, table string) ([]TableColumn, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT
			column_name,
			data_type,
			is_nullable,
			COALESCE(column_default, '')
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := p.pool.Query(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table columns: %w", err)
	}
	defer rows.Close()

	var columns []TableColumn
	for rows.Next() {
		var col TableColumn
		var nullable string
		if err := rows.Scan(&col.Name, &col.Type, &nullable, &col.Default); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		col.Nullable = (nullable == "YES")
		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// ExecuteQuery executes a SQL query using the connection pool
func (p *PostgresConnection) ExecuteQuery(ctx context.Context, query string) (QueryResult, error) {
	if p.pool == nil {
		return QueryResult{}, fmt.Errorf("not connected to database")
	}

	result := ExecuteQuery(ctx, p.pool, query)
	if result.Error != nil {
		return result, result.Error
	}

	return result, nil
}

