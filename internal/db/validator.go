package db

import (
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

// ValidationError represents a SQL validation error
type ValidationError struct {
	Line    int
	Column  int
	Message string
	Query   string
}

// ValidationResult contains the result of SQL validation
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// SQLValidator validates SQL queries using PostgreSQL parser
type SQLValidator struct{}

// NewSQLValidator creates a new SQL validator
func NewSQLValidator() *SQLValidator {
	return &SQLValidator{}
}

// Validate checks if the SQL query is valid
func (v *SQLValidator) Validate(query string) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Skip validation for empty queries
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return result
	}

	// Parse the query using PostgreSQL parser
	_, err := pg_query.Parse(query)
	if err != nil {
		result.Valid = false

		// Extract error details
		errMsg := err.Error()
		line, column := extractErrorPosition(errMsg)

		result.Errors = append(result.Errors, ValidationError{
			Line:    line,
			Column:  column,
			Message: cleanErrorMessage(errMsg),
			Query:   query,
		})
	}

	return result
}

// ValidateAndFormat validates and formats the query (formatting not yet implemented)
func (v *SQLValidator) ValidateAndFormat(query string) (string, ValidationResult) {
	result := v.Validate(query)
	// TODO: Add formatting when available
	return query, result
}

// extractErrorPosition extracts line and column from error message
func extractErrorPosition(errMsg string) (line int, column int) {
	// Default to line 1, column 0
	line, column = 1, 0

	// Try to extract position from error message
	// Format: "ERROR: syntax error at or near ... (SQLSTATE ...)"
	// Sometimes includes "at line X"

	if strings.Contains(errMsg, "at line") {
		// Try to parse "at line X"
		parts := strings.Split(errMsg, "at line")
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &line)
		}
	}

	return line, column
}

// cleanErrorMessage removes SQLSTATE and makes error more readable
func cleanErrorMessage(errMsg string) string {
	// Remove SQLSTATE prefix if present
	errMsg = strings.TrimPrefix(errMsg, "ERROR: ")

	// Remove SQLSTATE suffix if present
	if idx := strings.Index(errMsg, "(SQLSTATE"); idx != -1 {
		errMsg = strings.TrimSpace(errMsg[:idx])
	}

	return errMsg
}

// GetSuggestions provides helpful suggestions based on error type
func (v *SQLValidator) GetSuggestions(err ValidationError) []string {
	suggestions := []string{}

	lowerMsg := strings.ToLower(err.Message)

	// Common error patterns and suggestions
	if strings.Contains(lowerMsg, "syntax error") {
		suggestions = append(suggestions, "Check for missing semicolons or quotes")
		suggestions = append(suggestions, "Verify SQL keyword spelling")
	}

	if strings.Contains(lowerMsg, "relation") && strings.Contains(lowerMsg, "does not exist") {
		suggestions = append(suggestions, "Check table name spelling")
		suggestions = append(suggestions, "Verify table exists in current schema")
	}

	if strings.Contains(lowerMsg, "column") && strings.Contains(lowerMsg, "does not exist") {
		suggestions = append(suggestions, "Check column name spelling")
		suggestions = append(suggestions, "Verify column exists in table")
	}

	if strings.Contains(lowerMsg, "unterminated") {
		suggestions = append(suggestions, "Check for unclosed quotes or parentheses")
	}

	return suggestions
}
