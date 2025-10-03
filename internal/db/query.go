package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// QueryResult represents the result of a database query
type QueryResult struct {
	Columns      []string
	Rows         [][]string
	RowCount     int
	ExecutionMs  int64
	Error        error
}

// ExecuteQuery executes a SQL query and returns the results
func ExecuteQuery(ctx context.Context, conn *pgx.Conn, query string) QueryResult {
	startTime := time.Now()

	result := QueryResult{
		Columns: []string{},
		Rows:    [][]string{},
	}

	// Trim and normalize query
	trimmedQuery := strings.TrimSpace(query)
	upperQuery := strings.ToUpper(trimmedQuery)

	// Check if it's a SELECT-like query (returns rows)
	isSelectQuery := strings.HasPrefix(upperQuery, "SELECT") ||
		strings.HasPrefix(upperQuery, "WITH") ||
		strings.HasPrefix(upperQuery, "SHOW") ||
		strings.HasPrefix(upperQuery, "EXPLAIN")

	// Check if multiple statements (count semicolons, excluding trailing one)
	queryWithoutTrailing := strings.TrimRight(trimmedQuery, "; \t\n")
	isMultiStatement := strings.Count(queryWithoutTrailing, ";") > 0

	// Use Exec for multiple statements or non-SELECT queries
	if isMultiStatement || !isSelectQuery {
		// Use Exec() which supports multiple statements via simple protocol
		commandTag, err := conn.Exec(ctx, query)
		if err != nil {
			result.Error = err
			result.ExecutionMs = time.Since(startTime).Milliseconds()
			return result
		}

		// Return success message
		result.Columns = []string{"status"}
		result.Rows = [][]string{{fmt.Sprintf("Success: %s", commandTag.String())}}
		result.RowCount = 1
		result.ExecutionMs = time.Since(startTime).Milliseconds()
		return result
	}

	// Use Query() for single SELECT statements
	rows, err := conn.Query(ctx, query)
	if err != nil {
		result.Error = err
		result.ExecutionMs = time.Since(startTime).Milliseconds()
		return result
	}
	defer rows.Close()

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	for _, fd := range fieldDescriptions {
		result.Columns = append(result.Columns, string(fd.Name))
	}

	// Fetch all rows
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			result.Error = err
			result.ExecutionMs = time.Since(startTime).Milliseconds()
			return result
		}

		// Convert values to strings
		rowStrings := make([]string, len(values))
		for i, v := range values {
			rowStrings[i] = fmt.Sprintf("%v", v)
		}
		result.Rows = append(result.Rows, rowStrings)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		result.Error = err
		result.ExecutionMs = time.Since(startTime).Milliseconds()
		return result
	}

	result.RowCount = len(result.Rows)
	result.ExecutionMs = time.Since(startTime).Milliseconds()

	return result
}
