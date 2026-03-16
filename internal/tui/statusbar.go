package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarModel is the bottom status bar.
type StatusBarModel struct {
	width      int
	session    string
	model      string
	tokens     int
	maxTokens  int
	thinking   string
	connected  bool
	theme      Theme
}

// NewStatusBarModel creates a new status bar.
func NewStatusBarModel(theme Theme) StatusBarModel {
	return StatusBarModel{
		theme:     theme,
		session:   "agent:main:main",
		model:     "unknown",
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
}

// SetTheme updates the theme.
func (s *StatusBarModel) SetTheme(t Theme) {
	s.theme = t
}

// View renders the status bar.
func (s StatusBarModel) View() string {
	var connIcon string
	if s.connected {
		connIcon = s.theme.StatusConnected.Render("●")
	} else {
		connIcon = s.theme.StatusDisconnected.Render("○")
	}

	sessionItem := s.theme.StatusItem.Render(fmt.Sprintf("📋 %s", s.session))
	modelItem := s.theme.StatusItem.Render(fmt.Sprintf("🤖 %s", s.model))

	tokenStyle := s.tokenStyle()
	tokenPct := 0
	if s.maxTokens > 0 {
		tokenPct = s.tokens * 100 / s.maxTokens
	}
	tokenItem := tokenStyle.Render(fmt.Sprintf("🔢 %dk/%dk (%d%%)", s.tokens/1000, s.maxTokens/1000, tokenPct))

	thinkItem := s.theme.StatusItem.Render(fmt.Sprintf("💭 %s", s.thinking))

	left := fmt.Sprintf(" %s %s %s %s", connIcon, sessionItem, modelItem, tokenItem)
	right := fmt.Sprintf("%s ", thinkItem)

	gap := s.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	padding := ""
	for i := 0; i < gap; i++ {
		padding += " "
	}

	line := left + padding + right

	return s.theme.StatusBarStyle.Width(s.width).Render(line)
}

func (s StatusBarModel) tokenStyle() lipgloss.Style {
	if s.maxTokens == 0 {
		return s.theme.TokenLow
	}
	pct := float64(s.tokens) / float64(s.maxTokens)
	switch {
	case pct >= 0.8:
		return s.theme.TokenHigh
	case pct >= 0.5:
		return s.theme.TokenMed
	default:
		return s.theme.TokenLow
	}
}
