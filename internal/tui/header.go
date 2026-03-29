package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// HeaderModel is the top header bar showing branding and connection status.
// Website-like: clean gradient title with status badges
type HeaderModel struct {
	width     int
	connected bool
	url       string
	version   string
	theme     Theme
}

// NewHeaderModel creates a new header.
func NewHeaderModel(theme Theme, url, version string) HeaderModel {
	return HeaderModel{
		theme:   theme,
		url:     url,
		version: version,
	}
}

// SetWidth updates the header width.
func (h *HeaderModel) SetWidth(w int) {
	h.width = w
}

// SetConnected updates connection status.
func (h *HeaderModel) SetConnected(c bool) {
	h.connected = c
}

// SetTheme updates the theme.
func (h *HeaderModel) SetTheme(t Theme) {
	h.theme = t
}

// View renders the header.
// Website-like layout: [● Logo + Title] -------- [Status Badge] [URL]
func (h HeaderModel) View() string {
	p := h.theme.Palette

	// Create logo icon
	logoIcon := lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Render("◆")

	// Title with gradient effect
	title := renderGradientTitle(h.theme)

	// Version badge
	verBadge := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Foreground(p.FgMuted).
		Padding(0, 1).
		Render("v" + h.version)

	// Connection status badge - website-like pill
	var statusBadge string
	if h.connected {
		statusBadge = lipgloss.NewStyle().
			Background(p.Success).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Render("● ONLINE")
	} else {
		statusBadge = lipgloss.NewStyle().
			Background(p.Error).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Render("● OFFLINE")
	}

	// URL display
	urlDisplay := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(h.url)

	// Left section: logo + title
	left := fmt.Sprintf("%s %s  %s", logoIcon, title, verBadge)

	// Right section: status + url
	right := fmt.Sprintf("%s  %s", statusBadge, urlDisplay)

	// Calculate spacing
	gap := h.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	// Top border line - website-like divider
	topLine := ""
	if h.width > 4 {
		topLine = lipgloss.NewStyle().
			Foreground(p.CardBorder).
			Render("┌" + repeatString("─", h.width-2) + "┐")
	}

	// Main header content
	padding := repeatString(" ", gap)
	line := left + padding + right

	headerContent := lipgloss.NewStyle().
		Background(p.Bg).
		Padding(0, 1).
		Width(h.width).
		Render(line)

	// Bottom border line
	bottomLine := ""
	if h.width > 4 {
		bottomLine = lipgloss.NewStyle().
			Foreground(p.CardBorder).
			Render("├" + repeatString("─", h.width-2) + "┤")
	}

	// Build the complete header
	if bottomLine != "" {
		return fmt.Sprintf("%s\n%s\n%s", topLine, headerContent, bottomLine)
	}
	return headerContent
}

// renderGradientTitle renders "dY◆z OpenClaw" with gradient colors.
func renderGradientTitle(theme Theme) string {
	text := "OpenClaw"
	p := theme.Palette

	// Create a gradient from primary to secondary across the text
	colors := interpolateColors(string(p.Primary), string(p.Secondary), len(text))

	var result string
	for i, ch := range text {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(colors[i])).Bold(true)
		result += style.Render(string(ch))
	}

	return "dY◆z " + result
}

// interpolateColors creates a gradient of hex colors between two colors.
func interpolateColors(c1, c2 string, steps int) []string {
	if steps <= 1 {
		return []string{c1}
	}

	r1, g1, b1 := hexToRGB(c1)
	r2, g2, b2 := hexToRGB(c2)

	colors := make([]string, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		r := int(float64(r1) + t*float64(r2-r1))
		g := int(float64(g1) + t*float64(g2-g1))
		b := int(float64(b1) + t*float64(b2-b1))
		colors[i] = fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
	return colors
}

func hexToRGB(hex string) (int, int, int) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 255, 255, 255
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// repeatString creates a string of n copies of s.
func repeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
