package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DevvGwardo/openclaw-tui/internal/gateway"
	"github.com/DevvGwardo/openclaw-tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.1.0"

func main() {
	var (
		url       string
		token     string
		password  string
		session   string
		message   string
		thinking  string
		theme     string
		showHelp  bool
	)

	flag.StringVar(&url, "url", "", "Gateway WebSocket URL")
	flag.StringVar(&url, "u", "", "Gateway WebSocket URL (shorthand)")
	flag.StringVar(&token, "token", "", "Auth token")
	flag.StringVar(&token, "t", "", "Auth token (shorthand)")
	flag.StringVar(&password, "password", "", "Auth password")
	flag.StringVar(&password, "p", "", "Auth password (shorthand)")
	flag.StringVar(&session, "session", "agent:main:main", "Session key")
	flag.StringVar(&session, "s", "agent:main:main", "Session key (shorthand)")
	flag.StringVar(&message, "message", "", "Send message on connect")
	flag.StringVar(&message, "m", "", "Send message on connect (shorthand)")
	flag.StringVar(&thinking, "thinking", "adaptive", "Thinking level")
	flag.StringVar(&theme, "theme", "ocean", "Color theme (ocean, amber, rose, forest)")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showHelp, "h", false, "Show help (shorthand)")

	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	// Environment variable fallbacks
	if url == "" {
		url = os.Getenv("OPENCLAW_GATEWAY_URL")
	}
	if url == "" {
		url = "ws://127.0.0.1:18789"
	}
	if token == "" {
		token = os.Getenv("OPENCLAW_TOKEN")
	}

	// Create gateway client
	gw := gateway.NewClient(url, token, password)

	// Connect
	if err := gw.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", url, err)
		fmt.Fprintf(os.Stderr, "Make sure the OpenClaw gateway is running.\n")
		os.Exit(1)
	}
	defer gw.Close()

	// Create TUI model
	model := tui.NewModel(gw, session, thinking, message, tui.ThemeName(theme))
	model.SetURL(url)

	// Run Bubble Tea
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`🦞 OpenClaw TUI v%s

A terminal chat client for OpenClaw gateway.

USAGE:
    openclaw-tui [FLAGS]

FLAGS:
    -u, --url <URL>          Gateway WebSocket URL (default: ws://127.0.0.1:18789)
    -t, --token <TOKEN>      Authentication token
    -p, --password <PASS>    Authentication password
    -s, --session <KEY>      Initial session key (default: agent:main:main)
    -m, --message <MSG>      Send a message on connect
        --thinking <LEVEL>   Thinking level (none, adaptive, full)
        --theme <NAME>       Color theme (ocean, amber, rose, forest)
    -h, --help               Show this help

ENVIRONMENT:
    OPENCLAW_GATEWAY_URL     Gateway WebSocket URL
    OPENCLAW_TOKEN           Authentication token

EXAMPLES:
    openclaw-tui
    openclaw-tui -u ws://127.0.0.1:18789
    openclaw-tui -u wss://gateway.example.com -t YOUR_TOKEN
    openclaw-tui --theme amber -m "Hello!"
`, version)
}
