# 🗄️ Database Abstraction Layer - LazyDB

Complete specification for the database abstraction layer supporting PostgreSQL, MySQL, and SQLite.

---

## Overview

LazyDB's database layer provides a **unified interface** for interacting with multiple database types through a clean abstraction that hides driver-specific details while exposing common functionality.

### Design Goals

1. **Unified API**: Single interface for all database operations
2. **Type Safety**: Strong typing for connections and results
3. **Error Handling**: Consistent error types across drivers
4. **Performance**: Connection pooling and query optimization
5. **Extensibility**: Easy to add new database types
6. **Testability**: Mockable interfaces for testing

### Supported Databases

| Database | Driver | Min Version | Status |
|----------|--------|-------------|--------|
| PostgreSQL | pgx/v5 | 12.0+ | ✅ MVP |
| MySQL | go-sql-driver | 5.7+ | ✅ Phase 2 |
| SQLite | mattn/go-sqlite3 | 3.35+ | ✅ Phase 2 |
| MariaDB | go-sql-driver | 10.5+ | 🔄 Phase 2 |
| CockroachDB | pgx/v5 | 22.1+ | 📋 Future |
| MongoDB | mongo-driver | 5.0+ | 📋 Future |

---

## Architecture

### Layer Diagram

```
┌─────────────────────────────────────────────────────────┐
│                  LazyDB Application                      │
├─────────────────────────────────────────────────────────┤
│                    Database Interface                    │
│  (Connect, Execute, ListTables, Disconnect, etc.)       │
├──────────────┬──────────────┬──────────────┬────────────┤
│  PostgresDB  │   MySQLDB    │   SQLiteDB   │  Future... │
├──────────────┼──────────────┼──────────────┼────────────┤
│   pgx/v5     │ go-sql-driver│ go-sqlite3   │    ...     │
└──────────────┴──────────────┴──────────────┴────────────┘
```

### Package Structure

```
internal/db/
├── database.go         # Database interface + common types
├── connection.go       # Connection pool management
├── postgres.go         # PostgreSQL implementation
├── mysql.go            # MySQL implementation
├── sqlite.go           # SQLite implementation
├── query.go            # Query execution helpers
├── schema.go           # Schema introspection
├── types.go            # Type mapping and conversions
├── errors.go           # Error types and handling
└── testing/
    ├── mock.go         # Mock database for testing
    └── fixtures.go     # Test data fixtures
```

---

## Core Interface

### Database Interface

```go
package db

import (
	"context"
	"time"
)

// Database is the main interface that all database drivers must implement
type Database interface {
	// Connection Management
	Connect(ctx context.Context, config ConnectionConfig) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Query Execution
	Execute(ctx context.Context, query string, params ...interface{}) (*ResultSet, error)
	ExecuteMulti(ctx context.Context, queries []string) ([]*ResultSet, error)

	// Schema Introspection
	ListTables(ctx context.Context, schema string) ([]Table, error)
	ListViews(ctx context.Context, schema string) ([]View, error)
	ListSchemas(ctx context.Context) ([]Schema, error)
	DescribeTable(ctx context.Context, schema, table string) (*TableSchema, error)
	GetIndexes(ctx context.Context, schema, table string) ([]Index, error)
	GetConstraints(ctx context.Context, schema, table string) ([]Constraint, error)

	// Transaction Support
	BeginTx(ctx context.Context) (Transaction, error)

	// Database Metadata
	GetVersion(ctx context.Context) (string, error)
	GetCurrentDatabase(ctx context.Context) (string, error)
	GetCurrentSchema(ctx context.Context) (string, error)
	GetCurrentUser(ctx context.Context) (string, error)

	// Performance & Stats
	GetStats(ctx context.Context) (*Stats, error)

	// Type Information
	GetDriverName() string
	GetDialect() Dialect
}

// Transaction represents a database transaction
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Execute(ctx context.Context, query string, params ...interface{}) (*ResultSet, error)
}

// Dialect represents database-specific SQL dialect
type Dialect interface {
	QuoteIdentifier(name string) string
	Placeholder(n int) string
	LimitClause(limit, offset int64) string
	CurrentTimestamp() string
	DataTypeMapping(nativeType string) DataType
}
```

---

## Data Types

### Connection Configuration

```go
// ConnectionConfig holds database connection parameters
type ConnectionConfig struct {
	// Common fields
	Name         string        `json:"name"`
	Type         DatabaseType  `json:"type"`
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Database     string        `json:"database"`
	Username     string        `json:"username"`
	Password     string        `json:"password"`

	// Connection options
	SSLMode      string        `json:"ssl_mode"`      // PostgreSQL: disable, require, verify-ca, verify-full
	Timeout      time.Duration `json:"timeout"`
	MaxRetries   int           `json:"max_retries"`

	// Pool settings
	MaxConns     int           `json:"max_conns"`
	MinConns     int           `json:"min_conns"`
	MaxIdleTime  time.Duration `json:"max_idle_time"`
	MaxLifetime  time.Duration `json:"max_lifetime"`

	// Driver-specific options
	Options      map[string]string `json:"options"`

	// Environment
	Environment  string        `json:"environment"` // dev, staging, prod
	Tags         []string      `json:"tags"`

	// For SQLite
	FilePath     string        `json:"file_path,omitempty"`
}

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypeMariaDB    DatabaseType = "mariadb"
)

// Validate checks if the configuration is valid
func (c *ConnectionConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("connection name is required")
	}

	switch c.Type {
	case DatabaseTypeSQLite:
		if c.FilePath == "" {
			return fmt.Errorf("file_path is required for SQLite")
		}
	case DatabaseTypePostgreSQL, DatabaseTypeMySQL, DatabaseTypeMariaDB:
		if c.Host == "" {
			return fmt.Errorf("host is required for %s", c.Type)
		}
		if c.Database == "" {
			return fmt.Errorf("database name is required")
		}
		if c.Port == 0 {
			c.Port = c.Type.DefaultPort()
		}
	default:
		return fmt.Errorf("unsupported database type: %s", c.Type)
	}

	// Set defaults
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	if c.MaxConns == 0 {
		c.MaxConns = 10
	}
	if c.MinConns == 0 {
		c.MinConns = 2
	}

	return nil
}

// DefaultPort returns the default port for the database type
func (dt DatabaseType) DefaultPort() int {
	switch dt {
	case DatabaseTypePostgreSQL:
		return 5432
	case DatabaseTypeMySQL, DatabaseTypeMariaDB:
		return 3306
	case DatabaseTypeSQLite:
		return 0 // File-based, no port
	default:
		return 0
	}
}

// ConnectionString builds a driver-specific connection string
func (c *ConnectionConfig) ConnectionString() (string, error) {
	switch c.Type {
	case DatabaseTypePostgreSQL:
		return c.postgresConnectionString(), nil
	case DatabaseTypeMySQL, DatabaseTypeMariaDB:
		return c.mysqlConnectionString(), nil
	case DatabaseTypeSQLite:
		return c.sqliteConnectionString(), nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", c.Type)
	}
}

func (c *ConnectionConfig) postgresConnectionString() string {
	// Format: postgres://user:password@host:port/database?sslmode=disable
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.Username, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Database,
	}

	q := u.Query()
	if c.SSLMode != "" {
		q.Set("sslmode", c.SSLMode)
	} else {
		q.Set("sslmode", "disable")
	}
	q.Set("connect_timeout", fmt.Sprintf("%.0f", c.Timeout.Seconds()))

	// Add custom options
	for k, v := range c.Options {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func (c *ConnectionConfig) mysqlConnectionString() string {
	// Format: user:password@tcp(host:port)/database?parseTime=true
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Timeout.String(),
	)

	// Add custom options
	for k, v := range c.Options {
		dsn += fmt.Sprintf("&%s=%s", k, url.QueryEscape(v))
	}

	return dsn
}

func (c *ConnectionConfig) sqliteConnectionString() string {
	// Format: file:path/to/database.db?cache=shared&mode=rwc
	u := url.URL{
		Scheme: "file",
		Path:   c.FilePath,
	}

	q := u.Query()
	q.Set("cache", "shared")
	q.Set("mode", "rwc") // read-write-create
	q.Set("_timeout", fmt.Sprintf("%.0f", c.Timeout.Milliseconds()))

	// Add custom options
	for k, v := range c.Options {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String()
}
```

### Query Results

```go
// ResultSet represents query execution results
type ResultSet struct {
	Columns      []Column      `json:"columns"`
	Rows         []Row         `json:"rows"`
	RowCount     int64         `json:"row_count"`
	AffectedRows int64         `json:"affected_rows"`
	LastInsertID int64         `json:"last_insert_id,omitempty"`
	ExecutionTime time.Duration `json:"execution_time"`
	Query        string        `json:"query"`
	Error        error         `json:"error,omitempty"`
}

// Column represents a result set column
type Column struct {
	Name         string   `json:"name"`
	Type         DataType `json:"type"`
	NativeType   string   `json:"native_type"`
	Nullable     bool     `json:"nullable"`
	IsPrimaryKey bool     `json:"is_primary_key"`
	IsAutoIncrement bool  `json:"is_auto_increment"`
	MaxLength    int      `json:"max_length,omitempty"`
	Precision    int      `json:"precision,omitempty"`
	Scale        int      `json:"scale,omitempty"`
}

// Row represents a single row of data
type Row []interface{}

// DataType represents a normalized data type across databases
type DataType string

const (
	DataTypeString    DataType = "string"
	DataTypeInteger   DataType = "integer"
	DataTypeFloat     DataType = "float"
	DataTypeBoolean   DataType = "boolean"
	DataTypeDateTime  DataType = "datetime"
	DataTypeDate      DataType = "date"
	DataTypeTime      DataType = "time"
	DataTypeJSON      DataType = "json"
	DataTypeUUID      DataType = "uuid"
	DataTypeBinary    DataType = "binary"
	DataTypeUnknown   DataType = "unknown"
)

// Get returns the value at the given column index
func (r Row) Get(index int) interface{} {
	if index < 0 || index >= len(r) {
		return nil
	}
	return r[index]
}

// GetString returns the value as a string
func (r Row) GetString(index int) string {
	v := r.Get(index)
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// AsMap converts the row to a map using column names
func (r Row) AsMap(columns []Column) map[string]interface{} {
	m := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		m[col.Name] = r.Get(i)
	}
	return m
}
```

### Schema Types

```go
// Schema represents a database schema/namespace
type Schema struct {
	Name    string `json:"name"`
	Owner   string `json:"owner,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// Table represents a database table
type Table struct {
	Schema      string    `json:"schema"`
	Name        string    `json:"name"`
	Type        TableType `json:"type"`
	RowCount    int64     `json:"row_count,omitempty"`
	SizeBytes   int64     `json:"size_bytes,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type TableType string

const (
	TableTypeTable      TableType = "table"
	TableTypeView       TableType = "view"
	TableTypeMaterialized TableType = "materialized_view"
	TableTypeTemporary  TableType = "temporary"
)

// View represents a database view
type View struct {
	Schema     string    `json:"schema"`
	Name       string    `json:"name"`
	Definition string    `json:"definition"`
	Comment    string    `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

// TableSchema represents the full schema of a table
type TableSchema struct {
	Table       Table        `json:"table"`
	Columns     []ColumnInfo `json:"columns"`
	Indexes     []Index      `json:"indexes"`
	Constraints []Constraint `json:"constraints"`
}

// ColumnInfo represents detailed column information
type ColumnInfo struct {
	Name            string   `json:"name"`
	Position        int      `json:"position"`
	DataType        DataType `json:"data_type"`
	NativeType      string   `json:"native_type"`
	Nullable        bool     `json:"nullable"`
	DefaultValue    *string  `json:"default_value,omitempty"`
	IsPrimaryKey    bool     `json:"is_primary_key"`
	IsAutoIncrement bool     `json:"is_auto_increment"`
	MaxLength       int      `json:"max_length,omitempty"`
	Precision       int      `json:"precision,omitempty"`
	Scale           int      `json:"scale,omitempty"`
	Comment         string   `json:"comment,omitempty"`
}

// Index represents a database index
type Index struct {
	Name       string   `json:"name"`
	Table      string   `json:"table"`
	Columns    []string `json:"columns"`
	IsUnique   bool     `json:"is_unique"`
	IsPrimary  bool     `json:"is_primary"`
	Type       string   `json:"type,omitempty"` // btree, hash, gin, etc.
	Condition  string   `json:"condition,omitempty"` // partial index condition
}

// Constraint represents a table constraint
type Constraint struct {
	Name           string         `json:"name"`
	Type           ConstraintType `json:"type"`
	Table          string         `json:"table"`
	Columns        []string       `json:"columns"`

	// Foreign key specific
	ReferencedTable   string   `json:"referenced_table,omitempty"`
	ReferencedColumns []string `json:"referenced_columns,omitempty"`
	OnDelete          string   `json:"on_delete,omitempty"`
	OnUpdate          string   `json:"on_update,omitempty"`

	// Check constraint specific
	CheckClause string `json:"check_clause,omitempty"`
}

type ConstraintType string

const (
	ConstraintTypePrimaryKey ConstraintType = "PRIMARY KEY"
	ConstraintTypeForeignKey ConstraintType = "FOREIGN KEY"
	ConstraintTypeUnique     ConstraintType = "UNIQUE"
	ConstraintTypeCheck      ConstraintType = "CHECK"
	ConstraintTypeNotNull    ConstraintType = "NOT NULL"
)
```

### Statistics

```go
// Stats represents database connection and performance statistics
type Stats struct {
	// Connection stats
	MaxConns        int           `json:"max_conns"`
	OpenConns       int           `json:"open_conns"`
	InUseConns      int           `json:"in_use_conns"`
	IdleConns       int           `json:"idle_conns"`

	// Query stats
	QueriesExecuted int64         `json:"queries_executed"`
	QueriesFailed   int64         `json:"queries_failed"`
	AverageQueryTime time.Duration `json:"average_query_time"`

	// Database stats
	DatabaseSize    int64         `json:"database_size_bytes,omitempty"`
	TableCount      int           `json:"table_count,omitempty"`

	// Timestamps
	ConnectedAt     time.Time     `json:"connected_at"`
	LastQueryAt     time.Time     `json:"last_query_at"`
}
```

---

## Error Handling

### Error Types

```go
// DBError represents a database error with context
type DBError struct {
	Type       ErrorType
	Message    string
	Query      string
	Params     []interface{}
	Underlying error
	Severity   Severity
}

type ErrorType string

const (
	ErrorTypeConnection       ErrorType = "connection"
	ErrorTypeQuery            ErrorType = "query"
	ErrorTypeSyntax           ErrorType = "syntax"
	ErrorTypeConstraint       ErrorType = "constraint"
	ErrorTypeNotFound         ErrorType = "not_found"
	ErrorTypePermission       ErrorType = "permission"
	ErrorTypeTimeout          ErrorType = "timeout"
	ErrorTypeTransaction      ErrorType = "transaction"
	ErrorTypeUnsupported      ErrorType = "unsupported"
	ErrorTypeInternal         ErrorType = "internal"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
	SeverityCritical Severity = "critical"
)

func (e *DBError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Message, e.Underlying)
}

func (e *DBError) Unwrap() error {
	return e.Underlying
}

// IsConnectionError checks if the error is connection-related
func IsConnectionError(err error) bool {
	var dbErr *DBError
	if errors.As(err, &dbErr) {
		return dbErr.Type == ErrorTypeConnection
	}
	return false
}

// IsQueryError checks if the error is query-related
func IsQueryError(err error) bool {
	var dbErr *DBError
	if errors.As(err, &dbErr) {
		return dbErr.Type == ErrorTypeQuery || dbErr.Type == ErrorTypeSyntax
	}
	return false
}

// NewConnectionError creates a connection error
func NewConnectionError(msg string, err error) *DBError {
	return &DBError{
		Type:       ErrorTypeConnection,
		Message:    msg,
		Underlying: err,
		Severity:   SeverityError,
	}
}

// NewQueryError creates a query error
func NewQueryError(query string, err error) *DBError {
	return &DBError{
		Type:       ErrorTypeQuery,
		Message:    "query execution failed",
		Query:      query,
		Underlying: err,
		Severity:   SeverityError,
	}
}
```

---

## PostgreSQL Implementation

### PostgresDB Structure

```go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	config   *ConnectionConfig
	pool     *pgxpool.Pool
	dialect  Dialect
	stats    *Stats
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB() *PostgresDB {
	return &PostgresDB{
		dialect: &PostgresDialect{},
		stats:   &Stats{},
	}
}

// Connect establishes a connection to PostgreSQL
func (db *PostgresDB) Connect(ctx context.Context, config ConnectionConfig) error {
	if err := config.Validate(); err != nil {
		return NewConnectionError("invalid configuration", err)
	}

	db.config = &config

	// Build connection string
	connString, err := config.ConnectionString()
	if err != nil {
		return NewConnectionError("failed to build connection string", err)
	}

	// Create pool config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return NewConnectionError("failed to parse connection string", err)
	}

	// Set pool settings
	poolConfig.MaxConns = int32(config.MaxConns)
	poolConfig.MinConns = int32(config.MinConns)
	poolConfig.MaxConnIdleTime = config.MaxIdleTime
	poolConfig.MaxConnLifetime = config.MaxLifetime

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return NewConnectionError("failed to create connection pool", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return NewConnectionError("failed to ping database", err)
	}

	db.pool = pool
	db.stats.ConnectedAt = time.Now()

	return nil
}

// Disconnect closes the database connection
func (db *PostgresDB) Disconnect(ctx context.Context) error {
	if db.pool != nil {
		db.pool.Close()
		db.pool = nil
	}
	return nil
}

// Ping checks if the database is reachable
func (db *PostgresDB) Ping(ctx context.Context) error {
	if db.pool == nil {
		return NewConnectionError("not connected", nil)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.pool.Ping(ctx); err != nil {
		return NewConnectionError("ping failed", err)
	}

	return nil
}

// IsConnected checks if the database is currently connected
func (db *PostgresDB) IsConnected() bool {
	return db.pool != nil && db.Ping(context.Background()) == nil
}

// Execute runs a query and returns results
func (db *PostgresDB) Execute(ctx context.Context, query string, params ...interface{}) (*ResultSet, error) {
	if db.pool == nil {
		return nil, NewConnectionError("not connected", nil)
	}

	startTime := time.Now()

	// Execute query
	rows, err := db.pool.Query(ctx, query, params...)
	if err != nil {
		db.stats.QueriesFailed++
		return nil, NewQueryError(query, err)
	}
	defer rows.Close()

	// Build result set
	result := &ResultSet{
		Query: query,
	}

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	result.Columns = make([]Column, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		result.Columns[i] = Column{
			Name:       string(fd.Name),
			NativeType: fd.DataTypeOID.String(),
			Type:       db.dialect.DataTypeMapping(fd.DataTypeOID.String()),
		}
	}

	// Collect rows
	result.Rows = make([]Row, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, NewQueryError(query, err)
		}
		result.Rows = append(result.Rows, values)
	}

	if err := rows.Err(); err != nil {
		return nil, NewQueryError(query, err)
	}

	result.RowCount = int64(len(result.Rows))
	result.ExecutionTime = time.Since(startTime)

	// Update stats
	db.stats.QueriesExecuted++
	db.stats.LastQueryAt = time.Now()

	return result, nil
}

// ListTables returns all tables in the specified schema
func (db *PostgresDB) ListTables(ctx context.Context, schema string) ([]Table, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT
			schemaname,
			tablename,
			'table' as type,
			NULL as row_count
		FROM pg_tables
		WHERE schemaname = $1
		ORDER BY tablename
	`

	result, err := db.Execute(ctx, query, schema)
	if err != nil {
		return nil, err
	}

	tables := make([]Table, len(result.Rows))
	for i, row := range result.Rows {
		tables[i] = Table{
			Schema: row.GetString(0),
			Name:   row.GetString(1),
			Type:   TableType(row.GetString(2)),
		}
	}

	return tables, nil
}

// DescribeTable returns detailed schema information for a table
func (db *PostgresDB) DescribeTable(ctx context.Context, schema, table string) (*TableSchema, error) {
	query := `
		SELECT
			column_name,
			ordinal_position,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`

	result, err := db.Execute(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}

	tableSchema := &TableSchema{
		Table: Table{
			Schema: schema,
			Name:   table,
			Type:   TableTypeTable,
		},
		Columns: make([]ColumnInfo, len(result.Rows)),
	}

	for i, row := range result.Rows {
		nullable := row.GetString(3) == "YES"

		tableSchema.Columns[i] = ColumnInfo{
			Name:       row.GetString(0),
			Position:   i + 1,
			NativeType: row.GetString(2),
			DataType:   db.dialect.DataTypeMapping(row.GetString(2)),
			Nullable:   nullable,
		}

		// Get default value if present
		if defaultVal := row.GetString(4); defaultVal != "" {
			tableSchema.Columns[i].DefaultValue = &defaultVal
		}
	}

	return tableSchema, nil
}

// GetDriverName returns the driver name
func (db *PostgresDB) GetDriverName() string {
	return "postgres"
}

// GetDialect returns the database dialect
func (db *PostgresDB) GetDialect() Dialect {
	return db.dialect
}

// GetStats returns connection and query statistics
func (db *PostgresDB) GetStats(ctx context.Context) (*Stats, error) {
	if db.pool == nil {
		return nil, NewConnectionError("not connected", nil)
	}

	poolStats := db.pool.Stat()

	db.stats.MaxConns = int(poolStats.MaxConns())
	db.stats.OpenConns = int(poolStats.TotalConns())
	db.stats.InUseConns = int(poolStats.AcquiredConns())
	db.stats.IdleConns = int(poolStats.IdleConns())

	return db.stats, nil
}
```

### PostgreSQL Dialect

```go
// PostgresDialect implements PostgreSQL-specific SQL dialect
type PostgresDialect struct{}

func (d *PostgresDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func (d *PostgresDialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

func (d *PostgresDialect) LimitClause(limit, offset int64) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *PostgresDialect) CurrentTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

func (d *PostgresDialect) DataTypeMapping(nativeType string) DataType {
	switch nativeType {
	case "integer", "int", "int4", "smallint", "int2", "bigint", "int8":
		return DataTypeInteger
	case "real", "float4", "double precision", "float8", "numeric", "decimal":
		return DataTypeFloat
	case "boolean", "bool":
		return DataTypeBoolean
	case "timestamp", "timestamptz", "timestamp with time zone", "timestamp without time zone":
		return DataTypeDateTime
	case "date":
		return DataTypeDate
	case "time", "timetz", "time with time zone", "time without time zone":
		return DataTypeTime
	case "json", "jsonb":
		return DataTypeJSON
	case "uuid":
		return DataTypeUUID
	case "bytea":
		return DataTypeBinary
	case "text", "varchar", "character varying", "char", "character":
		return DataTypeString
	default:
		return DataTypeUnknown
	}
}
```

---

## MySQL Implementation

### MySQLDB Structure

```go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDB implements the Database interface for MySQL
type MySQLDB struct {
	config  *ConnectionConfig
	db      *sql.DB
	dialect Dialect
	stats   *Stats
}

// NewMySQLDB creates a new MySQL database connection
func NewMySQLDB() *MySQLDB {
	return &MySQLDB{
		dialect: &MySQLDialect{},
		stats:   &Stats{},
	}
}

// Connect establishes a connection to MySQL
func (db *MySQLDB) Connect(ctx context.Context, config ConnectionConfig) error {
	if err := config.Validate(); err != nil {
		return NewConnectionError("invalid configuration", err)
	}

	db.config = &config

	// Build connection string
	connString, err := config.ConnectionString()
	if err != nil {
		return NewConnectionError("failed to build connection string", err)
	}

	// Open connection
	sqlDB, err := sql.Open("mysql", connString)
	if err != nil {
		return NewConnectionError("failed to open connection", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(config.MaxConns)
	sqlDB.SetMaxIdleConns(config.MinConns)
	sqlDB.SetConnMaxIdleTime(config.MaxIdleTime)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return NewConnectionError("failed to ping database", err)
	}

	db.db = sqlDB
	db.stats.ConnectedAt = time.Now()

	return nil
}

// Execute runs a query and returns results
func (db *MySQLDB) Execute(ctx context.Context, query string, params ...interface{}) (*ResultSet, error) {
	if db.db == nil {
		return nil, NewConnectionError("not connected", nil)
	}

	startTime := time.Now()

	// Execute query
	rows, err := db.db.QueryContext(ctx, query, params...)
	if err != nil {
		db.stats.QueriesFailed++
		return nil, NewQueryError(query, err)
	}
	defer rows.Close()

	// Build result set
	result := &ResultSet{
		Query: query,
	}

	// Get column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, NewQueryError(query, err)
	}

	result.Columns = make([]Column, len(columnTypes))
	for i, ct := range columnTypes {
		nullable, _ := ct.Nullable()

		result.Columns[i] = Column{
			Name:       ct.Name(),
			NativeType: ct.DatabaseTypeName(),
			Type:       db.dialect.DataTypeMapping(ct.DatabaseTypeName()),
			Nullable:   nullable,
		}

		if length, ok := ct.Length(); ok {
			result.Columns[i].MaxLength = int(length)
		}
	}

	// Collect rows
	result.Rows = make([]Row, 0)
	for rows.Next() {
		// Create slice of interface{} for scanning
		values := make([]interface{}, len(columnTypes))
		valuePtrs := make([]interface{}, len(columnTypes))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, NewQueryError(query, err)
		}

		result.Rows = append(result.Rows, values)
	}

	if err := rows.Err(); err != nil {
		return nil, NewQueryError(query, err)
	}

	result.RowCount = int64(len(result.Rows))
	result.ExecutionTime = time.Since(startTime)

	// Update stats
	db.stats.QueriesExecuted++
	db.stats.LastQueryAt = time.Now()

	return result, nil
}

// ListTables returns all tables in the specified schema
func (db *MySQLDB) ListTables(ctx context.Context, schema string) ([]Table, error) {
	if schema == "" {
		schema = db.config.Database
	}

	query := `
		SELECT
			TABLE_SCHEMA,
			TABLE_NAME,
			TABLE_TYPE,
			TABLE_ROWS
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`

	result, err := db.Execute(ctx, query, schema)
	if err != nil {
		return nil, err
	}

	tables := make([]Table, len(result.Rows))
	for i, row := range result.Rows {
		tables[i] = Table{
			Schema:   row.GetString(0),
			Name:     row.GetString(1),
			Type:     TableType(row.GetString(2)),
			RowCount: int64(row.Get(3).(int64)),
		}
	}

	return tables, nil
}

// GetDriverName returns the driver name
func (db *MySQLDB) GetDriverName() string {
	return "mysql"
}

// GetDialect returns the database dialect
func (db *MySQLDB) GetDialect() Dialect {
	return db.dialect
}
```

### MySQL Dialect

```go
// MySQLDialect implements MySQL-specific SQL dialect
type MySQLDialect struct{}

func (d *MySQLDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf("`%s`", name)
}

func (d *MySQLDialect) Placeholder(n int) string {
	return "?"
}

func (d *MySQLDialect) LimitClause(limit, offset int64) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d, %d", offset, limit)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *MySQLDialect) CurrentTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

func (d *MySQLDialect) DataTypeMapping(nativeType string) DataType {
	switch nativeType {
	case "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT":
		return DataTypeInteger
	case "FLOAT", "DOUBLE", "DECIMAL":
		return DataTypeFloat
	case "BOOL", "BOOLEAN":
		return DataTypeBoolean
	case "DATETIME", "TIMESTAMP":
		return DataTypeDateTime
	case "DATE":
		return DataTypeDate
	case "TIME":
		return DataTypeTime
	case "JSON":
		return DataTypeJSON
	case "BINARY", "VARBINARY", "BLOB":
		return DataTypeBinary
	case "VARCHAR", "CHAR", "TEXT", "TINYTEXT", "MEDIUMTEXT", "LONGTEXT":
		return DataTypeString
	default:
		return DataTypeUnknown
	}
}
```

---

## SQLite Implementation

### SQLiteDB Structure

```go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB implements the Database interface for SQLite
type SQLiteDB struct {
	config  *ConnectionConfig
	db      *sql.DB
	dialect Dialect
	stats   *Stats
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB() *SQLiteDB {
	return &SQLiteDB{
		dialect: &SQLiteDialect{},
		stats:   &Stats{},
	}
}

// Connect establishes a connection to SQLite
func (db *SQLiteDB) Connect(ctx context.Context, config ConnectionConfig) error {
	if err := config.Validate(); err != nil {
		return NewConnectionError("invalid configuration", err)
	}

	db.config = &config

	// Build connection string
	connString, err := config.ConnectionString()
	if err != nil {
		return NewConnectionError("failed to build connection string", err)
	}

	// Open connection
	sqlDB, err := sql.Open("sqlite3", connString)
	if err != nil {
		return NewConnectionError("failed to open connection", err)
	}

	// SQLite doesn't support connection pooling in the traditional sense
	sqlDB.SetMaxOpenConns(1)

	// Test connection
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return NewConnectionError("failed to ping database", err)
	}

	db.db = sqlDB
	db.stats.ConnectedAt = time.Now()

	return nil
}

// Execute runs a query and returns results
func (db *SQLiteDB) Execute(ctx context.Context, query string, params ...interface{}) (*ResultSet, error) {
	// Implementation similar to MySQLDB
	// ...
	return nil, nil
}

// ListTables returns all tables
func (db *SQLiteDB) ListTables(ctx context.Context, schema string) ([]Table, error) {
	query := `
		SELECT
			name,
			type
		FROM sqlite_master
		WHERE type IN ('table', 'view')
		ORDER BY name
	`

	result, err := db.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	tables := make([]Table, len(result.Rows))
	for i, row := range result.Rows {
		tables[i] = Table{
			Schema: "main",
			Name:   row.GetString(0),
			Type:   TableType(row.GetString(1)),
		}
	}

	return tables, nil
}

// GetDriverName returns the driver name
func (db *SQLiteDB) GetDriverName() string {
	return "sqlite"
}

// GetDialect returns the database dialect
func (db *SQLiteDB) GetDialect() Dialect {
	return db.dialect
}
```

### SQLite Dialect

```go
// SQLiteDialect implements SQLite-specific SQL dialect
type SQLiteDialect struct{}

func (d *SQLiteDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func (d *SQLiteDialect) Placeholder(n int) string {
	return "?"
}

func (d *SQLiteDialect) LimitClause(limit, offset int64) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *SQLiteDialect) CurrentTimestamp() string {
	return "DATETIME('now')"
}

func (d *SQLiteDialect) DataTypeMapping(nativeType string) DataType {
	switch nativeType {
	case "INTEGER":
		return DataTypeInteger
	case "REAL", "NUMERIC":
		return DataTypeFloat
	case "BOOLEAN":
		return DataTypeBoolean
	case "TEXT":
		return DataTypeString
	case "BLOB":
		return DataTypeBinary
	default:
		return DataTypeUnknown
	}
}
```

---

## Factory Pattern

### Database Factory

```go
package db

// Factory creates database instances based on type
type Factory struct{}

// NewFactory creates a new database factory
func NewFactory() *Factory {
	return &Factory{}
}

// Create creates a new database instance
func (f *Factory) Create(dbType DatabaseType) (Database, error) {
	switch dbType {
	case DatabaseTypePostgreSQL:
		return NewPostgresDB(), nil
	case DatabaseTypeMySQL, DatabaseTypeMariaDB:
		return NewMySQLDB(), nil
	case DatabaseTypeSQLite:
		return NewSQLiteDB(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// CreateAndConnect creates a database instance and connects
func (f *Factory) CreateAndConnect(ctx context.Context, config ConnectionConfig) (Database, error) {
	db, err := f.Create(config.Type)
	if err != nil {
		return nil, err
	}

	if err := db.Connect(ctx, config); err != nil {
		return nil, err
	}

	return db, nil
}
```

---

## Testing

### Mock Database

```go
package testing

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	ConnectFunc      func(ctx context.Context, config db.ConnectionConfig) error
	ExecuteFunc      func(ctx context.Context, query string, params ...interface{}) (*db.ResultSet, error)
	ListTablesFunc   func(ctx context.Context, schema string) ([]db.Table, error)
	DisconnectFunc   func(ctx context.Context) error
	IsConnectedFunc  func() bool
}

func (m *MockDatabase) Connect(ctx context.Context, config db.ConnectionConfig) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(ctx, config)
	}
	return nil
}

func (m *MockDatabase) Execute(ctx context.Context, query string, params ...interface{}) (*db.ResultSet, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, query, params...)
	}
	return &db.ResultSet{}, nil
}

// ... implement other interface methods
```

### Test Fixtures

```go
package testing

// Fixtures provides test data for database testing
type Fixtures struct{}

// SampleConnectionConfig returns a test connection config
func (f *Fixtures) SampleConnectionConfig() db.ConnectionConfig {
	return db.ConnectionConfig{
		Name:     "test",
		Type:     db.DatabaseTypePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "test",
		Password: "test",
	}
}

// SampleResultSet returns a test result set
func (f *Fixtures) SampleResultSet() *db.ResultSet {
	return &db.ResultSet{
		Columns: []db.Column{
			{Name: "id", Type: db.DataTypeInteger},
			{Name: "name", Type: db.DataTypeString},
		},
		Rows: []db.Row{
			{1, "Alice"},
			{2, "Bob"},
		},
		RowCount: 2,
	}
}
```

---

## See Also

- [Architecture](./ARCHITECTURE.md) - System architecture overview
- [Implementation Plan](./IMPLEMENTATION_PLAN.md) - Development timeline
- [PRD](./PRD.md) - Product requirements

---

**Document Version**: v1.0
**Last Updated**: 2024-01
**Applies to LazyDB**: v0.1.0-dev (MVP Phase)
