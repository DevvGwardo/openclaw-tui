package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel handles text input for chat messages.
type InputModel struct {
	textarea        textarea.Model
	theme           Theme
	width           int
	focused         bool
	attachmentCount int
}

// NewInputModel creates a new text input.
func NewInputModel(theme Theme) InputModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message... (Enter to send, Shift+Enter for newline)"
	ta.Prompt = "" // no prompt character inside the textarea
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.Focus()

	// Apply theme colors
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(theme.Palette.Primary)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(theme.Palette.FgMuted)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(theme.Palette.Fg)
	ta.FocusedStyle.Prompt = lipgloss.NewStyle()
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(theme.Palette.FgMuted)
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(theme.Palette.FgMuted)
	ta.BlurredStyle.Prompt = lipgloss.NewStyle()

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
	m.textarea.FocusedStyle.Prompt = lipgloss.NewStyle()
}

// Value returns the current input text.
func (m InputModel) Value() string {
	return m.textarea.Value()
}

// Reset clears the input.
func (m *InputModel) Reset() {
	m.textarea.Reset()
}

// InsertNewline inserts a newline character at the cursor position.
// Used during bracketed paste to preserve multiline content.
func (m *InputModel) InsertNewline() {
	m.textarea.InsertString("\n")
}

// InsertRune inserts a single rune at the cursor position.
func (m *InputModel) InsertRune(r rune) {
	m.textarea.InsertString(string(r))
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

// SetAttachmentCount updates the number of pending attachments to display.
func (m *InputModel) SetAttachmentCount(n int) {
	m.attachmentCount = n
}

// Update handles input events.
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the input area.
func (m InputModel) View() string {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Palette.Primary).
		Padding(0, 1).
		Width(m.width - 2)

	content := m.textarea.View()

	// Show attachment indicator above the textarea
	if m.attachmentCount > 0 {
		badge := lipgloss.NewStyle().
			Foreground(m.theme.Palette.Accent).
			Bold(true).
			Render(fmt.Sprintf(" %d image(s) attached", m.attachmentCount))
		content = badge + "\n" + content
	}

	return border.Render(content)
}
