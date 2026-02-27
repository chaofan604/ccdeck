package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"claude-session-manager/internal/model"
	"claude-session-manager/internal/tmux"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type focusPanel int

const (
	panelTree focusPanel = iota
	panelPreview
)

type dialogMode int

const (
	dialogNone dialogMode = iota
	dialogNewGroup
	dialogNewSession
	dialogDeleteConfirm
	dialogRename
)

type tmuxExitMsg struct{ err error }

type refreshMsg struct {
	sessions map[string]bool
	content  string
}

type sendDoneMsg struct{ err error }

// treePos represents a position in the tree: group header or session
type treePos struct {
	groupIdx   int
	sessionIdx int // -1 = group header
}

var tmuxSpecialKeys = map[string]string{
	"enter": "Enter", "backspace": "BSpace", "tab": "Tab",
	"shift+tab": "BTab", "esc": "Escape",
	"up": "Up", "down": "Down", "left": "Left", "right": "Right",
	"delete": "DC", "home": "Home", "end": "End",
	"pgup": "PPage", "pgdown": "NPage", "insert": "IC",
	"f1": "F1", "f2": "F2", "f3": "F3", "f4": "F4",
	"f5": "F5", "f6": "F6", "f7": "F7", "f8": "F8",
	"f9": "F9", "f10": "F10", "f11": "F11", "f12": "F12",
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

// Model is the main TUI model.
type Model struct {
	store *model.Store

	// Panel focus
	focus focusPanel

	// Tree navigation
	groupIdx   int
	sessionIdx int // -1 = cursor on group header, >=0 = cursor on session
	expanded   map[int]bool

	// Dialog
	dialog       dialogMode
	inputs       []textinput.Model
	inputIdx     int
	deleteTarget string

	// Interact mode
	interactMode   bool
	previewContent string

	width  int
	height int

	statusMsg    string
	err          error
	tmuxSessions map[string]bool
}

// New creates a new TUI Model.
func New(store *model.Store) Model {
	exp := make(map[int]bool)
	for i := range store.Groups() {
		exp[i] = true
	}
	return Model{
		store:        store,
		sessionIdx:   -1,
		expanded:     exp,
		tmuxSessions: make(map[string]bool),
	}
}

// onGroupHeader returns true if cursor is on a group header (not a session).
func (m Model) onGroupHeader() bool {
	return m.sessionIdx < 0
}

// ---------------------------------------------------------------------------
// Init / tick / refresh
// ---------------------------------------------------------------------------

func (m Model) Init() tea.Cmd {
	return m.scheduleRefresh()
}

func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(_ time.Time) tea.Msg {
		return m.doRefresh()
	})
}

func (m Model) doRefresh() tea.Msg {
	sessions, _ := tmux.ListSessions()
	result := make(map[string]bool)
	for _, s := range sessions {
		result[s] = true
	}
	content := ""
	tn := m.selectedTmuxName()
	if tn != "" && result[tn] {
		captured, err := tmux.CapturePane(tn, 200)
		if err == nil {
			content = captured
		}
	}
	return refreshMsg{sessions: result, content: content}
}

func (m Model) selectedTmuxName() string {
	groups := m.store.Groups()
	if m.groupIdx >= len(groups) || m.sessionIdx < 0 {
		return ""
	}
	ss := groups[m.groupIdx].Sessions
	if m.sessionIdx >= len(ss) {
		return ""
	}
	return tmux.SanitizeName(groups[m.groupIdx].Name, ss[m.sessionIdx].Name)
}

// ---------------------------------------------------------------------------
// Tree navigation helpers
// ---------------------------------------------------------------------------

func (m Model) buildTree() []treePos {
	var tree []treePos
	for gi, g := range m.store.Groups() {
		tree = append(tree, treePos{gi, -1})
		if m.expanded[gi] {
			for si := range g.Sessions {
				tree = append(tree, treePos{gi, si})
			}
		}
	}
	return tree
}

func (m Model) currentTreeIdx() int {
	tree := m.buildTree()
	for i, p := range tree {
		if p.groupIdx == m.groupIdx && p.sessionIdx == m.sessionIdx {
			return i
		}
	}
	return 0
}

func (m *Model) moveTree(delta int) {
	tree := m.buildTree()
	if len(tree) == 0 {
		return
	}
	cur := m.currentTreeIdx()
	next := cur + delta
	if next < 0 {
		next = 0
	}
	if next >= len(tree) {
		next = len(tree) - 1
	}
	m.groupIdx = tree[next].groupIdx
	m.sessionIdx = tree[next].sessionIdx
}

func (m Model) activeCountForGroup(gi int) int {
	groups := m.store.Groups()
	if gi >= len(groups) {
		return 0
	}
	count := 0
	for _, s := range groups[gi].Sessions {
		tn := tmux.SanitizeName(groups[gi].Name, s.Name)
		if m.tmuxSessions[tn] {
			count++
		}
	}
	return count
}

// ---------------------------------------------------------------------------
// Update dispatcher
// ---------------------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case refreshMsg:
		m.tmuxSessions = msg.sessions
		m.previewContent = msg.content
		return m, m.scheduleRefresh()

	case sendDoneMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Send failed: %v", msg.err)
		}
		return m, nil

	case tmuxExitMsg:
		m.statusMsg = "Returned from tmux session"
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("tmux exited with error: %v", msg.err)
		}
		return m, nil

	case tea.KeyMsg:
		if m.dialog != dialogNone {
			return m.updateDialog(msg)
		}
		if m.interactMode {
			return m.updateInteract(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Normal mode
// ---------------------------------------------------------------------------

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Tab):
		if m.focus == panelTree {
			m.focus = panelPreview
		} else {
			m.focus = panelTree
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.focus == panelTree {
			m.moveTree(-1)
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.focus == panelTree {
			m.moveTree(1)
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		if m.focus == panelTree {
			if m.onGroupHeader() {
				m.expanded[m.groupIdx] = !m.expanded[m.groupIdx]
				return m, nil
			}
			// On a session in tree → select it and switch to preview
			m.focus = panelPreview
			return m, nil
		}
		// On preview panel → attach tmux
		if !m.onGroupHeader() {
			return m.attachSession()
		}
		return m, nil

	case key.Matches(msg, keys.Interact):
		if m.onGroupHeader() {
			return m, nil
		}
		// Allow i from either panel
		tn := m.selectedTmuxName()
		if tn == "" || !m.tmuxSessions[tn] {
			m.statusMsg = "Session not running. Press Enter on tree to start."
			return m, nil
		}
		m.focus = panelPreview
		m.interactMode = true
		m.statusMsg = ""
		return m, nil

	case key.Matches(msg, keys.NewGrp):
		m.dialog = dialogNewGroup
		m.inputs = []textinput.Model{newInput("Group name", "e.g. Work", 30)}
		m.inputIdx = 0
		m.inputs[0].Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.NewSess):
		if len(m.store.Groups()) == 0 {
			m.statusMsg = "Create a group first (press g)"
			return m, nil
		}
		m.dialog = dialogNewSession
		m.inputs = []textinput.Model{
			newInput("Project path", "~/projects/my-app", 60),
			newInput("Session ID / Name", "session id or rename", 60),
			newInput("Display name (optional)", "e.g. api-refactor", 30),
		}
		m.inputIdx = 0
		m.inputs[0].Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.Delete):
		if m.focus != panelTree || len(m.store.Groups()) == 0 {
			return m, nil
		}
		if m.onGroupHeader() {
			m.dialog = dialogDeleteConfirm
			m.deleteTarget = "group"
		} else {
			m.dialog = dialogDeleteConfirm
			m.deleteTarget = "session"
		}
		return m, nil

	case key.Matches(msg, keys.Rename):
		if m.focus != panelTree || len(m.store.Groups()) == 0 {
			return m, nil
		}
		if m.onGroupHeader() {
			m.dialog = dialogRename
			m.deleteTarget = "group"
			g := m.store.Groups()[m.groupIdx]
			m.inputs = []textinput.Model{newInput("New name", g.Name, 30)}
			m.inputs[0].SetValue(g.Name)
			m.inputs[0].Focus()
			return m, textinput.Blink
		}
		m.dialog = dialogRename
		m.deleteTarget = "session"
		s := m.store.Sessions(m.groupIdx)[m.sessionIdx]
		m.inputs = []textinput.Model{newInput("New name", s.Name, 30)}
		m.inputs[0].SetValue(s.Name)
		m.inputs[0].Focus()
		return m, textinput.Blink
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// Interact mode
// ---------------------------------------------------------------------------

func (m Model) updateInteract(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tn := m.selectedTmuxName()
	if tn == "" || !m.tmuxSessions[tn] {
		m.interactMode = false
		m.statusMsg = "Session ended"
		return m, nil
	}

	keyStr := msg.String()

	if keyStr == "ctrl+q" {
		m.interactMode = false
		m.statusMsg = "Exited interact mode"
		return m, nil
	}

	if tmuxKey, ok := tmuxSpecialKeys[keyStr]; ok {
		return m, m.sendSpecialCmd(tn, tmuxKey)
	}
	if strings.HasPrefix(keyStr, "ctrl+") {
		return m, m.sendSpecialCmd(tn, "C-"+strings.TrimPrefix(keyStr, "ctrl+"))
	}
	if strings.HasPrefix(keyStr, "alt+") {
		return m, m.sendSpecialCmd(tn, "M-"+strings.TrimPrefix(keyStr, "alt+"))
	}
	if msg.Type == tea.KeyRunes {
		return m, m.sendTextCmd(tn, string(msg.Runes))
	}
	if msg.Type == tea.KeySpace {
		return m, m.sendTextCmd(tn, " ")
	}
	return m, nil
}

func (m Model) sendTextCmd(tn, text string) tea.Cmd {
	return func() tea.Msg {
		return sendDoneMsg{err: tmux.SendText(tn, text)}
	}
}

func (m Model) sendSpecialCmd(tn, keyName string) tea.Cmd {
	return func() tea.Msg {
		return sendDoneMsg{err: tmux.SendSpecial(tn, keyName)}
	}
}

// ---------------------------------------------------------------------------
// Dialog handling
// ---------------------------------------------------------------------------

func (m Model) updateDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.dialog = dialogNone
		m.inputs = nil
		return m, nil
	case msg.Type == tea.KeyTab:
		if len(m.inputs) > 1 {
			m.inputs[m.inputIdx].Blur()
			m.inputIdx = (m.inputIdx + 1) % len(m.inputs)
			m.inputs[m.inputIdx].Focus()
			return m, textinput.Blink
		}
		return m, nil
	case msg.Type == tea.KeyShiftTab:
		if len(m.inputs) > 1 {
			m.inputs[m.inputIdx].Blur()
			m.inputIdx = (m.inputIdx - 1 + len(m.inputs)) % len(m.inputs)
			m.inputs[m.inputIdx].Focus()
			return m, textinput.Blink
		}
		return m, nil
	case msg.Type == tea.KeyEnter:
		return m.submitDialog()
	}

	if m.dialog == dialogDeleteConfirm {
		if key.Matches(msg, keys.Yes) {
			return m.confirmDelete()
		}
		if msg.String() == "n" || key.Matches(msg, keys.Escape) {
			m.dialog = dialogNone
			return m, nil
		}
		return m, nil
	}

	var cmd tea.Cmd
	if m.inputIdx < len(m.inputs) {
		m.inputs[m.inputIdx], cmd = m.inputs[m.inputIdx].Update(msg)
	}
	return m, cmd
}

func (m Model) submitDialog() (tea.Model, tea.Cmd) {
	switch m.dialog {
	case dialogNewGroup:
		name := strings.TrimSpace(m.inputs[0].Value())
		if name == "" {
			m.statusMsg = "Group name cannot be empty"
			return m, nil
		}
		idx := m.store.AddGroup(name)
		if err := m.store.Save(); err != nil {
			m.err = err
		}
		m.groupIdx = idx
		m.sessionIdx = -1
		m.expanded[idx] = true
		m.statusMsg = fmt.Sprintf("Created group: %s", name)

	case dialogNewSession:
		path := strings.TrimSpace(m.inputs[0].Value())
		sessionID := strings.TrimSpace(m.inputs[1].Value())
		displayName := strings.TrimSpace(m.inputs[2].Value())
		if path == "" || sessionID == "" {
			m.statusMsg = "Path and Session ID are required"
			return m, nil
		}
		if displayName == "" {
			displayName = sessionID
		}
		idx := m.store.AddSession(m.groupIdx, displayName, sessionID, path)
		if err := m.store.Save(); err != nil {
			m.err = err
		}
		m.expanded[m.groupIdx] = true
		m.sessionIdx = idx
		m.statusMsg = fmt.Sprintf("Created session: %s", displayName)

	case dialogRename:
		name := strings.TrimSpace(m.inputs[0].Value())
		if name == "" {
			m.statusMsg = "Name cannot be empty"
			return m, nil
		}
		if m.deleteTarget == "group" {
			m.store.Data.Groups[m.groupIdx].Name = name
		} else {
			m.store.Data.Groups[m.groupIdx].Sessions[m.sessionIdx].Name = name
		}
		if err := m.store.Save(); err != nil {
			m.err = err
		}
		m.statusMsg = fmt.Sprintf("Renamed to: %s", name)
	}

	m.dialog = dialogNone
	m.inputs = nil
	return m, nil
}

func (m Model) confirmDelete() (tea.Model, tea.Cmd) {
	if m.deleteTarget == "group" {
		name := m.store.Groups()[m.groupIdx].Name
		m.store.DeleteGroup(m.groupIdx)
		if len(m.store.Groups()) == 0 {
			m.groupIdx = 0
			m.sessionIdx = -1
		} else if m.groupIdx >= len(m.store.Groups()) {
			m.groupIdx = len(m.store.Groups()) - 1
		}
		m.sessionIdx = -1
		m.statusMsg = fmt.Sprintf("Deleted group: %s", name)
	} else {
		sessions := m.store.Sessions(m.groupIdx)
		if m.sessionIdx < len(sessions) {
			name := sessions[m.sessionIdx].Name
			m.store.DeleteSession(m.groupIdx, m.sessionIdx)
			remaining := len(m.store.Sessions(m.groupIdx))
			if remaining == 0 {
				m.sessionIdx = -1
			} else if m.sessionIdx >= remaining {
				m.sessionIdx = remaining - 1
			}
			m.statusMsg = fmt.Sprintf("Deleted session: %s", name)
		}
	}
	_ = m.store.Save()
	m.dialog = dialogNone
	return m, nil
}

// ---------------------------------------------------------------------------
// Attach to tmux
// ---------------------------------------------------------------------------

func (m Model) attachSession() (tea.Model, tea.Cmd) {
	sessions := m.store.Sessions(m.groupIdx)
	if m.sessionIdx < 0 || m.sessionIdx >= len(sessions) {
		return m, nil
	}
	sess := sessions[m.sessionIdx]
	group := m.store.Groups()[m.groupIdx]
	tmuxName := tmux.SanitizeName(group.Name, sess.Name)

	if !tmux.SessionExists(tmuxName) {
		path := expandPath(sess.Path)
		if err := tmux.NewSession(tmuxName, path, sess.SessionID); err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
			m.err = err
			return m, nil
		}
	}

	cmd := tmux.AttachCmd(tmuxName)
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return tmuxExitMsg{err: err}
	})
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// ── Header bar ────────────────────────────────────────────────────────
	logoIcon := "◆"
	headerText := fmt.Sprintf(" %s  Claude Session Manager", logoIcon)
	header := titleStyle.Width(m.width).Render(headerText)

	contentHeight := m.height - 4
	if contentHeight < 8 {
		contentHeight = 8
	}

	leftWidth := m.width / 3
	if leftWidth < 30 {
		leftWidth = 30
	}
	if leftWidth > 50 {
		leftWidth = 50
	}
	rightWidth := m.width - leftWidth - 4

	leftPanel := m.renderTreePanel(leftWidth, contentHeight)
	rightPanel := m.renderPreviewPanel(rightWidth, contentHeight)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// ── Footer ────────────────────────────────────────────────────────────
	var helpLine string
	if m.interactMode {
		helpLine = interactHelpStyle.Render(interactHelpText())
	} else {
		helpLine = helpStyle.Render(normalHelpText())
	}
	statusLine := ""
	if m.err != nil {
		statusLine = errorStyle.Render("✗ " + m.err.Error())
	} else if m.statusMsg != "" {
		statusLine = statusBarStyle.Render("• " + m.statusMsg)
	}
	footer := helpLine
	if statusLine != "" {
		footer = helpLine + "  " + statusLine
	}

	page := lipgloss.JoinVertical(lipgloss.Left, header, content, footer)

	if m.dialog != dialogNone {
		overlay := m.renderDialog()
		page = m.overlayDialog(page, overlay)
	}
	return page
}

// ---------------------------------------------------------------------------
// Tree panel (left) — groups + sessions inline
// ---------------------------------------------------------------------------

func (m Model) renderTreePanel(width, height int) string {
	groups := m.store.Groups()
	var lines []string

	// ── Panel title ───────────────────────────────────────────────────────
	titleIcon := "☰"
	if m.focus == panelTree {
		lines = append(lines, panelTitleStyle.Render(fmt.Sprintf(" %s SESSIONS", titleIcon)))
	} else {
		lines = append(lines, panelTitleDimStyle.Render(fmt.Sprintf(" %s SESSIONS", titleIcon)))
	}

	if len(groups) == 0 {
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("  No groups yet."))
		lines = append(lines, dimStyle.Render("  Press 'g' to create one."))
	}

	for gi, g := range groups {
		// ── Group header ──────────────────────────────────────────────────
		expandIcon := "▾"
		if !m.expanded[gi] {
			expandIcon = "▸"
		}

		activeCount := m.activeCountForGroup(gi)
		total := len(g.Sessions)

		name := groupNameStyle.Render(g.Name)
		countPart := groupCountStyle.Render(fmt.Sprintf("(%d)", total))
		activePart := ""
		if activeCount > 0 {
			activePart = " " + statusRunning.Render(fmt.Sprintf("● %d", activeCount))
		}
		groupLine := fmt.Sprintf(" %d.%s %s %s%s", gi+1, expandIcon, name, countPart, activePart)

		isSelected := m.groupIdx == gi && m.sessionIdx < 0
		if isSelected && m.focus == panelTree && !m.interactMode {
			lines = append(lines, selectedItemStyle.Width(width-2).Render(" ›"+groupLine))
		} else if isSelected {
			lines = append(lines, selectedDimStyle.Render(" ›"+groupLine))
		} else {
			lines = append(lines, itemStyle.Render(groupLine))
		}

		// ── Sessions under group ──────────────────────────────────────────
		if !m.expanded[gi] {
			continue
		}
		for si, s := range g.Sessions {
			tn := tmux.SanitizeName(g.Name, s.Name)
			isRunning := m.tmuxSessions[tn]

			connector := "├─"
			if si == len(g.Sessions)-1 {
				connector = "└─"
			}
			connectorStr := treeConnectorStyle.Render(connector)

			sessName := truncate(s.Name, width-16)
			suffix := treeLabelStyle.Render(" claude")

			isSessSelected := m.groupIdx == gi && m.sessionIdx == si
			var statusDot string
			if isSessSelected {
				statusDot = statusRunning.Render("●")
			} else if isRunning {
				statusDot = dimStyle.Render("●")
			} else {
				statusDot = statusStopped.Render("×")
			}

			body := fmt.Sprintf(" %s %s %s%s", connectorStr, statusDot, sessName, suffix)

			if isSessSelected && m.focus == panelTree && !m.interactMode {
				lines = append(lines, selectedItemStyle.Width(width-2).Render("  "+body))
			} else if isSessSelected {
				lines = append(lines, selectedDimStyle.Render("  "+body))
			} else {
				lines = append(lines, treeSessionStyle.Render("  "+body))
			}
		}
	}

	listContent := strings.Join(lines, "\n")
	usedRows := len(lines)
	pad := height - 2 - usedRows
	if pad > 0 {
		listContent += strings.Repeat("\n", pad)
	}

	style := panelStyle.Width(width).Height(height)
	if m.focus == panelTree && !m.interactMode {
		style = panelActiveStyle.Width(width).Height(height)
	}
	return style.Render(listContent)
}

// ---------------------------------------------------------------------------
// Preview panel (right) — metadata + live content
// ---------------------------------------------------------------------------

func (m Model) renderPreviewPanel(width, height int) string {
	body := m.renderPreviewContent(width-2, height-2)

	style := panelStyle.Width(width).Height(height)
	if m.interactMode {
		style = panelInteractStyle.Width(width).Height(height)
	} else if m.focus == panelPreview {
		style = panelActiveStyle.Width(width).Height(height)
	}
	return style.Render(body)
}

func (m Model) renderPreviewContent(width, maxRows int) string {
	groups := m.store.Groups()

	// ── Panel title ───────────────────────────────────────────────────────
	var titleLine string
	if m.interactMode {
		titleLine = previewInteractTitleStyle.Render(" ⚡ LIVE") + "  " + liveTagStyle.Render("INTERACTIVE")
	} else if m.focus == panelPreview {
		titleLine = panelTitleStyle.Render(" ◎ PREVIEW")
	} else {
		titleLine = panelTitleDimStyle.Render(" ◎ PREVIEW")
	}

	if len(groups) == 0 || m.groupIdx >= len(groups) {
		return padHeight(titleLine+"\n\n"+dimStyle.Render("  No group selected"), maxRows)
	}

	if m.onGroupHeader() {
		return m.renderGroupSummary(titleLine, width, maxRows)
	}

	sessions := m.store.Sessions(m.groupIdx)
	if m.sessionIdx >= len(sessions) {
		return padHeight(titleLine+"\n\n"+dimStyle.Render("  No session selected"), maxRows)
	}

	sess := sessions[m.sessionIdx]
	group := groups[m.groupIdx]
	tn := m.selectedTmuxName()
	isRunning := m.tmuxSessions[tn]

	metaBlock := m.renderMetaHeader(sess, group, tn, isRunning, width)
	header := titleLine + "\n" + metaBlock
	headerHeight := strings.Count(header, "\n") + 1

	if !isRunning {
		hint := "\n  " + dimStyle.Render("▶ Press Enter to launch tmux session")
		hint += "\n  " + dimStyle.Render("  Then press i to interact in-place")
		body := header + "\n" + hint
		return padHeight(body, maxRows)
	}

	content := m.previewContent
	if content == "" {
		content = dimStyle.Render("  ⏳ Waiting for output...")
	}

	contentLines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	availableRows := maxRows - headerHeight - 1
	if availableRows < 3 {
		availableRows = 3
	}

	if len(contentLines) > availableRows {
		moreCount := len(contentLines) - availableRows
		contentLines = contentLines[len(contentLines)-availableRows:]
		contentLines[0] = dimStyle.Render(fmt.Sprintf("  ↑ %d more lines above", moreCount))
	}

	for i, line := range contentLines {
		if lipgloss.Width(line) > width {
			runes := []rune(line)
			if len(runes) > width-1 {
				contentLines[i] = string(runes[:width-1]) + "…"
			}
		}
	}

	displayContent := previewContentStyle.Render(strings.Join(contentLines, "\n"))
	renderedLines := len(contentLines)
	if pad := availableRows - renderedLines; pad > 0 {
		displayContent += strings.Repeat("\n", pad)
	}

	sep := metaSepStyle.Render(strings.Repeat("─", width))
	return header + "\n" + sep + "\n" + displayContent
}

func (m Model) renderGroupSummary(titleLine string, width, maxRows int) string {
	group := m.store.Groups()[m.groupIdx]
	activeCount := m.activeCountForGroup(m.groupIdx)

	name := metaNameStyle.Render(group.Name)
	var statusBadge string
	if activeCount > 0 {
		statusBadge = statusRunning.Render(fmt.Sprintf("● %d active", activeCount))
	} else {
		statusBadge = dimStyle.Render("○ idle")
	}

	line1 := " " + name + "  " + statusBadge
	line2 := "  " + metaIconStyle.Render("📦") + " " + metaLabelStyle.Render("Sessions ") + metaValueStyle.Render(fmt.Sprintf("%d total", len(group.Sessions)))
	line3 := "  " + metaIconStyle.Render("🕐") + " " + metaLabelStyle.Render("Created  ") + metaValueStyle.Render(group.CreatedAt.Format("2006-01-02 15:04")) +
		dimStyle.Render("  ("+timeAgo(group.CreatedAt)+")")
	line4 := "  " + metaTagStyle.Render("claude") + " " + metaGroupTagStyle.Render(group.Name)

	sep := metaSepStyle.Render(strings.Repeat("─", width))
	body := titleLine + "\n" + line1 + "\n" + line2 + "\n" + line3 + "\n" + line4 + "\n" + sep

	if len(group.Sessions) > 0 {
		body += "\n\n" + dimStyle.Render("  ↑↓ Navigate sessions • Enter: start • i: interact")
		body += "\n" + dimStyle.Render("  n: add session • d: delete • r: rename")
	} else {
		body += "\n\n" + dimStyle.Render("  No sessions yet.")
		body += "\n" + dimStyle.Render("  Press 'n' to add a session.")
	}

	return padHeight(body, maxRows)
}

func (m Model) renderMetaHeader(sess model.Session, group model.Group, tn string, isRunning bool, width int) string {
	// ── Line 1: Name + status badge ───────────────────────────────────────
	name := metaNameStyle.Render(sess.Name)
	var statusBadge string
	if m.interactMode {
		statusBadge = statusWaiting.Render("● interactive")
	} else if isRunning {
		statusBadge = statusRunning.Render("● connected")
	} else {
		statusBadge = metaDisconnectedStyle.Render("○ stopped")
	}
	line1 := " " + name + "  " + statusBadge

	// ── Line 2: Path ──────────────────────────────────────────────────────
	pathDisplay := sess.Path
	if ep := expandPath(pathDisplay); ep != pathDisplay {
		pathDisplay = ep
	}
	line2 := "  " + metaIconStyle.Render("📁") + " " + metaValueStyle.Render(truncate(pathDisplay, width-8))

	// ── Line 3: Time ──────────────────────────────────────────────────────
	line3 := "  " + metaIconStyle.Render("⏰") + " " + metaValueStyle.Render(timeAgo(sess.CreatedAt))

	// ── Line 4: Tags ──────────────────────────────────────────────────────
	line4 := "  " + metaTagStyle.Render("claude") + " " + metaGroupTagStyle.Render(group.Name)

	// ── Line 5+6: Status details ──────────────────────────────────────────
	sep := metaSepStyle.Render(strings.Repeat("─", width))
	var connLabel string
	if isRunning {
		connLabel = metaConnectedStyle.Render("● Connected")
	} else {
		connLabel = metaDisconnectedStyle.Render("○ Disconnected")
	}
	line5 := metaLabelStyle.Render("  Status:  ") + connLabel
	line6 := metaLabelStyle.Render("  Session: ") + metaValueStyle.Render(sess.SessionID)

	return line1 + "\n" + line2 + "\n" + line3 + "\n" + line4 + "\n" + sep + "\n" + line5 + "\n" + line6
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// ---------------------------------------------------------------------------
// Dialogs
// ---------------------------------------------------------------------------

func (m Model) renderDialog() string {
	switch m.dialog {
	case dialogNewGroup:
		title := dialogTitleStyle.Render("✦ New Group")
		label := dialogLabelStyle.Render("Group Name:")
		input := m.inputs[0].View()
		hint := dimStyle.Render("↵ confirm  esc cancel")
		return dialogStyle.Render(fmt.Sprintf("%s\n\n%s\n%s\n\n%s", title, label, input, hint))

	case dialogNewSession:
		title := dialogTitleStyle.Render("✦ New Session")
		var fields []string
		labels := []string{"📁 Project Path:", "🔑 Session ID / Rename:", "📝 Display Name (optional):"}
		for i, l := range labels {
			fields = append(fields, dialogLabelStyle.Render(l)+"\n"+m.inputs[i].View())
		}
		hint := dimStyle.Render("tab next field  ↵ confirm  esc cancel")
		return dialogStyle.Render(title + "\n\n" + strings.Join(fields, "\n\n") + "\n\n" + hint)

	case dialogDeleteConfirm:
		title := dialogTitleStyle.Render("⚠ Confirm Delete")
		name := ""
		if m.deleteTarget == "group" && m.groupIdx < len(m.store.Groups()) {
			name = m.store.Groups()[m.groupIdx].Name
		} else if m.deleteTarget == "session" {
			ss := m.store.Sessions(m.groupIdx)
			if m.sessionIdx >= 0 && m.sessionIdx < len(ss) {
				name = ss[m.sessionIdx].Name
			}
		}
		msg := metaValueStyle.Render(fmt.Sprintf("Delete %s ", m.deleteTarget)) +
			metaNameStyle.Render(fmt.Sprintf("'%s'", name)) +
			metaValueStyle.Render(" ?")
		hint := dimStyle.Render("y yes  n/esc no")
		return dialogStyle.Render(fmt.Sprintf("%s\n\n%s\n\n%s", title, msg, hint))

	case dialogRename:
		title := dialogTitleStyle.Render(fmt.Sprintf("✎ Rename %s", m.deleteTarget))
		label := dialogLabelStyle.Render("New name:")
		input := m.inputs[0].View()
		hint := dimStyle.Render("↵ confirm  esc cancel")
		return dialogStyle.Render(fmt.Sprintf("%s\n\n%s\n%s\n\n%s", title, label, input, hint))
	}
	return ""
}

func (m Model) overlayDialog(background, dialog string) string {
	bgLines := strings.Split(background, "\n")
	dlgLines := strings.Split(dialog, "\n")
	dlgWidth := lipgloss.Width(dialog)
	dlgHeight := len(dlgLines)
	startY := max((m.height-dlgHeight)/2, 0)
	startX := max((m.width-dlgWidth)/2, 0)

	for len(bgLines) < m.height {
		bgLines = append(bgLines, strings.Repeat(" ", m.width))
	}
	for i, dlgLine := range dlgLines {
		row := startY + i
		if row >= len(bgLines) {
			break
		}
		bgRunes := []rune(bgLines[row])
		for len(bgRunes) < m.width {
			bgRunes = append(bgRunes, ' ')
		}
		prefix := string(bgRunes[:startX])
		suffix := ""
		end := startX + lipgloss.Width(dlgLine)
		if end < len(bgRunes) {
			suffix = string(bgRunes[end:])
		}
		bgLines[row] = prefix + dlgLine + suffix
	}
	return strings.Join(bgLines, "\n")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newInput(placeholder, hint string, width int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = hint
	ti.Width = width
	ti.CharLimit = 256
	ti.PromptStyle = lipgloss.NewStyle().Foreground(accentColor)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E0E0FF"))
	return ti
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return home + path[1:]
		}
	}
	return path
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func padHeight(content string, targetRows int) string {
	lines := strings.Count(content, "\n") + 1
	if lines < targetRows {
		content += strings.Repeat("\n", targetRows-lines)
	}
	return content
}
