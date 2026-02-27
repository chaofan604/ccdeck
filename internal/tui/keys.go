package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Tab      key.Binding
	Enter    key.Binding
	NewSess  key.Binding
	NewGrp   key.Binding
	Delete   key.Binding
	Rename   key.Binding
	Interact key.Binding
	FullTmux key.Binding
	Quit     key.Binding
	Help     key.Binding
	Escape   key.Binding
	Yes      key.Binding
	No       key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch panel"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "attach/send"),
	),
	NewSess: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new session"),
	),
	NewGrp: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "new group"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Rename: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "rename"),
	),
	Interact: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "interact"),
	),
	FullTmux: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "full tmux"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Yes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
}

func normalHelpText() string {
	return " ↑↓ Navigate  Tab Switch Panel  ↵ Expand/Attach  i Interact  n New  g Group  d Del  r Rename  q Quit"
}

func interactHelpText() string {
	return " ⚡ LIVE MODE  All keys → Claude  │  Ctrl+Q exit"
}
