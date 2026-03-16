package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// ChatModel manages the scrollable chat log.
type ChatModel struct {
	viewport viewport.Model
	messages []ChatMsg
	theme    Theme
	width    int
	height   int
}

// NewChatModel creates a new chat viewport.
func NewChatModel(theme Theme) ChatModel {
	vp := viewport.New(80, 20)
	vp.YPosition = 0

	return ChatModel{
		viewport: vp,
		theme:    theme,
	}
}

// SetSize updates the chat viewport dimensions.
func (c *ChatModel) SetSize(w, h int) {
	c.width = w
	c.height = h
	c.viewport.Width = w
	c.viewport.Height = h
	c.renderAll()
}

// SetTheme updates the theme.
func (c *ChatModel) SetTheme(t Theme) {
	c.theme = t
	c.renderAll()
}

// AddMessage appends a message to the log.
func (c *ChatModel) AddMessage(msg ChatMsg) {
	c.messages = append(c.messages, msg)
	c.renderAll()
	c.viewport.GotoBottom()
}

// UpdateLastAssistant updates the last assistant message (for streaming).
func (c *ChatModel) UpdateLastAssistant(content string, streaming bool) {
	for i := len(c.messages) - 1; i >= 0; i-- {
		if c.messages[i].Role == RoleAssistant {
			c.messages[i].Content = content
			c.messages[i].Streaming = streaming
			c.renderAll()
			c.viewport.GotoBottom()
			return
		}
	}
}

// AddToolToLastAssistant adds a tool call to the last assistant message.
func (c *ChatModel) AddToolToLastAssistant(tool ToolCall) {
	for i := len(c.messages) - 1; i >= 0; i-- {
		if c.messages[i].Role == RoleAssistant {
			c.messages[i].Tools = append(c.messages[i].Tools, tool)
			c.renderAll()
			c.viewport.GotoBottom()
			return
		}
	}
}

// Clear removes all messages.
func (c *ChatModel) Clear() {
	c.messages = nil
	c.viewport.SetContent("")
}

// MessageCount returns the number of messages.
func (c *ChatModel) MessageCount() int {
	return len(c.messages)
}

// Update handles viewport events.
func (c ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)
	return c, cmd
}

// View renders the chat viewport.
func (c ChatModel) View() string {
	return c.viewport.View()
}

func (c *ChatModel) renderAll() {
	var sb strings.Builder
	for _, msg := range c.messages {
		sb.WriteString(RenderMessage(msg, c.theme, c.width))
	}
	c.viewport.SetContent(sb.String())
}
