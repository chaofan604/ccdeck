package tmux

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var safeNameRe = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// SanitizeName converts a string into a valid tmux session name.
func SanitizeName(group, session string) string {
	name := fmt.Sprintf("claude_%s_%s", group, session)
	name = safeNameRe.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if len(name) > 64 {
		name = name[:64]
	}
	return name
}

// IsInstalled checks whether tmux is available on the system.
func IsInstalled() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// SessionExists checks if a tmux session with the given name is running.
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// NewSession creates a detached tmux session running claude with the given session ID.
func NewSession(name, workdir, claudeSessionID string) error {
	shellCmd := fmt.Sprintf("claude -r %s", claudeSessionID)
	args := []string{"new-session", "-d", "-s", name, "-c", workdir, shellCmd}
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux new-session failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// AttachCmd returns an exec.Cmd that attaches to the given tmux session.
func AttachCmd(name string) *exec.Cmd {
	return exec.Command("tmux", "attach-session", "-t", name)
}

// KillSession terminates a tmux session.
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux kill-session failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// ListSessions returns the names of all running tmux sessions.
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	return result, nil
}

// CapturePane captures the visible content of a tmux pane as plain text.
func CapturePane(name string, lines int) (string, error) {
	start := fmt.Sprintf("-%d", lines)
	cmd := exec.Command("tmux", "capture-pane", "-t", name, "-p", "-S", start)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("capture-pane failed: %w", err)
	}
	return string(out), nil
}

// SendText sends literal text to a tmux pane (uses -l flag, no key interpretation).
func SendText(name, text string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", name, "-l", text)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("send-keys literal failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SendSpecial sends a named special key to a tmux pane (e.g. "Enter", "BSpace", "Escape", "C-c").
func SendSpecial(name, keyName string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", name, keyName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("send-keys special failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
