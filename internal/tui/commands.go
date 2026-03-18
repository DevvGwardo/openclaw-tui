package tui

import (
	"fmt"
	"strings"
)

// Command represents a parsed slash command.
type Command struct {
	Name string
	Args string
}

// ParseCommand parses a slash command from input text.
// Returns nil if the input is not a command.
func ParseCommand(input string) *Command {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return nil
	}

	parts := strings.SplitN(input[1:], " ", 2)
	cmd := &Command{Name: strings.ToLower(parts[0])}
	if len(parts) > 1 {
		cmd.Args = strings.TrimSpace(parts[1])
	}
	return cmd
}

// CommandHelp returns the help text for all commands.
func CommandHelp(theme Theme) string {
	commands := []struct {
		cmd  string
		desc string
	}{
		{"/help", "Show this help message"},
		{"/status", "Show gateway status"},
		{"/model", "Switch model"},
		{"/agent", "Switch agent"},
		{"/session", "Switch session"},
		{"/think <level>", "Set thinking level (none, adaptive, full)"},
		{"/theme <name>", "Switch theme (ocean, amber, rose, forest, aquarium)"},
		{"/bg [mode]", "Cycle/set background (off, starfield, tunnel, plasma, fire, matrix, ocean, cube, skibidi, sigma, npc, ohio, rizz, gyatt, amogus, bussin, aquarium)"},
		{"/attach <path>", "Attach an image (PNG, JPG, GIF, WEBP)"},
		{"/unattach [name]", "Remove pending attachment(s)"},
		{"/new", "Reset/create new session"},
		{"/clear", "Clear chat history"},
		{"/exit", "Exit the application"},
	}

	var sb strings.Builder
	sb.WriteString(theme.AssistPrefix.Render("🦞 Available Commands"))
	sb.WriteString("\n\n")

	for _, c := range commands {
		name := theme.InputPrompt.Render(fmt.Sprintf("  %-18s", c.cmd))
		desc := theme.Muted.Render(c.desc)
		sb.WriteString(fmt.Sprintf("%s %s\n", name, desc))
	}

	sb.WriteString("\n")
	sb.WriteString(theme.Muted.Render("  Keyboard Shortcuts:"))
	sb.WriteString("\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"Ctrl+C", "Clear input / exit (press twice)"},
		{"Ctrl+D", "Exit"},
		{"Escape", "Abort active run"},
		{"Enter", "Send message"},
		{"Shift+Enter", "New line"},
		{"Ctrl+L", "Clear chat"},
		{"Alt+M", "Toggle mouse mode (scroll ↔ full tracking)"},
		{"Shift+Click", "Select text for copy (in any mouse mode)"},
		{"PgUp / PgDown", "Scroll chat half page"},
		{"Alt+Up/Down", "Scroll chat one line"},
		{"Ctrl+Up/Down", "Scroll chat one line (alt)"},
		{"Home / End", "Jump to top / bottom of chat"},
	}

	for _, s := range shortcuts {
		key := theme.InputPrompt.Render(fmt.Sprintf("  %-18s", s.key))
		desc := theme.Muted.Render(s.desc)
		sb.WriteString(fmt.Sprintf("%s %s\n", key, desc))
	}

	return sb.String()
}
