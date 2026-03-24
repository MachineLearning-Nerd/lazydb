package unit

import (
	"testing"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

func TestIsReadOnlyQuery_AllowsSelect(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"simple select", "SELECT * FROM users"},
		{"select with where", "SELECT id, name FROM users WHERE id = 1"},
		{"select with limit", "SELECT * FROM users LIMIT 10"},
		{"select with join", "SELECT u.*, o.total FROM users u JOIN orders o ON u.id = o.user_id"},
		{"CTE select", "WITH cte AS (SELECT * FROM users) SELECT * FROM cte"},
		{"recursive CTE", "WITH RECURSIVE tree AS (SELECT id, parent_id FROM categories WHERE parent_id IS NULL UNION ALL SELECT c.id, c.parent_id FROM categories c JOIN tree t ON c.parent_id = t.id) SELECT * FROM tree"},
		{"select constant", "SELECT 1"},
		{"union select", "(SELECT 1) UNION (SELECT 2)"},
		{"comment embedded semicolon", "SELECT 1 /* ; DROP TABLE users; */"},
		{"table name containing limit", "SELECT * FROM rate_limit_events"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.IsReadOnlyQuery(tt.query)
			if err != nil {
				t.Errorf("IsReadOnlyQuery(%q) returned error: %v, expected nil", tt.query, err)
			}
		})
	}
}

func TestIsReadOnlyQuery_RejectsDML(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"INSERT", "INSERT INTO users (name) VALUES ('test')"},
		{"UPDATE", "UPDATE users SET name = 'test' WHERE id = 1"},
		{"DELETE", "DELETE FROM users WHERE id = 1"},
		{"DROP TABLE", "DROP TABLE users"},
		{"CREATE TABLE", "CREATE TABLE test (id int)"},
		{"ALTER TABLE", "ALTER TABLE users ADD COLUMN email text"},
		{"TRUNCATE", "TRUNCATE users"},
		{"empty query", ""},
		{"multi-statement injection", "SELECT 1; DROP TABLE users"},
		{"SELECT INTO", "SELECT * INTO new_table FROM users"},
		{"CTAS", "CREATE TABLE evil AS SELECT * FROM users"},
		{"EXPLAIN", "EXPLAIN SELECT * FROM users"},
		{"COPY", "COPY users TO '/tmp/data.csv'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.IsReadOnlyQuery(tt.query)
			if err == nil {
				t.Errorf("IsReadOnlyQuery(%q) returned nil, expected error", tt.query)
			}
		})
	}
}

func TestHasOuterLimit(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"select with limit", "SELECT * FROM users LIMIT 10", true},
		{"select without limit", "SELECT * FROM users", false},
		{"table name containing limit", "SELECT * FROM rate_limit_events", false},
		{"subquery limit only", "SELECT * FROM (SELECT * FROM users LIMIT 5) sub", false},
		{"CTE limit only", "WITH cte AS (SELECT * FROM users LIMIT 1) SELECT * FROM cte", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := db.HasOuterLimit(tt.query)
			if got != tt.expected {
				t.Errorf("HasOuterLimit(%q) = %v, expected %v", tt.query, got, tt.expected)
			}
		})
	}
}
