package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// ChatModel manages the scrollable chat log.
type ChatModel struct {
	viewport    viewport.Model
	messages    []ChatMsg
	theme       Theme
	width       int
	height      int
	renderCache []string // cached rendered string per message
	dirty       []bool   // true if message needs re-render
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
	c.invalidateAll()
	c.renderAll()
}

// SetTheme updates the theme.
func (c *ChatModel) SetTheme(t Theme) {
	c.theme = t
	c.invalidateAll()
	c.renderAll()
}

// atBottom reports whether the viewport is scrolled to (or near) the bottom.
func (c *ChatModel) atBottom() bool {
	return c.viewport.AtBottom() || c.viewport.TotalLineCount() <= c.viewport.Height
}

// AddMessage appends a message to the log.
func (c *ChatModel) AddMessage(msg ChatMsg) {
	wasAtBottom := c.atBottom()
	c.messages = append(c.messages, msg)
	c.renderCache = append(c.renderCache, "")
	c.dirty = append(c.dirty, true)
	c.renderAll()
	if wasAtBottom {
		c.viewport.GotoBottom()
	}
}

// UpdateLastAssistant updates the last assistant message (for streaming).
func (c *ChatModel) UpdateLastAssistant(content string, streaming bool) {
	wasAtBottom := c.atBottom()
	for i := len(c.messages) - 1; i >= 0; i-- {
		if c.messages[i].Role == RoleAssistant {
			c.messages[i].Content = content
			c.messages[i].Streaming = streaming
			c.dirty[i] = true
			c.renderAll()
			if wasAtBottom {
				c.viewport.GotoBottom()
			}
			return
		}
	}
}

// AddToolToLastAssistant adds a tool call to the last assistant message.
func (c *ChatModel) AddToolToLastAssistant(tool ToolCall) {
	wasAtBottom := c.atBottom()
	for i := len(c.messages) - 1; i >= 0; i-- {
		if c.messages[i].Role == RoleAssistant {
			c.messages[i].Tools = append(c.messages[i].Tools, tool)
			c.dirty[i] = true
			c.renderAll()
			if wasAtBottom {
				c.viewport.GotoBottom()
			}
			return
		}
	}
}

// Clear removes all messages.
func (c *ChatModel) Clear() {
	c.messages = nil
	c.renderCache = nil
	c.dirty = nil
	c.viewport.SetContent("")
}

// MessageCount returns the number of messages.
func (c *ChatModel) MessageCount() int {
	return len(c.messages)
}

// Height returns the viewport height.
func (c ChatModel) Height() int {
	return c.viewport.Height
}

// ScrollUp scrolls the chat up by n lines.
func (c *ChatModel) ScrollUp(n int) {
	c.viewport.LineUp(n)
}

// ScrollDown scrolls the chat down by n lines.
func (c *ChatModel) ScrollDown(n int) {
	c.viewport.LineDown(n)
}

// ScrollToTop scrolls to the top of the chat.
func (c *ChatModel) ScrollToTop() {
	c.viewport.GotoTop()
}

// ScrollToBottom scrolls to the bottom of the chat.
func (c *ChatModel) ScrollToBottom() {
	c.viewport.GotoBottom()
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

// invalidateAll marks every message as needing re-render (for theme/size changes).
func (c *ChatModel) invalidateAll() {
	for i := range c.dirty {
		c.dirty[i] = true
	}
}

func (c *ChatModel) renderAll() {
	// Re-render only dirty messages, reuse cached strings for the rest.
	for i, msg := range c.messages {
		if c.dirty[i] {
			c.renderCache[i] = RenderMessage(msg, c.theme, c.width)
			c.dirty[i] = false
		}
	}
	var sb strings.Builder
	for _, cached := range c.renderCache {
		sb.WriteString(cached)
	}
	c.viewport.SetContent(sb.String())
}
