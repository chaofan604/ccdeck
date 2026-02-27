# Claude Session Manager

A terminal UI tool for managing multiple [Claude Code](https://docs.anthropic.com/en/docs/claude-code) sessions. Organize sessions into groups, launch them in tmux, preview output in real-time, and interact with Claude directly from the TUI.

## Demo

[![asciicast](https://asciinema.org/a/RQvC5K40LImbfNZG.svg)](https://asciinema.org/a/RQvC5K40LImbfNZG)

## Features

- **Session Groups** — Organize Claude sessions into named groups (e.g. by project, team, or task)
- **Tree View** — Left panel shows all groups and sessions in a collapsible tree with live status indicators
- **Real-time Preview** — Right panel displays live tmux output from the selected session
- **LIVE Mode** — Type directly into the TUI and have keystrokes forwarded to Claude in real-time (press `i`)
- **Full Tmux Attach** — Jump into the full tmux session for unrestricted terminal access (press `Enter` on preview)
- **Auto Recovery** — Session metadata persists to disk. After a reboot, sessions are automatically recreated when you open them
- **Rich Metadata** — View session name, status, project path, session ID, creation time, and tags at a glance

## Prerequisites

- **Go 1.21+**
- **tmux** installed and available in `$PATH`
- **Claude Code CLI** (`claude`) installed

## Install

```bash
git clone <repo-url> && cd agent_manger_tui
# simple build
go build -o claude-sm ./cmd/

# build with version info
go build -ldflags "-X 'main.Version=v0.1.0' -X 'main.Commit=$(git rev-parse --short HEAD)'" -o claude-sm ./cmd/
```

Move the binary to your PATH:

```bash
mv claude-sm /usr/local/bin/
```

## Usage

```bash
claude-sm

# show version
claude-sm --version
# or
claude-sm -v
```

### Quick Start

1. Press `g` to create a group (e.g. "work")
2. Press `n` to add a session — provide:
   - **Project path**: the working directory (e.g. `~/projects/my-app`)
   - **Session ID**: Claude session ID or rename (from `claude --resume`)
   - **Display name** (optional): a short label for the TUI
3. Navigate to the session and press `Enter` to launch it in tmux
4. Press `Tab` to switch to the preview panel, then `i` for LIVE mode or `Enter` for full tmux

### Keyboard Shortcuts

#### Normal Mode

| Key | Action |
|---|---|
| `↑` / `k` | Navigate up in tree |
| `↓` / `j` | Navigate down in tree |
| `Tab` | Switch focus between left (tree) and right (preview) panel |
| `Enter` | Tree: expand/collapse group. Preview: attach to full tmux session |
| `i` | Enter LIVE interactive mode (keystrokes forwarded to Claude) |
| `g` | Create a new group |
| `n` | Create a new session in the current group |
| `d` | Delete selected group or session |
| `r` | Rename selected group or session |
| `q` / `Ctrl+C` | Quit |

#### LIVE Mode

| Key | Action |
|---|---|
| All keys | Forwarded to the Claude tmux session |
| `Ctrl+Q` | Exit LIVE mode, return to normal |

#### Dialogs

| Key | Action |
|---|---|
| `Tab` | Next input field |
| `Shift+Tab` | Previous input field |
| `Enter` | Confirm |
| `Esc` | Cancel |

## Layout

```
┌─────────────────────────────────────────────────────────┐
│ ◆  Claude Session Manager                               │
├──────────────────┬──────────────────────────────────────┤
│ ☰ SESSIONS       │ ◎ PREVIEW                            │
│                  │                                      │
│ 1.▾ work (3) ●2  │  my-api  ● connected                │
│   ├─ █ my-api    │  📁 ~/projects/my-api                │
│   ├─ █ frontend  │  ⏰ 2 hours ago                      │
│   └─ × backend   │  claude  work                        │
│ 2.▸ personal (1) │  ──────────────────                   │
│                  │  Status:  ● Connected                │
│                  │  Session: abc123                      │
│                  │  ──────────────────                   │
│                  │  > Claude output here...              │
│                  │                                      │
├──────────────────┴──────────────────────────────────────┤
│ ↑↓ Navigate  Tab Switch Panel  ↵ Expand/Attach  q Quit │
└─────────────────────────────────────────────────────────┘
```

## Data Storage

Session and group metadata is stored in:

```
~/.config/claude-session-manager/data.json
```

This file persists across reboots. Tmux sessions are ephemeral — when you select a session after a reboot, the tool automatically creates a new tmux session and runs `claude -r <session_id>` to restore the Claude conversation.

## Project Structure

```
.
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── model/
│   │   ├── types.go          # Session, Group, AppData structs
│   │   └── store.go          # JSON persistence
│   ├── tmux/
│   │   └── tmux.go           # tmux command wrappers
│   └── tui/
│       ├── app.go            # Main TUI model, update, view
│       ├── keys.go           # Key bindings
│       └── styles.go         # lipgloss styles
├── go.mod
└── go.sum
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [bubbles](https://github.com/charmbracelet/bubbles) — TUI components (text input)

## License

MIT
