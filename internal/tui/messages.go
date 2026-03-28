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

// RenderMessage renders a chat message with the given theme and width.
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

	// ── Header: clean, minimal ──
	name := lipgloss.NewStyle().
		Foreground(p.Secondary).
		Bold(true).
		Render("You")
	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(name) - lipgloss.Width(ts) - 4
	if headerGap < 1 {
		headerGap = 1
	}
	header := name + strings.Repeat(" ", headerGap) + ts

	// ── Content ──
	contentWidth := width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}
	content := lipgloss.NewStyle().
		Foreground(p.Fg).
		Width(contentWidth).
		Render(msg.Content)

	// Wrap in a subtle card with left border accent
	body := lipgloss.NewStyle().
		Foreground(p.Fg).
		Padding(0, 2).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderForeground(p.Secondary).
		Width(width - 2).
		Render(content)

	return "\n" + header + "\n" + body + "\n"
}

func renderAssistantMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	// ── Header: clean, minimal ──
	name := lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Render("🦞 OpenClaw")
	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(name) - lipgloss.Width(ts) - 4
	if headerGap < 1 {
		headerGap = 1
	}
	header := name + strings.Repeat(" ", headerGap) + ts

	// ── Content ──
	contentWidth := width - 4
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
			Render("▌")
	} else {
		content = renderGlamourMarkdown(msg.Content, contentWidth)
	}

	// Tool calls — use strings.Builder to avoid slice allocations
	if len(msg.Tools) > 0 {
		var sb strings.Builder
		sb.Grow(len(msg.Tools) * 60)
		for i, tool := range msg.Tools {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(renderToolCall(tool, theme, contentWidth))
		}
		content += "\n" + sb.String()
	}

	body := lipgloss.NewStyle().
		Foreground(p.Fg).
		Padding(0, 2).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderForeground(p.Primary).
		Width(width - 2).
		Render(content)

	return "\n" + header + "\n" + body + "\n"
}

func renderSystemMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	// Centered, subtle divider style
	text := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Italic(true).
		Render(msg.Content)

	textW := lipgloss.Width(text)
	sideLen := (width - textW - 2) / 2
	if sideLen < 1 {
		sideLen = 1
	}
	dash := lipgloss.NewStyle().
		Foreground(p.BgSubtle).
		Render(strings.Repeat("─", sideLen))

	return "\n" + dash + " " + text + " " + dash + "\n"
}

func renderErrorMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	// ── Header ──
	name := lipgloss.NewStyle().
		Foreground(p.Error).
		Bold(true).
		Render("✗ Error")
	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(name) - lipgloss.Width(ts) - 4
	if headerGap < 1 {
		headerGap = 1
	}
	header := name + strings.Repeat(" ", headerGap) + ts

	// ── Content ──
	contentWidth := width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}
	content := lipgloss.NewStyle().
		Foreground(p.Fg).
		Width(contentWidth).
		Render(msg.Content)

	body := lipgloss.NewStyle().
		Foreground(p.Fg).
		Padding(0, 2).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderForeground(p.Error).
		Width(width - 2).
		Render(content)

	return "\n" + header + "\n" + body + "\n"
}

func renderToolCall(tool ToolCall, theme Theme, width int) string {
	var icon string
	var style lipgloss.Style

	switch tool.Status {
	case "running":
		icon = "⏳"
		style = theme.ToolRunning
	case "done":
		icon = "✓"
		style = theme.ToolDone
	case "failed":
		icon = "✗"
		style = theme.ToolFailed
	default:
		icon = "⏳"
		style = theme.ToolRunning
	}

	header := style.Render(fmt.Sprintf("%s %s", icon, tool.Name))
	if tool.Output != "" {
		output := theme.Muted.Width(width - 4).Render(tool.Output)
		var sb strings.Builder
		sb.WriteString(header)
		sb.WriteString("\n")
		sb.WriteString(output)
		return sb.String()
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
	// Glamour adds trailing newlines and blank lines between paragraphs.
	// Trim trailing newlines and collapse double blank lines to single.
	rendered = strings.TrimRight(rendered, "\n")
	for strings.Contains(rendered, "\n\n\n") {
		rendered = strings.ReplaceAll(rendered, "\n\n\n", "\n\n")
	}
	return rendered
}
