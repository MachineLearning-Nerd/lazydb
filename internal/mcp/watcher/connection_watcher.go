package watcher

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	"github.com/MachineLearning-Nerd/lazydb/internal/storage"
	"github.com/fsnotify/fsnotify"
)

// ConnectionWatcher monitors connections.json and auto-switches database connections
type ConnectionWatcher struct {
	connMgr         *db.ConnectionManager
	preferredConn   string
	currentConn     db.Connection
	currentConnName string
	verbose         bool
	watcher         *fsnotify.Watcher
	mu              sync.RWMutex // Protects currentConn and currentConnName
}

// NewConnectionWatcher creates a new connection watcher
func NewConnectionWatcher(connMgr *db.ConnectionManager, preferredConn string, verbose bool) *ConnectionWatcher {
	return &ConnectionWatcher{
		connMgr:       connMgr,
		preferredConn: preferredConn,
		verbose:       verbose,
	}
}

// Watch monitors connections.json for changes and reconnects when needed
func (w *ConnectionWatcher) Watch(ctx context.Context) error {
	// Get connections file path
	filePath, err := storage.GetConnectionsFile()
	if err != nil {
		return fmt.Errorf("failed to get connections file path: %w", err)
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()
	w.watcher = watcher

	// Watch the connections file
	if err := watcher.Add(filePath); err != nil {
		return fmt.Errorf("failed to watch connections file: %w", err)
	}

	if w.verbose {
		fmt.Fprintf(os.Stderr, "[WATCHER] Watching %s for connection changes...\n", filePath)
	}

	// Initial load
	if err := w.ReloadConnection(ctx); err != nil {
		return fmt.Errorf("failed to load initial connection: %w", err)
	}

	// Watch for changes with fallback polling
	ticker := time.NewTicker(2 * time.Second) // Fallback polling every 2 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if w.verbose {
				fmt.Fprintf(os.Stderr, "[WATCHER] Shutting down...\n")
			}
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher events channel closed")
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				if w.verbose {
					fmt.Fprintf(os.Stderr, "[WATCHER] Detected file change (%s), reloading...\n", event.Op)
				}
				// Small delay to ensure file write is complete
				time.Sleep(100 * time.Millisecond)
				if err := w.ReloadConnection(ctx); err != nil {
					fmt.Fprintf(os.Stderr, "[WATCHER] Reload error: %v\n", err)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			fmt.Fprintf(os.Stderr, "[WATCHER] Error: %v\n", err)

		case <-ticker.C:
			// Fallback: check periodically in case file events are missed
			if err := w.ReloadConnection(ctx); err != nil {
				if w.verbose {
					fmt.Fprintf(os.Stderr, "[WATCHER] Periodic reload error: %v\n", err)
				}
			}
		}
	}
}

// ReloadConnection checks if active connection changed and reconnects if needed
func (w *ConnectionWatcher) ReloadConnection(ctx context.Context) error {
	// Load current connections from file
	savedConfig, err := storage.LoadConnections()
	if err != nil {
		return fmt.Errorf("failed to load connections: %w", err)
	}

	// Reload all connections into manager (in case new ones were added)
	for _, connConfig := range savedConfig.Connections {
		// Check if connection already exists
		if _, err := w.connMgr.GetConnection(connConfig.Name); err != nil {
			// Connection doesn't exist, add it
			conn := db.NewPostgresConnection(connConfig)
			w.connMgr.AddConnection(connConfig.Name, conn)
			if w.verbose {
				fmt.Fprintf(os.Stderr, "[WATCHER] Added new connection: %s\n", connConfig.Name)
			}
		}
	}

	// Determine which connection to use
	var targetConnName string
	if w.preferredConn != "" {
		targetConnName = w.preferredConn // Use specified connection (--connection flag)
	} else {
		targetConnName = savedConfig.ActiveConnection // Use active connection
	}

	if targetConnName == "" {
		return fmt.Errorf("no active connection specified")
	}

	// Check if connection changed
	w.mu.RLock()
	unchanged := (targetConnName == w.currentConnName && w.currentConn != nil)
	w.mu.RUnlock()

	if unchanged {
		return nil // No change, skip reconnect
	}

	// Connection changed - perform switch
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.verbose {
		if w.currentConnName != "" {
			fmt.Fprintf(os.Stderr, "[WATCHER] Switching from '%s' to '%s'...\n", w.currentConnName, targetConnName)
		} else {
			fmt.Fprintf(os.Stderr, "[WATCHER] Initial connection to '%s'...\n", targetConnName)
		}
	}

	// Disconnect old connection gracefully
	if w.currentConn != nil {
		w.currentConn.Disconnect(ctx)
	}

	// Get new connection
	conn, err := w.connMgr.GetConnection(targetConnName)
	if err != nil {
		return fmt.Errorf("connection '%s' not found: %w", targetConnName, err)
	}

	// Connect if not already connected
	if conn.Status() != db.StatusConnected {
		if err := conn.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to '%s': %w", targetConnName, err)
		}
	}

	// Update current connection
	w.currentConn = conn
	w.currentConnName = targetConnName

	if w.verbose {
		connCfg := conn.Config()
		fmt.Fprintf(os.Stderr, "[WATCHER] ✓ Now connected to: %s (database: %s)\n", connCfg.Name, connCfg.Database)
	}

	return nil
}

// GetCurrentConnection returns the current active connection
func (w *ConnectionWatcher) GetCurrentConnection() (db.Connection, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.currentConn == nil {
		return nil, fmt.Errorf("no connection available")
	}
	return w.currentConn, nil
}

// GetCurrentConnectionName returns the name of the current connection
func (w *ConnectionWatcher) GetCurrentConnectionName() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentConnName
}
