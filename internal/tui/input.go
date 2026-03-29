package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel handles text input for chat messages.
// Website-like: clean card input with clear send button
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
	ta.Placeholder = "Type your message here..."
	ta.Prompt = "" // no prompt character inside the textarea
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetHeight(5) // taller for comfortable multiline editing
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

	ta.KeyMap.InsertNewline.SetKeys("alt+enter")

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
func (m *InputModel) InsertNewline() {
	m.textarea.InsertString("\n")
}

// InsertRune inserts a single rune at the cursor position.
func (m *InputModel) InsertRune(r rune) {
	m.textarea.InsertString(string(r))
}

// InsertString inserts a string at the cursor position.
func (m *InputModel) InsertString(s string) {
	m.textarea.InsertString(s)
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
// Website-like: card-style input with clear send indicator
func (m InputModel) View() string {
	p := m.theme.Palette

	// Attachment indicator
	var attachmentBadge string
	if m.attachmentCount > 0 {
		attachmentBadge = lipgloss.NewStyle().
			Background(p.Accent).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Render(fmt.Sprintf("📎 %d", m.attachmentCount))
	}

	// Send hint
	sendHint := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render("↵ send")

	newlineHint := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render("alt+↵ newline")

	hints := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(sendHint + "  " + newlineHint)

	// Card border style - website-like
	borderColor := p.CardBorder
	if m.focused {
		borderColor = p.Primary
	}

	cardBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		BorderBackground(p.Bg).
		Padding(0, 1).
		Width(m.width - 2)

	content := m.textarea.View()

	// Build the input section header
	inputHeader := ""
	if m.attachmentCount > 0 {
		inputHeader = attachmentBadge + "  "
	}

	// Wrap content in the card
	cardContent := inputHeader + content + "  " + hints

	return cardBorder.Render(cardContent)
}
