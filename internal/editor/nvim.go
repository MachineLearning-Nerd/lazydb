package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// IsNvimAvailable checks if Neovim is installed and available
func IsNvimAvailable() bool {
	_, err := exec.LookPath("nvim")
	return err == nil
}

// OpenInNeovimCmd creates a Bubbletea command to open Neovim
// Returns a tea.Cmd that will suspend the TUI and launch Neovim
func OpenInNeovimCmd(text string) tea.Cmd {
	// Check if Neovim is available
	if !IsNvimAvailable() {
		return func() tea.Msg {
			return NvimErrorMsg{Err: fmt.Errorf("neovim is not installed or not in PATH")}
		}
	}

	// Create a temporary file
	tempFile, err := createTempFile(text)
	if err != nil {
		return func() tea.Msg {
			return NvimErrorMsg{Err: fmt.Errorf("failed to create temp file: %w", err)}
		}
	}

	// Create Neovim command
	cmd := exec.Command("nvim", tempFile)

	// Return ExecProcess which will suspend Bubbletea and run Neovim
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		// This callback runs after Neovim exits
		if err != nil {
			os.Remove(tempFile)
			return NvimErrorMsg{Err: fmt.Errorf("failed to run neovim: %w", err)}
		}

		// Read the edited content
		editedText, err := readFile(tempFile)
		os.Remove(tempFile) // Clean up

		if err != nil {
			return NvimErrorMsg{Err: fmt.Errorf("failed to read edited file: %w", err)}
		}

		return NvimSuccessMsg{Text: editedText}
	})
}

// Message types for Neovim results
type NvimErrorMsg struct {
	Err error
}

type NvimSuccessMsg struct {
	Text string
}

// createTempFile creates a temporary .sql file with the given content
func createTempFile(content string) (string, error) {
	// Create temp directory if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "lazydb")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	// Create temp file with .sql extension
	tempFile, err := os.CreateTemp(tempDir, "query_*.sql")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Write content to temp file
	if _, err := tempFile.WriteString(content); err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}


// readFile reads the content of the given file
func readFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
