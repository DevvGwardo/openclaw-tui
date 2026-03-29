package tui

import "github.com/charmbracelet/lipgloss"

// ThemeName identifies a color theme.
type ThemeName string

const (
	ThemeOcean    ThemeName = "ocean"
	ThemeAmber    ThemeName = "amber"
	ThemeRose     ThemeName = "rose"
	ThemeForest   ThemeName = "forest"
	ThemeAquarium ThemeName = "aquarium"
	ThemeWebsite  ThemeName = "website"
)

// ThemeNames lists all available themes.
var ThemeNames = []ThemeName{ThemeOcean, ThemeAmber, ThemeRose, ThemeForest, ThemeAquarium, ThemeWebsite}

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
	AssistBg    lipgloss.Color
	AssistBorder lipgloss.Color
	CardBg      lipgloss.Color
	CardBorder  lipgloss.Color
	NavBg       lipgloss.Color
	NavFg       lipgloss.Color
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
		AssistBg:     lipgloss.Color("#1E2A3A"),
		AssistBorder: lipgloss.Color("#5EBED6"),
		CardBg:       lipgloss.Color("#1E2A3A"),
		CardBorder:   lipgloss.Color("#3A4A5A"),
		NavBg:        lipgloss.Color("#252640"),
		NavFg:        lipgloss.Color("#8888A0"),
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
		AssistBg:     lipgloss.Color("#2A2518"),
		AssistBorder: lipgloss.Color("#F6C453"),
		CardBg:       lipgloss.Color("#2A2518"),
		CardBorder:   lipgloss.Color("#4A4030"),
		NavBg:        lipgloss.Color("#2A2720"),
		NavFg:        lipgloss.Color("#A09880"),
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
		AssistBg:     lipgloss.Color("#2A1825"),
		AssistBorder: lipgloss.Color("#E87CA0"),
		CardBg:       lipgloss.Color("#2A1825"),
		CardBorder:   lipgloss.Color("#4A3040"),
		NavBg:        lipgloss.Color("#2A2530"),
		NavFg:        lipgloss.Color("#9888A0"),
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
		AssistBg:     lipgloss.Color("#182A1E"),
		AssistBorder: lipgloss.Color("#6CC890"),
		CardBg:       lipgloss.Color("#182A1E"),
		CardBorder:   lipgloss.Color("#304030"),
		NavBg:        lipgloss.Color("#202A22"),
		NavFg:        lipgloss.Color("#80A088"),
	},
	ThemeAquarium: {
		Primary:      lipgloss.Color("#00B4D8"),
		Secondary:    lipgloss.Color("#F4A261"),
		Accent:       lipgloss.Color("#48CAE4"),
		Bg:           lipgloss.Color("#0A1628"),
		BgSubtle:     lipgloss.Color("#0F2035"),
		Fg:           lipgloss.Color("#CAF0F8"),
		FgMuted:      lipgloss.Color("#5E8BA0"),
		Success:      lipgloss.Color("#2EC4B6"),
		Warning:      lipgloss.Color("#F4A261"),
		Error:        lipgloss.Color("#E76F51"),
		UserBg:       lipgloss.Color("#0D1E30"),
		AssistBg:     lipgloss.Color("#0D1E30"),
		AssistBorder: lipgloss.Color("#00B4D8"),
		CardBg:       lipgloss.Color("#0D1E30"),
		CardBorder:   lipgloss.Color("#1A3A50"),
		NavBg:        lipgloss.Color("#0F2035"),
		NavFg:        lipgloss.Color("#5E8BA0"),
	},
	// Website theme: modern, clean, with card-based UI
	ThemeWebsite: {
		Primary:      lipgloss.Color("#3B82F6"),
		Secondary:    lipgloss.Color("#8B5CF6"),
		Accent:       lipgloss.Color("#06B6D4"),
		Bg:           lipgloss.Color("#0F172A"),
		BgSubtle:     lipgloss.Color("#1E293B"),
		Fg:           lipgloss.Color("#F8FAFC"),
		FgMuted:      lipgloss.Color("#94A3B8"),
		Success:      lipgloss.Color("#10B981"),
		Warning:      lipgloss.Color("#F59E0B"),
		Error:        lipgloss.Color("#EF4444"),
		UserBg:       lipgloss.Color("#1E3A5F"),
		AssistBg:     lipgloss.Color("#1E293B"),
		AssistBorder: lipgloss.Color("#3B82F6"),
		CardBg:       lipgloss.Color("#1E293B"),
		CardBorder:   lipgloss.Color("#334155"),
		NavBg:        lipgloss.Color("#1E293B"),
		NavFg:        lipgloss.Color("#94A3B8"),
	},
}

// Theme holds the resolved lipgloss styles for a theme.
type Theme struct {
	Name    ThemeName
	Palette Palette

	// Styles - Website-like UI
	HeaderStyle         lipgloss.Style
	HeaderTitle        lipgloss.Style
	HeaderInfo         lipgloss.Style
	NavStyle           lipgloss.Style
	NavItem            lipgloss.Style
	NavItemActive      lipgloss.Style
	StatusBarStyle     lipgloss.Style
	StatusItem         lipgloss.Style
	StatusConnected    lipgloss.Style
	StatusDisconnected lipgloss.Style

	// Message styles - Card-based
	UserCardStyle       lipgloss.Style
	AssistCardStyle    lipgloss.Style
	UserPrefix         lipgloss.Style
	UserMessage        lipgloss.Style
	AssistPrefix       lipgloss.Style
	AssistBorder       lipgloss.Style
	AssistMessage      lipgloss.Style
	SystemMessage      lipgloss.Style
	ErrorMessage       lipgloss.Style

	// Input styles
	InputStyle          lipgloss.Style
	InputPrompt         lipgloss.Style
	CodeBlock           lipgloss.Style

	// Tool styles
	ToolRunning         lipgloss.Style
	ToolDone            lipgloss.Style
	ToolFailed          lipgloss.Style

	// Token styles
	TokenLow            lipgloss.Style
	TokenMed            lipgloss.Style
	TokenHigh           lipgloss.Style

	// Misc
	Muted               lipgloss.Style
	CardStyle           lipgloss.Style
}

// NewTheme creates a Theme from a theme name.
func NewTheme(name ThemeName) Theme {
	p, ok := palettes[name]
	if !ok {
		p = palettes[ThemeOcean]
		name = ThemeOcean
	}

	// Website-like: card-based messages with clear visual hierarchy
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
			Bold(true).
			Padding(0, 1),

		HeaderInfo: lipgloss.NewStyle().
			Foreground(p.FgMuted),

		// Navigation bar style (website-like top bar)
		NavStyle: lipgloss.NewStyle().
			Background(p.NavBg).
			Foreground(p.NavFg).
			Padding(0, 2),

		NavItem: lipgloss.NewStyle().
			Foreground(p.NavFg).
			Padding(0, 1),

		NavItemActive: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true).
			Padding(0, 1),

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

		// Card-based message styles (website-like)
		UserCardStyle: lipgloss.NewStyle().
			Background(p.UserBg).
			Foreground(p.Fg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Secondary).
			Padding(1, 2).
			Margin(0, 0),

		AssistCardStyle: lipgloss.NewStyle().
			Background(p.AssistBg).
			Foreground(p.Fg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.AssistBorder).
			Padding(1, 2).
			Margin(0, 0),

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
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.CardBorder).
			Padding(1, 2),

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

		CardStyle: lipgloss.NewStyle().
			Background(p.CardBg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.CardBorder).
			Padding(1, 2),
	}
}
