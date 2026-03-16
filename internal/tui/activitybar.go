package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Braille spinner frames.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Fun waiting phrases.
var waitingPhrases = []string{
	"Thinking deeply...",
	"Crafting a response...",
	"Processing your request...",
	"Analyzing the problem...",
	"Gathering thoughts...",
	"Working on it...",
	"Almost there...",
	"Crunching the details...",
}

// ActivityBarModel shows a streaming/waiting indicator.
type ActivityBarModel struct {
	active    bool
	startTime time.Time
	frame     int
	theme     Theme
	width     int
	phrase    int
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
	a.frame = (a.frame + 1) % len(spinnerFrames)

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
func (a ActivityBarModel) View() string {
	if !a.active {
		return ""
	}

	spinner := lipgloss.NewStyle().
		Foreground(a.theme.Palette.Primary).
		Bold(true).
		Render(spinnerFrames[a.frame])

	phrase := lipgloss.NewStyle().
		Foreground(a.theme.Palette.FgMuted).
		Italic(true).
		Render(waitingPhrases[a.phrase])

	elapsed := time.Since(a.startTime).Truncate(time.Second)
	timer := lipgloss.NewStyle().
		Foreground(a.theme.Palette.FgMuted).
		Render(fmt.Sprintf("(%s)", elapsed))

	return fmt.Sprintf("  %s %s %s", spinner, phrase, timer)
}
