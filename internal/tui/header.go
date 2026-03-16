package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// HeaderModel is the top header bar showing branding and connection status.
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
func (h HeaderModel) View() string {
	title := renderGradientTitle(h.theme)
	ver := h.theme.HeaderInfo.Render(fmt.Sprintf("v%s", h.version))

	var connStatus string
	if h.connected {
		connStatus = h.theme.StatusConnected.Render("● connected")
	} else {
		connStatus = h.theme.StatusDisconnected.Render("○ disconnected")
	}

	urlDisplay := h.theme.Muted.Render(h.url)

	left := fmt.Sprintf("%s  %s  %s", title, ver, connStatus)
	right := urlDisplay

	gap := h.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	padding := ""
	for i := 0; i < gap; i++ {
		padding += " "
	}

	line := left + padding + right

	return h.theme.HeaderStyle.Width(h.width).Render(line)
}

// renderGradientTitle renders "🦞 OpenClaw" with gradient colors.
func renderGradientTitle(theme Theme) string {
	text := "OpenClaw"
	p := theme.Palette

	// Create a simple gradient from primary to secondary across the text
	colors := interpolateColors(string(p.Primary), string(p.Secondary), len(text))

	var result string
	for i, ch := range text {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(colors[i])).Bold(true)
		result += style.Render(string(ch))
	}

	return "🦞 " + result
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
