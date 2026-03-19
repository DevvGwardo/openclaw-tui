package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PaletteCommand describes a command shown in the command palette.
type PaletteCommand struct {
	Name     string
	Desc     string
	Shortcut string
	HasArgs  bool
	// SubOptions are shown as a second-level picker after selection.
	SubOptions []string
}

// PaletteCommands is the canonical list of slash commands for the palette.
var PaletteCommands = []PaletteCommand{
	{Name: "help", Desc: "Show help", Shortcut: ""},
	{Name: "theme", Desc: "Switch theme", HasArgs: true, SubOptions: []string{"ocean", "amber", "rose", "forest", "aquarium"}},
	{Name: "bg", Desc: "Background animation", HasArgs: true, SubOptions: []string{"off", "starfield", "tunnel", "plasma", "fire", "matrix", "ocean", "cube", "aquarium", "skibidi", "sigma", "npc", "ohio", "rizz", "gyatt", "amogus", "bussin"}},
	{Name: "model", Desc: "Switch model", HasArgs: true, SubOptions: []string{"kimi-coding/k2p5", "minimax/MiniMax-M2.7", "kimi", "minimax"}},
	{Name: "session", Desc: "Switch session", HasArgs: true},
	{Name: "agent", Desc: "Switch agent", Shortcut: ""},
	{Name: "think", Desc: "Thinking level", HasArgs: true, SubOptions: []string{"none", "adaptive", "full"}},
	{Name: "status", Desc: "Gateway status", Shortcut: ""},
	{Name: "new", Desc: "Reset session", Shortcut: ""},
	{Name: "abort", Desc: "Abort active run", Shortcut: "Esc"},
	{Name: "clear", Desc: "Clear chat history", Shortcut: "Ctrl+L"},
	{Name: "attach", Desc: "Attach image file", HasArgs: true},
	{Name: "unattach", Desc: "Remove attachments", HasArgs: true, SubOptions: []string{"all"}},
	{Name: "feed", Desc: "Drop food in aquarium"},
	{Name: "exit", Desc: "Exit", Shortcut: "Ctrl+D"},
}

// CommandPaletteModel is a floating modal for browsing and selecting slash commands.
type CommandPaletteModel struct {
	active   bool
	filter   string // text after '/'
	selected int    // index into filtered list
	filtered []PaletteCommand
	theme    Theme

	// Sub-option picker state
	subActive  bool
	subItems   []string
	subSel     int
	parentCmd  string

	// Dynamic options populated at runtime
	agentOptions []string
}

// NewCommandPaletteModel creates a new command palette.
func NewCommandPaletteModel(theme Theme) CommandPaletteModel {
	cp := CommandPaletteModel{theme: theme}
	cp.filtered = PaletteCommands
	return cp
}

// IsActive returns whether the palette is visible.
func (cp *CommandPaletteModel) IsActive() bool {
	return cp.active
}

// Open shows the palette with optional initial filter.
func (cp *CommandPaletteModel) Open(filter string) {
	cp.active = true
	cp.subActive = false
	cp.SetFilter(filter)
}

// SetAgentOptions updates the available agents for the /agent command.
func (cp *CommandPaletteModel) SetAgentOptions(agents []string) {
	cp.agentOptions = agents
}

// Close hides the palette.
func (cp *CommandPaletteModel) Close() {
	cp.active = false
	cp.filter = ""
	cp.selected = 0
	cp.subActive = false
	cp.filtered = PaletteCommands
}

// SetFilter updates the filter text and recomputes matches.
func (cp *CommandPaletteModel) SetFilter(f string) {
	cp.filter = f
	cp.applyFilter()
	cp.selected = 0
}

// SetTheme updates the palette theme.
func (cp *CommandPaletteModel) SetTheme(t Theme) {
	cp.theme = t
}

// MoveUp moves selection up.
func (cp *CommandPaletteModel) MoveUp() {
	if cp.subActive {
		if cp.subSel > 0 {
			cp.subSel--
		}
		return
	}
	if cp.selected > 0 {
		cp.selected--
	}
}

// MoveDown moves selection down.
func (cp *CommandPaletteModel) MoveDown() {
	if cp.subActive {
		if cp.subSel < len(cp.subItems)-1 {
			cp.subSel++
		}
		return
	}
	if cp.selected < len(cp.filtered)-1 {
		cp.selected++
	}
}

// Selected returns the command string to execute, or "" if nothing selected.
// Returns the full slash command (e.g. "/theme ocean").
func (cp *CommandPaletteModel) Selected() string {
	if cp.subActive {
		if cp.subSel >= 0 && cp.subSel < len(cp.subItems) {
			return "/" + cp.parentCmd + " " + cp.subItems[cp.subSel]
		}
		return ""
	}
	if len(cp.filtered) == 0 {
		return ""
	}
	cmd := cp.filtered[cp.selected]
	// Use dynamic agent options if this is the agent command
	if cmd.Name == "agent" && len(cp.agentOptions) > 0 {
		cp.subActive = true
		cp.subItems = cp.agentOptions
		cp.subSel = 0
		cp.parentCmd = cmd.Name
		return "" // signal: don't execute yet
	}
	if len(cmd.SubOptions) > 0 {
		// Enter sub-option picker
		cp.subActive = true
		cp.subItems = cmd.SubOptions
		cp.subSel = 0
		cp.parentCmd = cmd.Name
		return "" // signal: don't execute yet
	}
	return "/" + cmd.Name
}

func (cp *CommandPaletteModel) applyFilter() {
	if cp.filter == "" {
		cp.filtered = PaletteCommands
		return
	}
	query := strings.ToLower(cp.filter)
	var results []PaletteCommand
	for _, cmd := range PaletteCommands {
		if fuzzyMatch(query, cmd.Name) || fuzzyMatch(query, strings.ToLower(cmd.Desc)) {
			results = append(results, cmd)
		}
	}
	cp.filtered = results
}

// fuzzyMatch checks if all chars of query appear in target in order.
func fuzzyMatch(query, target string) bool {
	qi := 0
	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

const (
	paletteMinWidth = 40
	paletteMaxWidth = 60
	paletteMaxItems = 12
)

// View renders the command palette overlay.
func (cp *CommandPaletteModel) View(screenWidth, screenHeight int) string {
	if !cp.active {
		return ""
	}

	p := cp.theme.Palette

	if cp.subActive {
		return cp.viewSub(screenWidth, screenHeight)
	}

	// Compute content width
	contentWidth := paletteMinWidth
	for _, cmd := range cp.filtered {
		lineLen := 4 + len(cmd.Name) + 2 + len(cmd.Desc) // "  ▸ /name  desc"
		if cmd.Shortcut != "" {
			lineLen += 2 + len(cmd.Shortcut)
		}
		if lineLen+4 > contentWidth {
			contentWidth = lineLen + 4
		}
	}
	if contentWidth > paletteMaxWidth {
		contentWidth = paletteMaxWidth
	}
	innerWidth := contentWidth - 2 // account for border

	// Search input line
	filterStyle := lipgloss.NewStyle().Foreground(p.Fg)
	searchLine := filterStyle.Render(cp.filter) + filterStyle.Render(" █")
	searchLine = padRight(searchLine, innerWidth)

	// Separator
	sep := lipgloss.NewStyle().Foreground(p.FgMuted).Render(strings.Repeat("─", innerWidth))

	// Command items
	visible := cp.filtered
	scrollOffset := 0
	if len(visible) > paletteMaxItems {
		// Keep selected item visible
		if cp.selected >= scrollOffset+paletteMaxItems {
			scrollOffset = cp.selected - paletteMaxItems + 1
		}
		if cp.selected < scrollOffset {
			scrollOffset = cp.selected
		}
		visible = visible[scrollOffset:]
		if len(visible) > paletteMaxItems {
			visible = visible[:paletteMaxItems]
		}
	}

	selectedStyle := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(p.FgMuted)
	shortcutStyle := lipgloss.NewStyle().Foreground(p.FgMuted).Italic(true)

	var lines []string
	for i, cmd := range visible {
		idx := i + scrollOffset
		prefix := "  "
		var nameStyle, descStyle lipgloss.Style
		if idx == cp.selected {
			prefix = "▸ "
			nameStyle = selectedStyle
			descStyle = selectedStyle
		} else {
			nameStyle = dimStyle
			descStyle = dimStyle
		}

		name := nameStyle.Render("/" + cmd.Name)
		desc := descStyle.Render(cmd.Desc)

		line := prefix + name
		// Pad between name and desc
		nameVisLen := 2 + 1 + len(cmd.Name) // prefix + "/" + name
		gap := 14 - (1 + len(cmd.Name))     // align descriptions
		if gap < 2 {
			gap = 2
		}
		line += strings.Repeat(" ", gap) + desc

		if cmd.Shortcut != "" {
			line += "  " + shortcutStyle.Render(cmd.Shortcut)
		}

		line = padRight(line, innerWidth)
		_ = nameVisLen
		lines = append(lines, line)
	}

	// Footer
	total := len(PaletteCommands)
	showing := len(cp.filtered)
	footer := dimStyle.Render(fmt.Sprintf(" %d of %d commands", showing, total))
	footer = padRight(footer, innerWidth)

	// Assemble content
	var content strings.Builder
	content.WriteString(searchLine + "\n")
	content.WriteString(sep + "\n")
	for _, line := range lines {
		content.WriteString(line + "\n")
	}
	content.WriteString(footer)

	// Box with border
	borderColor := p.Primary
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(p.Bg).
		Width(innerWidth).
		Padding(0, 0)

	title := lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Background(p.Bg).
		Padding(0, 1).
		Render("Commands")

	rendered := box.Render(content.String())

	// Insert title into top border
	renderedLines := strings.Split(rendered, "\n")
	if len(renderedLines) > 0 {
		topBorder := renderedLines[0]
		// Center the title in the top border
		titleWidth := lipgloss.Width(title)
		borderWidth := lipgloss.Width(topBorder)
		if titleWidth+4 < borderWidth {
			insertPos := (borderWidth - titleWidth) / 2
			// Replace chars in the top border with title
			topRunes := []rune(stripAnsi(topBorder))
			_ = topRunes
			// Simpler approach: rebuild top border
			leftLen := (innerWidth - titleWidth) / 2
			if leftLen < 1 {
				leftLen = 1
			}
			rightLen := innerWidth + 2 - titleWidth - leftLen - 2 // -2 for corners
			if rightLen < 1 {
				rightLen = 1
			}
			_ = insertPos
			borderFg := lipgloss.NewStyle().Foreground(borderColor)
			newTop := borderFg.Render("╭") +
				borderFg.Render(strings.Repeat("─", leftLen)) +
				title +
				borderFg.Render(strings.Repeat("─", rightLen)) +
				borderFg.Render("╮")
			renderedLines[0] = newTop
		}
		rendered = strings.Join(renderedLines, "\n")
	}

	// Center horizontally
	boxWidth := lipgloss.Width(renderedLines[0])
	leftPad := (screenWidth - boxWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	padding := strings.Repeat(" ", leftPad)
	paddedLines := strings.Split(rendered, "\n")
	for i, line := range paddedLines {
		paddedLines[i] = padding + line
	}

	return strings.Join(paddedLines, "\n")
}

func (cp *CommandPaletteModel) viewSub(screenWidth, screenHeight int) string {
	p := cp.theme.Palette

	innerWidth := paletteMinWidth - 2
	title := fmt.Sprintf("/%s", cp.parentCmd)

	selectedStyle := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(p.FgMuted)

	// Apply scrolling to sub-options
	visible := cp.subItems
	scrollOffset := 0
	if len(visible) > paletteMaxItems {
		if cp.subSel >= scrollOffset+paletteMaxItems {
			scrollOffset = cp.subSel - paletteMaxItems + 1
		}
		if cp.subSel < scrollOffset {
			scrollOffset = cp.subSel
		}
		visible = visible[scrollOffset:]
		if len(visible) > paletteMaxItems {
			visible = visible[:paletteMaxItems]
		}
	}

	var lines []string
	for i, opt := range visible {
		idx := i + scrollOffset
		prefix := "  "
		var style lipgloss.Style
		if idx == cp.subSel {
			prefix = "▸ "
			style = selectedStyle
		} else {
			style = dimStyle
		}
		line := prefix + style.Render(opt)
		line = padRight(line, innerWidth)
		lines = append(lines, line)
	}

	showing := len(visible)
	total := len(cp.subItems)
	var footerText string
	if showing < total {
		footerText = fmt.Sprintf(" %d of %d options", showing, total)
	} else {
		footerText = fmt.Sprintf(" %d options", total)
	}
	footer := dimStyle.Render(footerText)
	footer = padRight(footer, innerWidth)

	var content strings.Builder
	for _, line := range lines {
		content.WriteString(line + "\n")
	}
	content.WriteString(footer)

	borderColor := p.Primary
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(p.Bg).
		Width(innerWidth).
		Padding(0, 0)

	titleRendered := lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Background(p.Bg).
		Padding(0, 1).
		Render(title)

	rendered := box.Render(content.String())

	// Insert title into top border
	renderedLines := strings.Split(rendered, "\n")
	if len(renderedLines) > 0 {
		titleWidth := lipgloss.Width(titleRendered)
		leftLen := (innerWidth - titleWidth) / 2
		if leftLen < 1 {
			leftLen = 1
		}
		rightLen := innerWidth + 2 - titleWidth - leftLen - 2
		if rightLen < 1 {
			rightLen = 1
		}
		borderFg := lipgloss.NewStyle().Foreground(borderColor)
		newTop := borderFg.Render("╭") +
			borderFg.Render(strings.Repeat("─", leftLen)) +
			titleRendered +
			borderFg.Render(strings.Repeat("─", rightLen)) +
			borderFg.Render("╮")
		renderedLines[0] = newTop
		rendered = strings.Join(renderedLines, "\n")
	}

	// Center horizontally
	boxWidth := lipgloss.Width(renderedLines[0])
	leftPad := (screenWidth - boxWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	padding := strings.Repeat(" ", leftPad)
	paddedLines := strings.Split(rendered, "\n")
	for i, line := range paddedLines {
		paddedLines[i] = padding + line
	}

	return strings.Join(paddedLines, "\n")
}

// padRight pads a string with spaces to reach the desired visible width.
func padRight(s string, width int) string {
	vis := lipgloss.Width(s)
	if vis >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

