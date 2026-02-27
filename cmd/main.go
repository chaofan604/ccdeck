package main

import (
	"fmt"
	"os"

	"claude-session-manager/internal/model"
	"claude-session-manager/internal/tmux"
	"claude-session-manager/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

// These are populated at build time via -ldflags.
// Example:
//   go build -ldflags "-X 'main.Version=v0.1.0' -X 'main.Commit=abc1234'" -o claude-sm ./cmd/
var (
	Version = "dev"
	Commit  = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("claude-sm %s (%s)\n", Version, Commit)
			return
		}
	}

	if !tmux.IsInstalled() {
		fmt.Fprintln(os.Stderr, "Error: tmux is not installed. Please install tmux first.")
		os.Exit(1)
	}

	store, err := model.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	app := tui.New(store)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
