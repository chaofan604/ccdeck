package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Store handles persistence of session and group data.
type Store struct {
	path string
	Data AppData
}

// NewStore creates a Store that reads/writes to ~/.config/claude-session-manager/data.json.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".config", "claude-session-manager")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create config directory: %w", err)
	}
	s := &Store{path: filepath.Join(dir, "data.json")}
	if err := s.Load(); err != nil {
		return nil, err
	}
	// Clean up empty "Default" group if other groups exist
	if len(s.Data.Groups) > 1 {
		cleaned := make([]Group, 0, len(s.Data.Groups))
		for _, g := range s.Data.Groups {
			if g.Name == "Default" && len(g.Sessions) == 0 {
				continue
			}
			cleaned = append(cleaned, g)
		}
		if len(cleaned) != len(s.Data.Groups) {
			s.Data.Groups = cleaned
			_ = s.Save()
		}
	}
	return s, nil
}

// Load reads data from disk. If the file doesn't exist, initializes empty data.
func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		s.Data = AppData{Groups: []Group{}}
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot read data file: %w", err)
	}
	return json.Unmarshal(data, &s.Data)
}

// Save writes current data to disk.
func (s *Store) Save() error {
	data, err := json.MarshalIndent(s.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal data: %w", err)
	}
	return os.WriteFile(s.path, data, 0o644)
}

// AddGroup creates a new group and returns its index.
func (s *Store) AddGroup(name string) int {
	g := Group{
		ID:        genID(),
		Name:      name,
		Sessions:  []Session{},
		CreatedAt: time.Now(),
	}
	s.Data.Groups = append(s.Data.Groups, g)
	return len(s.Data.Groups) - 1
}

// DeleteGroup removes a group by index.
func (s *Store) DeleteGroup(idx int) {
	if idx < 0 || idx >= len(s.Data.Groups) {
		return
	}
	s.Data.Groups = append(s.Data.Groups[:idx], s.Data.Groups[idx+1:]...)
}

// AddSession adds a session to a group and returns the session index.
func (s *Store) AddSession(groupIdx int, name, sessionID, path string) int {
	if groupIdx < 0 || groupIdx >= len(s.Data.Groups) {
		return -1
	}
	sess := Session{
		ID:        genID(),
		Name:      name,
		SessionID: sessionID,
		Path:      path,
		CreatedAt: time.Now(),
	}
	g := &s.Data.Groups[groupIdx]
	g.Sessions = append(g.Sessions, sess)
	return len(g.Sessions) - 1
}

// DeleteSession removes a session from a group by indices.
func (s *Store) DeleteSession(groupIdx, sessIdx int) {
	if groupIdx < 0 || groupIdx >= len(s.Data.Groups) {
		return
	}
	g := &s.Data.Groups[groupIdx]
	if sessIdx < 0 || sessIdx >= len(g.Sessions) {
		return
	}
	g.Sessions = append(g.Sessions[:sessIdx], g.Sessions[sessIdx+1:]...)
}

// Groups returns all groups.
func (s *Store) Groups() []Group {
	return s.Data.Groups
}

// Sessions returns sessions for a given group index.
func (s *Store) Sessions(groupIdx int) []Session {
	if groupIdx < 0 || groupIdx >= len(s.Data.Groups) {
		return nil
	}
	return s.Data.Groups[groupIdx].Sessions
}

func genID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
