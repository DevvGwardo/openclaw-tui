package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarModel is the bottom status bar.
// Website-like: clean pills/badges for status info
type StatusBarModel struct {
	width      int
	session    string
	model      string
	tokens     int
	maxTokens  int
	thinking   string
	connected  bool
	mouseMode  bool
	theme      Theme
}

// NewStatusBarModel creates a new status bar.
func NewStatusBarModel(theme Theme) StatusBarModel {
	return StatusBarModel{
		theme:     theme,
		session:   "agent:main:main",
		model:     "connecting...",
		thinking:  "adaptive",
		maxTokens: 200000,
	}
}

// SetWidth updates the width.
func (s *StatusBarModel) SetWidth(w int) {
	s.width = w
}

// SetSession updates the session display.
func (s *StatusBarModel) SetSession(sess string) {
	s.session = sess
}

// SetModel updates the model display.
func (s *StatusBarModel) SetModel(model string) {
	s.model = model
}

// Model returns the current model name.
func (s StatusBarModel) Model() string {
	return s.model
}

// SetTokens updates token count.
func (s *StatusBarModel) SetTokens(current, max int) {
	s.tokens = current
	s.maxTokens = max
}

// SetThinking updates the thinking level display.
func (s *StatusBarModel) SetThinking(level string) {
	s.thinking = level
}

// SetConnected updates connection status.
func (s *StatusBarModel) SetConnected(c bool) {
	s.connected = c
	// If we're now connected but still showing "connecting...", update to "ready"
	if c && s.model == "connecting..." {
		s.model = "ready"
	}
}

// SetMouseMode updates the mouse mode indicator.
func (s *StatusBarModel) SetMouseMode(on bool) {
	s.mouseMode = on
}

// SetTheme updates the theme.
func (s *StatusBarModel) SetTheme(t Theme) {
	s.theme = t
}

// View renders the status bar.
// Website-like: horizontal pill badges separated by dividers
func (s StatusBarModel) View() string {
	p := s.theme.Palette

	// Connection status with icon
	var connIcon string
	var connBadge string
	if s.connected {
		connIcon = lipgloss.NewStyle().Foreground(p.Success).Render("●")
		connBadge = lipgloss.NewStyle().
			Background(p.Success).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Render(connIcon + " CONNECTED")
	} else {
		connIcon = lipgloss.NewStyle().Foreground(p.Error).Render("●")
		connBadge = lipgloss.NewStyle().
			Background(p.Error).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Render(connIcon + " DISCONNECTED")
	}

	// Session pill
	sessionBadge := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Foreground(p.FgMuted).
		Padding(0, 1).
		Render("SESSION: " + s.session)

	// Model pill
	modelBadge := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Foreground(p.Primary).
		Padding(0, 1).
		Render("MODEL: " + s.model)

	// Token bar with percentage
	tokenPct := 0
	if s.maxTokens > 0 {
		tokenPct = s.tokens * 100 / s.maxTokens
	}
	tokenColor := s.tokenColor()
	tokenBar := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Foreground(tokenColor).
		Padding(0, 1).
		Render(fmt.Sprintf("TOKENS: %d%%", tokenPct))

	// Thinking level pill
	thinkBadge := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Foreground(p.FgMuted).
		Padding(0, 1).
		Render("THINK: " + s.thinking)

	// Mouse mode indicator
	mouseIcon := "◎"
	mouseBadge := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Foreground(p.FgMuted).
		Padding(0, 1).
		Render("MOUSE: " + mouseIcon)

	// Separator
	sep := lipgloss.NewStyle().
		Foreground(p.CardBorder).
		Render("│")

	// Build left section
	left := fmt.Sprintf(" %s %s %s %s %s", connBadge, sep, sessionBadge, sep, modelBadge)

	// Build right section
	right := fmt.Sprintf("%s %s %s %s ", tokenBar, sep, mouseBadge, sep, thinkBadge)

	// Calculate spacing
	gap := s.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	padding := repeatStringStr(" ", gap)

	line := left + padding + right

	// Top border
	topBorder := ""
	if s.width > 2 {
		topBorder = lipgloss.NewStyle().
			Foreground(p.CardBorder).
			Render("└" + repeatStringStr("─", s.width-2) + "┘")
	}

	// Status bar content
	statusContent := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Width(s.width).
		Render(line)

	return statusContent + "\n" + topBorder
}

func (s StatusBarModel) tokenColor() lipgloss.Color {
	if s.maxTokens == 0 {
		return s.theme.Palette.FgMuted
	}
	pct := float64(s.tokens) / float64(s.maxTokens)
	switch {
	case pct >= 0.8:
		return s.theme.Palette.Error
	case pct >= 0.5:
		return s.theme.Palette.Warning
	default:
		return s.theme.Palette.Success
	}
}

func repeatStringStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
