package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

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

// RenderMessage renders a chat message with the given theme and width.
func RenderMessage(msg ChatMsg, theme Theme, width int) string {
	contentWidth := width - 4
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

	content := theme.UserMessage.Width(width).Render(msg.Content)

	return fmt.Sprintf("\n%s\n%s\n", header, content)
}

func renderAssistantMessage(msg ChatMsg, theme Theme, width int) string {
	prefix := theme.AssistPrefix.Render("🦞 Assistant")
	ts := theme.Muted.Render(msg.Timestamp.Format("15:04"))
	header := fmt.Sprintf("%s  %s", prefix, ts)

	// Render content with left border
	content := renderMarkdownSimple(msg.Content, theme, width-3)
	if msg.Streaming {
		content += theme.Muted.Render(" ▌")
	}

	// Add left border
	borderChar := theme.AssistBorder.Render("┃")
	lines := strings.Split(content, "\n")
	var bordered []string
	for _, line := range lines {
		bordered = append(bordered, borderChar+" "+line)
	}

	// Render tool calls
	var toolLines string
	for _, tool := range msg.Tools {
		toolLines += "\n" + renderToolCall(tool, theme, width-3)
	}

	return fmt.Sprintf("\n%s\n%s%s\n", header, strings.Join(bordered, "\n"), toolLines)
}

func renderSystemMessage(msg ChatMsg, theme Theme, width int) string {
	content := theme.SystemMessage.Width(width).Render(msg.Content)
	return fmt.Sprintf("\n%s\n", content)
}

func renderErrorMessage(msg ChatMsg, theme Theme, width int) string {
	prefix := theme.ErrorMessage.Render("✗ Error")
	content := theme.ErrorMessage.Width(width).Render(msg.Content)
	return fmt.Sprintf("\n%s\n%s\n", prefix, content)
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

// renderMarkdownSimple does basic markdown rendering (bold, italic, code).
func renderMarkdownSimple(text string, theme Theme, width int) string {
	// Basic code block handling
	var result strings.Builder
	lines := strings.Split(text, "\n")
	inCodeBlock := false

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				lang := strings.TrimPrefix(line, "```")
				if lang != "" {
					result.WriteString(theme.Muted.Render("  ─── " + lang + " ───"))
				} else {
					result.WriteString(theme.Muted.Render("  ──────"))
				}
				result.WriteString("\n")
			} else {
				result.WriteString(theme.Muted.Render("  ──────"))
				result.WriteString("\n")
			}
			continue
		}

		if inCodeBlock {
			result.WriteString(theme.CodeBlock.Render("  " + line))
			result.WriteString("\n")
			continue
		}

		// Inline formatting
		rendered := renderInlineMarkdown(line, theme)
		result.WriteString(rendered)
		result.WriteString("\n")
	}

	s := result.String()
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
	}
	return s
}

// renderInlineMarkdown handles bold, italic, and inline code.
func renderInlineMarkdown(line string, theme Theme) string {
	// Very simple: replace **bold** and *italic* and `code`
	// A full markdown parser would be heavier; this covers the common cases.
	line = replaceInlinePatterns(line, "**", lipgloss.NewStyle().Bold(true).Foreground(theme.Palette.Fg))
	line = replaceInlinePatterns(line, "*", lipgloss.NewStyle().Italic(true).Foreground(theme.Palette.Fg))
	line = replaceInlineCode(line, theme)
	return line
}

func replaceInlinePatterns(s, delim string, style lipgloss.Style) string {
	for {
		start := strings.Index(s, delim)
		if start == -1 {
			break
		}
		end := strings.Index(s[start+len(delim):], delim)
		if end == -1 {
			break
		}
		end += start + len(delim)
		inner := s[start+len(delim) : end]
		rendered := style.Render(inner)
		s = s[:start] + rendered + s[end+len(delim):]
	}
	return s
}

func replaceInlineCode(s string, theme Theme) string {
	for {
		start := strings.Index(s, "`")
		if start == -1 {
			break
		}
		end := strings.Index(s[start+1:], "`")
		if end == -1 {
			break
		}
		end += start + 1
		inner := s[start+1 : end]
		rendered := theme.CodeBlock.Render(inner)
		s = s[:start] + rendered + s[end+1:]
	}
	return s
}
