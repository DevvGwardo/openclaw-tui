package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel handles text input for chat messages.
type InputModel struct {
	textarea textarea.Model
	theme    Theme
	width    int
	focused  bool
}

// NewInputModel creates a new text input.
func NewInputModel(theme Theme) InputModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message... (Enter to send, Shift+Enter for newline)"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.Focus()

	// Apply theme colors
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(theme.Palette.Primary)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(theme.Palette.FgMuted)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(theme.Palette.Fg)
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(theme.Palette.Primary)
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(theme.Palette.FgMuted)
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(theme.Palette.FgMuted)

	ta.KeyMap.InsertNewline.SetKeys("shift+enter")

	return InputModel{
		textarea: ta,
		theme:    theme,
		focused:  true,
	}
}

// SetWidth updates the input width.
func (m *InputModel) SetWidth(w int) {
	m.width = w
	m.textarea.SetWidth(w - 4)
}

// SetTheme updates theme colors.
func (m *InputModel) SetTheme(t Theme) {
	m.theme = t
	m.textarea.Cursor.Style = lipgloss.NewStyle().Foreground(t.Palette.Primary)
	m.textarea.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(t.Palette.FgMuted)
	m.textarea.FocusedStyle.Text = lipgloss.NewStyle().Foreground(t.Palette.Fg)
	m.textarea.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(t.Palette.Primary)
}

// Value returns the current input text.
func (m InputModel) Value() string {
	return m.textarea.Value()
}

// Reset clears the input.
func (m *InputModel) Reset() {
	m.textarea.Reset()
}

// Focus gives focus to the input.
func (m *InputModel) Focus() {
	m.textarea.Focus()
	m.focused = true
}

// Blur removes focus from the input.
func (m *InputModel) Blur() {
	m.textarea.Blur()
	m.focused = false
}

// Update handles input events.
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the input area.
func (m InputModel) View() string {
	prompt := m.theme.InputPrompt.Render("› ")
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Palette.Primary).
		Padding(0, 1).
		Width(m.width - 2)

	return border.Render(prompt + m.textarea.View())
}
