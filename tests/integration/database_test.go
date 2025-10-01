package integration

import (
	"context"
	"os"
	"testing"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// These tests require a running PostgreSQL instance
// Skip if TEST_POSTGRES_DSN is not set

func getTestDSN() string {
	return os.Getenv("TEST_POSTGRES_DSN")
}

func TestPostgresConnection(t *testing.T) {
	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("Skipping integration test: TEST_POSTGRES_DSN not set")
	}

	config := db.ConnectionConfig{
		Name:     "test-connection",
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		Username: "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	}

	conn := db.NewPostgresConnection(config)
	ctx := context.Background()

	// Test connection
	err := conn.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Disconnect(ctx)

	// Test status
	if conn.Status() != db.StatusConnected {
		t.Errorf("Expected status Connected, got %v", conn.Status())
	}

	// Test ping
	err = conn.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestQueryExecution(t *testing.T) {
	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("Skipping integration test: TEST_POSTGRES_DSN not set")
	}

	config := db.ConnectionConfig{
		Name:     "test-query",
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		Username: "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	}

	conn := db.NewPostgresConnection(config)
	ctx := context.Background()

	err := conn.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Disconnect(ctx)

	// Test simple query
	result := db.ExecuteQuery(ctx, conn.Conn(), "SELECT 1 as num, 'test' as text")

	if result.Error != nil {
		t.Fatalf("Query execution failed: %v", result.Error)
	}

	if result.RowCount != 1 {
		t.Errorf("Expected 1 row, got %d", result.RowCount)
	}

	if len(result.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(result.Columns))
	}
}
