package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GetQueriesDir returns the directory for saved queries
func GetQueriesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	queriesDir := filepath.Join(homeDir, ".lazydb", "queries")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(queriesDir, 0755); err != nil {
		return "", err
	}

	return queriesDir, nil
}

// SaveQuery saves a query to a file
func SaveQuery(query string, filename string) error {
	queriesDir, err := GetQueriesDir()
	if err != nil {
		return err
	}

	// If no filename provided, generate one with timestamp
	if filename == "" {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename = fmt.Sprintf("query_%s.sql", timestamp)
	}

	// Ensure .sql extension
	if filepath.Ext(filename) != ".sql" {
		filename += ".sql"
	}

	filePath := filepath.Join(queriesDir, filename)

	return os.WriteFile(filePath, []byte(query), 0644)
}

// LoadQuery loads a query from a file
func LoadQuery(filename string) (string, error) {
	queriesDir, err := GetQueriesDir()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(queriesDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ListQueries returns a list of saved query files
func ListQueries() ([]string, error) {
	queriesDir, err := GetQueriesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(queriesDir)
	if err != nil {
		return nil, err
	}

	var queries []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			queries = append(queries, entry.Name())
		}
	}

	return queries, nil
}
