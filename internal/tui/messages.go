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
	glamourMu     sync.Mutex
	glamourCache  *glamour.TermRenderer
	glamourCacheW int
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
	return glamourCache
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

// RenderMessage renders a chat message with the given theme and width.
// Website-like card styling with clear visual hierarchy
func RenderMessage(msg ChatMsg, theme Theme, width int) string {
	switch msg.Role {
	case RoleUser:
		return renderUserMessage(msg, theme, width)
	case RoleAssistant:
		return renderAssistantMessage(msg, theme, width)
	case RoleSystem:
		return renderSystemMessage(msg, theme, width)
	case RoleError:
		return renderErrorMessage(msg, theme, width)
	default:
		return msg.Content
	}
}

func renderUserMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	// Card header with avatar icon and name
	avatar := lipgloss.NewStyle().
		Background(p.Secondary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Render("U")

	name := lipgloss.NewStyle().
		Foreground(p.Secondary).
		Bold(true).
		Render("You")

	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	header := fmt.Sprintf(" %s %s   %s", avatar, name, ts)

	// Content area with card styling
	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	content := lipgloss.NewStyle().
		Foreground(p.Fg).
		Width(contentWidth).
		Render(msg.Content)

	// Card wrapper with rounded border and left accent
	cardStyle := lipgloss.NewStyle().
		Background(p.UserBg).
		Foreground(p.Fg).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Secondary).
		Padding(1, 2).
		Width(width - 2)

	// Left accent bar effect
	accentBar := lipgloss.NewStyle().
		Background(p.Secondary).
		Width(3).
		Render("")

	body := lipgloss.NewStyle().
		Width(width - 5).
		Render(content)

	return "\n" + header + "\n" + cardStyle.Render(accentBar+body) + "\n"
}

func renderAssistantMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	// Card header with avatar icon and name
	avatar := lipgloss.NewStyle().
		Background(p.Primary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Render("AI")

	name := lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Render("OpenClaw")

	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	header := fmt.Sprintf(" %s %s   %s", avatar, name, ts)

	// Content area
	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	var content string
	if msg.Streaming {
		content = lipgloss.NewStyle().
			Foreground(p.Fg).
			Width(contentWidth).
			Render(msg.Content)
		content += lipgloss.NewStyle().
			Foreground(p.Primary).
			Render(" ◐")
	} else {
		content = renderGlamourMarkdown(msg.Content, contentWidth)
	}

	// Tool calls section
	if len(msg.Tools) > 0 {
		var sb strings.Builder
		sb.Grow(len(msg.Tools) * 80)
		for i, tool := range msg.Tools {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(renderToolCall(tool, theme, contentWidth))
		}
		content += "\n" + sb.String()
	}

	// Card styling
	cardStyle := lipgloss.NewStyle().
		Background(p.AssistBg).
		Foreground(p.Fg).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.AssistBorder).
		Padding(1, 2).
		Width(width - 2)

	// Left accent bar
	accentBar := lipgloss.NewStyle().
		Background(p.AssistBorder).
		Width(3).
		Render("")

	body := lipgloss.NewStyle().
		Width(width - 5).
		Render(content)

	return "\n" + header + "\n" + cardStyle.Render(accentBar+body) + "\n"
}

func renderSystemMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	// Centered, subtle divider with icon
	icon := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render("◆")

	text := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Italic(true).
		Render(msg.Content)

	textW := lipgloss.Width(text)
	sideLen := (width - textW - 6) / 2
	if sideLen < 1 {
		sideLen = 1
	}
	dash := lipgloss.NewStyle().
		Foreground(p.CardBorder).
		Render(strings.Repeat("─", sideLen))

	return "\n" + dash + " " + icon + " " + text + " " + icon + " " + dash + "\n"
}

func renderErrorMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	// Card header with error icon
	avatar := lipgloss.NewStyle().
		Background(p.Error).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Render("!")

	name := lipgloss.NewStyle().
		Foreground(p.Error).
		Bold(true).
		Render("Error")

	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	header := fmt.Sprintf(" %s %s   %s", avatar, name, ts)

	// Content
	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}
	content := lipgloss.NewStyle().
		Foreground(p.Fg).
		Width(contentWidth).
		Render(msg.Content)

	// Card with error styling
	cardStyle := lipgloss.NewStyle().
		Background(p.UserBg).
		Foreground(p.Fg).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Error).
		Padding(1, 2).
		Width(width - 2)

	accentBar := lipgloss.NewStyle().
		Background(p.Error).
		Width(3).
		Render("")

	body := lipgloss.NewStyle().
		Width(width - 5).
		Render(content)

	return "\n" + header + "\n" + cardStyle.Render(accentBar+body) + "\n"
}

func renderToolCall(tool ToolCall, theme Theme, width int) string {
	p := theme.Palette

	var icon string
	var style lipgloss.Style
	var iconBg lipgloss.Color

	switch tool.Status {
	case "running":
		icon = "▶"
		style = theme.ToolRunning
		iconBg = p.Warning
	case "done":
		icon = "✓"
		style = theme.ToolDone
		iconBg = p.Success
	case "failed":
		icon = "✗"
		style = theme.ToolFailed
		iconBg = p.Error
	default:
		icon = "▶"
		style = theme.ToolRunning
		iconBg = p.Warning
	}

	iconBadge := lipgloss.NewStyle().
		Background(iconBg).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Render(icon)

	header := lipgloss.NewStyle().
		Foreground(p.Fg).
		Render(" " + tool.Name)

	statusDot := style.Render("(" + tool.Status + ")")

	toolLine := iconBadge + header + " " + statusDot

	if tool.Output != "" {
		output := theme.Muted.Width(width - 4).Render(tool.Output)
		var sb strings.Builder
		sb.WriteString(toolLine)
		sb.WriteString("\n")
		sb.WriteString(output)
		return sb.String()
	}
	return toolLine
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
	// Glamour adds trailing newlines and blank lines between paragraphs.
	// Trim trailing newlines and collapse double blank lines to single.
	rendered = strings.TrimRight(rendered, "\n")
	for strings.Contains(rendered, "\n\n\n") {
		rendered = strings.ReplaceAll(rendered, "\n\n\n", "\n\n")
	}
	return rendered
}
