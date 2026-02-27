package model

import "time"

// Session represents a Claude Code session with its project context.
type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SessionID string    `json:"session_id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// Group organizes related sessions together.
type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Sessions  []Session `json:"sessions"`
	CreatedAt time.Time `json:"created_at"`
}

// AppData is the top-level data structure persisted to disk.
type AppData struct {
	Groups []Group `json:"groups"`
}
