package tui

import "github.com/charmbracelet/lipgloss"

// ThemeName identifies a color theme.
type ThemeName string

const (
	ThemeOcean  ThemeName = "ocean"
	ThemeAmber  ThemeName = "amber"
	ThemeRose   ThemeName = "rose"
	ThemeForest ThemeName = "forest"
)

// ThemeNames lists all available themes.
var ThemeNames = []ThemeName{ThemeOcean, ThemeAmber, ThemeRose, ThemeForest}

// Palette defines the colors for a theme.
type Palette struct {
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	Bg          lipgloss.Color
	BgSubtle    lipgloss.Color
	Fg          lipgloss.Color
	FgMuted     lipgloss.Color
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Error       lipgloss.Color
	UserBg      lipgloss.Color
	AssistBorder lipgloss.Color
}

var palettes = map[ThemeName]Palette{
	ThemeOcean: {
		Primary:      lipgloss.Color("#5EBED6"),
		Secondary:    lipgloss.Color("#F0A870"),
		Accent:       lipgloss.Color("#7DD3FC"),
		Bg:           lipgloss.Color("#1A1B2E"),
		BgSubtle:     lipgloss.Color("#252640"),
		Fg:           lipgloss.Color("#E8E8F0"),
		FgMuted:      lipgloss.Color("#8888A0"),
		Success:      lipgloss.Color("#6CC890"),
		Warning:      lipgloss.Color("#F6C453"),
		Error:        lipgloss.Color("#E87070"),
		UserBg:       lipgloss.Color("#1E2A3A"),
		AssistBorder: lipgloss.Color("#5EBED6"),
	},
	ThemeAmber: {
		Primary:      lipgloss.Color("#F6C453"),
		Secondary:    lipgloss.Color("#F2A65A"),
		Accent:       lipgloss.Color("#FFD580"),
		Bg:           lipgloss.Color("#1C1A16"),
		BgSubtle:     lipgloss.Color("#2A2720"),
		Fg:           lipgloss.Color("#F0E8D8"),
		FgMuted:      lipgloss.Color("#A09880"),
		Success:      lipgloss.Color("#6CC890"),
		Warning:      lipgloss.Color("#F6C453"),
		Error:        lipgloss.Color("#E87070"),
		UserBg:       lipgloss.Color("#2A2518"),
		AssistBorder: lipgloss.Color("#F6C453"),
	},
	ThemeRose: {
		Primary:      lipgloss.Color("#E87CA0"),
		Secondary:    lipgloss.Color("#C8A0E0"),
		Accent:       lipgloss.Color("#F0A0C0"),
		Bg:           lipgloss.Color("#1C1A20"),
		BgSubtle:     lipgloss.Color("#2A2530"),
		Fg:           lipgloss.Color("#F0E8F0"),
		FgMuted:      lipgloss.Color("#9888A0"),
		Success:      lipgloss.Color("#6CC890"),
		Warning:      lipgloss.Color("#F6C453"),
		Error:        lipgloss.Color("#E87070"),
		UserBg:       lipgloss.Color("#2A1825"),
		AssistBorder: lipgloss.Color("#E87CA0"),
	},
	ThemeForest: {
		Primary:      lipgloss.Color("#6CC890"),
		Secondary:    lipgloss.Color("#D8B870"),
		Accent:       lipgloss.Color("#90E0B0"),
		Bg:           lipgloss.Color("#161C18"),
		BgSubtle:     lipgloss.Color("#202A22"),
		Fg:           lipgloss.Color("#E0F0E0"),
		FgMuted:      lipgloss.Color("#80A088"),
		Success:      lipgloss.Color("#6CC890"),
		Warning:      lipgloss.Color("#F6C453"),
		Error:        lipgloss.Color("#E87070"),
		UserBg:       lipgloss.Color("#182A1E"),
		AssistBorder: lipgloss.Color("#6CC890"),
	},
}

// Theme holds the resolved lipgloss styles for a theme.
type Theme struct {
	Name    ThemeName
	Palette Palette

	// Styles
	HeaderStyle      lipgloss.Style
	HeaderTitle      lipgloss.Style
	HeaderInfo       lipgloss.Style
	StatusBarStyle   lipgloss.Style
	StatusItem       lipgloss.Style
	StatusConnected  lipgloss.Style
	StatusDisconnected lipgloss.Style
	UserPrefix       lipgloss.Style
	UserMessage      lipgloss.Style
	AssistPrefix     lipgloss.Style
	AssistBorder     lipgloss.Style
	AssistMessage    lipgloss.Style
	SystemMessage    lipgloss.Style
	ErrorMessage     lipgloss.Style
	InputStyle       lipgloss.Style
	InputPrompt      lipgloss.Style
	CodeBlock        lipgloss.Style
	ToolRunning      lipgloss.Style
	ToolDone         lipgloss.Style
	ToolFailed       lipgloss.Style
	TokenLow         lipgloss.Style
	TokenMed         lipgloss.Style
	TokenHigh        lipgloss.Style
	Muted            lipgloss.Style
}

// NewTheme creates a Theme from a theme name.
func NewTheme(name ThemeName) Theme {
	p, ok := palettes[name]
	if !ok {
		p = palettes[ThemeOcean]
		name = ThemeOcean
	}

	return Theme{
		Name:    name,
		Palette: p,

		HeaderStyle: lipgloss.NewStyle().
			Background(p.Bg).
			Foreground(p.Fg).
			Bold(true).
			Padding(0, 1),

		HeaderTitle: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),

		HeaderInfo: lipgloss.NewStyle().
			Foreground(p.FgMuted),

		StatusBarStyle: lipgloss.NewStyle().
			Background(p.BgSubtle).
			Foreground(p.Fg).
			Padding(0, 1),

		StatusItem: lipgloss.NewStyle().
			Foreground(p.FgMuted).
			Padding(0, 1),

		StatusConnected: lipgloss.NewStyle().
			Foreground(p.Success),

		StatusDisconnected: lipgloss.NewStyle().
			Foreground(p.Error),

		UserPrefix: lipgloss.NewStyle().
			Foreground(p.Secondary).
			Bold(true),

		UserMessage: lipgloss.NewStyle().
			Foreground(p.Fg),

		AssistPrefix: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),

		AssistBorder: lipgloss.NewStyle().
			Foreground(p.AssistBorder),

		AssistMessage: lipgloss.NewStyle().
			Foreground(p.Fg),

		SystemMessage: lipgloss.NewStyle().
			Foreground(p.FgMuted).
			Italic(true),

		ErrorMessage: lipgloss.NewStyle().
			Foreground(p.Error).
			Bold(true),

		InputStyle: lipgloss.NewStyle().
			Foreground(p.Fg),

		InputPrompt: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),

		CodeBlock: lipgloss.NewStyle().
			Background(p.BgSubtle).
			Foreground(p.Accent).
			Padding(0, 1),

		ToolRunning: lipgloss.NewStyle().
			Foreground(p.Warning),

		ToolDone: lipgloss.NewStyle().
			Foreground(p.Success),

		ToolFailed: lipgloss.NewStyle().
			Foreground(p.Error),

		TokenLow: lipgloss.NewStyle().
			Foreground(p.Success),

		TokenMed: lipgloss.NewStyle().
			Foreground(p.Warning),

		TokenHigh: lipgloss.NewStyle().
			Foreground(p.Error),

		Muted: lipgloss.NewStyle().
			Foreground(p.FgMuted),
	}
}
