package tui

import "github.com/charmbracelet/lipgloss"

var (
	// ── Color palette ─────────────────────────────────────────────────────
	accentColor    = lipgloss.Color("#7C3AED")
	accentDimColor = lipgloss.Color("#5B21B6")
	dimColor       = lipgloss.Color("#6B7280")
	textColor      = lipgloss.Color("#D1D5DB")
	brightColor    = lipgloss.Color("#F3F4F6")
	activeColor    = lipgloss.Color("#A78BFA")
	successColor   = lipgloss.Color("#10B981")
	warningColor   = lipgloss.Color("#F59E0B")
	dangerColor    = lipgloss.Color("#EF4444")
	infoColor      = lipgloss.Color("#3B82F6")
	surfaceColor   = lipgloss.Color("#111827")
	surface2Color  = lipgloss.Color("#1F2937")
	borderColor    = lipgloss.Color("#374151")
	borderDimColor = lipgloss.Color("#1F2937")
	highlightBg    = lipgloss.Color("#312E81")
	interactColor  = lipgloss.Color("#F59E0B")

	// ── Title bar ─────────────────────────────────────────────────────────
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(accentColor).
			Padding(0, 1)

	// ── Panel frames ──────────────────────────────────────────────────────
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderDimColor).
			Padding(0, 1)

	panelActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accentColor).
				Padding(0, 1)

	panelInteractStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(warningColor).
				Padding(0, 1)

	// ── Panel titles ──────────────────────────────────────────────────────
	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(activeColor).
			MarginBottom(1)

	panelTitleDimStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(dimColor).
				MarginBottom(1)

	// ── Tree items ────────────────────────────────────────────────────────
	groupNameStyle = lipgloss.NewStyle().
			Foreground(brightColor).
			Bold(true)

	groupCountStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	selectedItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(highlightBg).
			Bold(true)

	selectedDimStyle = lipgloss.NewStyle().
			Foreground(activeColor)

	selectArrowStyle = lipgloss.NewStyle().
				Foreground(activeColor).
				Bold(true)

	itemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			PaddingLeft(2)

	treeSessionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF"))

	treeConnectorStyle = lipgloss.NewStyle().
				Foreground(borderColor)

	treeLabelStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true)

	// ── Status indicators ─────────────────────────────────────────────────
	statusRunning = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	statusStopped = lipgloss.NewStyle().
			Foreground(dangerColor)

	statusWaiting = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	// ── Preview metadata ──────────────────────────────────────────────────
	metaNameStyle = lipgloss.NewStyle().
			Foreground(brightColor).
			Bold(true)

	metaLabelStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	metaValueStyle = lipgloss.NewStyle().
			Foreground(textColor)

	metaTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(accentColor).
			Padding(0, 1).
			Bold(true)

	metaGroupTagStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(infoColor).
				Padding(0, 1)

	metaConnectedStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	metaDisconnectedStyle = lipgloss.NewStyle().
				Foreground(dangerColor)

	metaSepStyle = lipgloss.NewStyle().
			Foreground(borderColor)

	metaIconStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	// ── Preview content ───────────────────────────────────────────────────
	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(successColor)

	previewInteractTitleStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(warningColor)

	previewTitleOfflineStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(dimColor)

	previewContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D1D5DB"))

	// ── General UI ────────────────────────────────────────────────────────
	dimStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	interactHelpStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(dangerColor).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(borderColor)

	// ── Dialog ────────────────────────────────────────────────────────────
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(1, 2).
			Width(55).
			Background(surfaceColor)

	dialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(activeColor).
				MarginBottom(1)

	dialogLabelStyle = lipgloss.NewStyle().
				Foreground(textColor).
				MarginBottom(0)

	// ── Interact badge ────────────────────────────────────────────────────
	liveTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(warningColor).
			Bold(true).
			Padding(0, 1)
)
