package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// glamourRenderer caches a glamour renderer instance per word-wrap width.
var (
	glamourMu       sync.Mutex
	glamourCache    *glamour.TermRenderer
	glamourCacheW   int
)

// getGlamourRenderer returns a cached glamour renderer for the given width.
func getGlamourRenderer(width int) *glamour.TermRenderer {
	glamourMu.Lock()
	defer glamourMu.Unlock()
	if glamourCache != nil && glamourCacheW == width {
		return glamourCache
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	glamourCache = r
	glamourCacheW = width
	return r
}

// MsgRole identifies who sent a message.
type MsgRole string

const (
	RoleUser      MsgRole = "user"
	RoleAssistant MsgRole = "assistant"
	RoleSystem    MsgRole = "system"
	RoleError     MsgRole = "error"
)

// ChatMsg represents a single message in the chat log.
type ChatMsg struct {
	Role      MsgRole
	Content   string
	Timestamp time.Time
	Streaming bool
	RunID     string
	Tools     []ToolCall
}

// ToolCall represents a tool execution within an assistant message.
type ToolCall struct {
	Name   string
	Status string // "running", "done", "failed"
	Output string
}

// msgIndent is the left indent applied to message content for visual hierarchy.
const msgIndent = "  "

// RenderMessage renders a chat message with the given theme and width.
func RenderMessage(msg ChatMsg, theme Theme, width int) string {
	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	switch msg.Role {
	case RoleUser:
		return renderUserMessage(msg, theme, contentWidth)
	case RoleAssistant:
		return renderAssistantMessage(msg, theme, contentWidth)
	case RoleSystem:
		return renderSystemMessage(msg, theme, contentWidth)
	case RoleError:
		return renderErrorMessage(msg, theme, contentWidth)
	default:
		return msg.Content
	}
}

func renderUserMessage(msg ChatMsg, theme Theme, width int) string {
	prefix := theme.UserPrefix.Render("› You")
	ts := theme.Muted.Render(msg.Timestamp.Format("15:04"))
	header := fmt.Sprintf("%s  %s", prefix, ts)

	// Indent content lines
	content := theme.UserMessage.Render(msg.Content)
	content = indentBlock(content, msgIndent)

	return fmt.Sprintf("\n%s\n%s\n", header, content)
}

func renderAssistantMessage(msg ChatMsg, theme Theme, width int) string {
	prefix := theme.AssistPrefix.Render("🦞 Assistant")
	ts := theme.Muted.Render(msg.Timestamp.Format("15:04"))
	header := fmt.Sprintf("%s  %s", prefix, ts)

	// Render content with glamour markdown
	content := renderGlamourMarkdown(msg.Content, width-5)
	if msg.Streaming {
		content += theme.Muted.Render(" ▌")
	}

	// Add left border with indent
	borderChar := theme.AssistBorder.Render("┃")
	lines := strings.Split(content, "\n")
	var bordered []string
	for _, line := range lines {
		bordered = append(bordered, msgIndent+borderChar+" "+line)
	}

	// Render tool calls
	var toolLines string
	for _, tool := range msg.Tools {
		toolLines += "\n" + msgIndent + renderToolCall(tool, theme, width-5)
	}

	return fmt.Sprintf("\n%s\n%s%s\n", header, strings.Join(bordered, "\n"), toolLines)
}

func renderSystemMessage(msg ChatMsg, theme Theme, width int) string {
	content := theme.SystemMessage.Render(msg.Content)
	content = indentBlock(content, msgIndent)
	return fmt.Sprintf("\n%s\n", content)
}

func renderErrorMessage(msg ChatMsg, theme Theme, width int) string {
	prefix := theme.ErrorMessage.Render("✗ Error")
	content := theme.ErrorMessage.Render(msg.Content)
	content = indentBlock(content, msgIndent)
	return fmt.Sprintf("\n%s\n%s\n", prefix, content)
}

// indentBlock prepends indent to each line in the block.
func indentBlock(s, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

func renderToolCall(tool ToolCall, theme Theme, width int) string {
	var icon string
	var style lipgloss.Style

	switch tool.Status {
	case "running":
		icon = "⏳"
		style = theme.ToolRunning
	case "done":
		icon = "✅"
		style = theme.ToolDone
	case "failed":
		icon = "❌"
		style = theme.ToolFailed
	default:
		icon = "⏳"
		style = theme.ToolRunning
	}

	header := style.Render(fmt.Sprintf("  %s %s", icon, tool.Name))
	if tool.Output != "" {
		output := theme.Muted.Width(width - 4).Render(tool.Output)
		return header + "\n" + output
	}
	return header
}

// renderGlamourMarkdown renders markdown content using the glamour library.
// Falls back to plain text if glamour fails.
func renderGlamourMarkdown(text string, width int) string {
	if width < 10 {
		width = 10
	}
	r := getGlamourRenderer(width)
	if r == nil {
		return text
	}
	rendered, err := r.Render(text)
	if err != nil {
		return text
	}
	// Glamour adds trailing newlines; trim them for consistent formatting.
	rendered = strings.TrimRight(rendered, "\n")
	return rendered
}
