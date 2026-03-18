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

	// ── Header row: name + timestamp ──
	name := lipgloss.NewStyle().
		Foreground(p.Secondary).
		Bold(true).
		Background(p.UserBg).
		Render("  You")
	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Background(p.UserBg).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(name) - lipgloss.Width(ts)
	if headerGap < 1 {
		headerGap = 1
	}
	headerFill := lipgloss.NewStyle().
		Background(p.UserBg).
		Render(strings.Repeat(" ", headerGap))
	header := name + headerFill + ts

	// ── Content with background ──
	contentWidth := width - 6 // 3 padding each side
	if contentWidth < 20 {
		contentWidth = 20
	}
	content := lipgloss.NewStyle().
		Foreground(p.Fg).
		Width(contentWidth).
		Render(msg.Content)

	// Wrap content in padded block with background
	body := lipgloss.NewStyle().
		Background(p.UserBg).
		Padding(0, 3).
		Width(width).
		Render(content)

	// ── Bottom edge: thin accent line ──
	accent := lipgloss.NewStyle().
		Foreground(p.Secondary).
		Render(strings.Repeat("▔", width))

	return "\n" + header + "\n" + body + "\n" + accent + "\n"
}

func renderAssistantMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	// ── Header row: name + timestamp ──
	name := lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Background(p.BgSubtle).
		Render("  🦞 OpenClaw")
	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Background(p.BgSubtle).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(name) - lipgloss.Width(ts)
	if headerGap < 1 {
		headerGap = 1
	}
	headerFill := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Render(strings.Repeat(" ", headerGap))
	header := name + headerFill + ts

	// ── Content ──
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
			Foreground(p.FgMuted).
			Render(" ▌")
	} else {
		content = renderGlamourMarkdown(msg.Content, contentWidth)
	}

	// Tool calls
	if len(msg.Tools) > 0 {
		var toolParts []string
		for _, tool := range msg.Tools {
			toolParts = append(toolParts, renderToolCall(tool, theme, contentWidth))
		}
		content += "\n" + strings.Join(toolParts, "\n")
	}

	body := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Padding(0, 3).
		Width(width).
		Render(content)

	// ── Bottom edge: thin accent line ──
	accent := lipgloss.NewStyle().
		Foreground(p.Primary).
		Render(strings.Repeat("▔", width))

	return "\n" + header + "\n" + body + "\n" + accent + "\n"
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

	// ── Header row ──
	name := lipgloss.NewStyle().
		Foreground(p.Error).
		Bold(true).
		Background(p.BgSubtle).
		Render("  ✗ Error")
	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Background(p.BgSubtle).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(name) - lipgloss.Width(ts)
	if headerGap < 1 {
		headerGap = 1
	}
	headerFill := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Render(strings.Repeat(" ", headerGap))
	header := name + headerFill + ts

	// ── Content ──
	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}
	content := lipgloss.NewStyle().
		Foreground(p.Fg).
		Width(contentWidth).
		Render(msg.Content)

	body := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Padding(0, 3).
		Width(width).
		Render(content)

	accent := lipgloss.NewStyle().
		Foreground(p.Error).
		Render(strings.Repeat("▔", width))

	return "\n" + header + "\n" + body + "\n" + accent + "\n"
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
	// Glamour adds trailing newlines and blank lines between paragraphs.
	// Trim trailing newlines and collapse double blank lines to single.
	rendered = strings.TrimRight(rendered, "\n")
	for strings.Contains(rendered, "\n\n\n") {
		rendered = strings.ReplaceAll(rendered, "\n\n\n", "\n\n")
	}
	return rendered
}
