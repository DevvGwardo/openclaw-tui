package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ActivityBarModel shows a streaming/waiting indicator.
// Website-like: clean animated dots with progress feel
type ActivityBarModel struct {
	active    bool
	startTime time.Time
	frame     int
	theme     Theme
	width     int
	phrase    int
}

// Fun waiting phrases.
var waitingPhrases = []string{
	"Thinking...",
	"Processing...",
	"Analyzing...",
	"Working...",
	"Loading...",
	"Calculating...",
}

// NewActivityBarModel creates a new activity bar.
func NewActivityBarModel(theme Theme) ActivityBarModel {
	return ActivityBarModel{
		theme: theme,
	}
}

// Start begins the activity indicator.
func (a *ActivityBarModel) Start() {
	a.active = true
	a.startTime = time.Now()
	a.frame = 0
	a.phrase = 0
}

// Stop stops the activity indicator.
func (a *ActivityBarModel) Stop() {
	a.active = false
}

// IsActive returns whether the bar is active.
func (a ActivityBarModel) IsActive() bool {
	return a.active
}

// Tick advances the spinner frame.
func (a *ActivityBarModel) Tick() {
	if !a.active {
		return
	}
	a.frame = (a.frame + 1) % 4

	// Change phrase every ~5 seconds (50 ticks at 100ms)
	elapsed := time.Since(a.startTime)
	a.phrase = int(elapsed.Seconds()/5) % len(waitingPhrases)
}

// SetTheme updates the theme.
func (a *ActivityBarModel) SetTheme(t Theme) {
	a.theme = t
}

// SetWidth updates the width.
func (a *ActivityBarModel) SetWidth(w int) {
	a.width = w
}

// View renders the activity bar.
// Website-like: animated dots with typing effect
func (a ActivityBarModel) View() string {
	if !a.active {
		return ""
	}

	p := a.theme.Palette

	// Animated dots - website-like typing indicator
	dotFrames := []string{"●   ", " ●  ", "  ● ", "   ●"}
	dots := dotFrames[a.frame]

	dotsStyle := lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Render(dots)

	phrase := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(waitingPhrases[a.phrase])

	elapsed := time.Since(a.startTime).Truncate(time.Second)

	// Progress bar effect
	progress := ""
	progressLen := 10
	filled := (a.frame * progressLen) / 4
	for i := 0; i < progressLen; i++ {
		if i < filled {
			progress += lipgloss.NewStyle().Foreground(p.Primary).Render("█")
		} else {
			progress += lipgloss.NewStyle().Foreground(p.BgSubtle).Render("░")
		}
	}

	progressBar := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render("[" + progress + "] " + elapsed.String())

	// Center the content
	content := fmt.Sprintf(" %s %s  %s", dotsStyle, phrase, progressBar)

	// Wrap in a subtle bar
	barStyle := lipgloss.NewStyle().
		Background(p.BgSubtle).
		Foreground(p.Fg).
		Width(a.width - 2).
		Padding(0, 1)

	return barStyle.Render(content)
}
