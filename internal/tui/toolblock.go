package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ToolBlockView renders a tool execution block similar to SCKelemen/tui ToolBlock.
// This provides a visual container for tool call information.
type ToolBlockView struct {
	Name    string
	Status  string // "running", "done", "failed"
	Content string
	theme   Theme
	width   int
}

// NewToolBlockView creates a new tool block.
func NewToolBlockView(name, status, content string, theme Theme, width int) ToolBlockView {
	return ToolBlockView{
		Name:    name,
		Status:  status,
		Content: content,
		theme:   theme,
		width:   width,
	}
}

// Render produces the styled tool block string.
func (t ToolBlockView) Render() string {
	var icon string
	var borderColor lipgloss.Color
	var statusText string

	switch t.Status {
	case "running":
		icon = "⏳"
		borderColor = t.theme.Palette.Warning
		statusText = "Running"
	case "done":
		icon = "✅"
		borderColor = t.theme.Palette.Success
		statusText = "Done"
	case "failed":
		icon = "❌"
		borderColor = t.theme.Palette.Error
		statusText = "Failed"
	default:
		icon = "⏳"
		borderColor = t.theme.Palette.Warning
		statusText = "Pending"
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Bold(true)

	header := headerStyle.Render(fmt.Sprintf("%s %s — %s", icon, t.Name, statusText))

	if t.Content == "" {
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Width(t.width - 4)
		return boxStyle.Render(header)
	}

	contentStyle := lipgloss.NewStyle().
		Foreground(t.theme.Palette.FgMuted).
		Width(t.width - 8)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(t.width - 4)

	return boxStyle.Render(header + "\n" + contentStyle.Render(t.Content))
}
