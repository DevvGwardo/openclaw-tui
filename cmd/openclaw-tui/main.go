package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	flag.StringVar(&theme, "theme", "aquarium", "Color theme (ocean, amber, rose, forest, aquarium)")
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

	// Auto-read token from OpenClaw config if not provided
	if token == "" && password == "" {
		if cfgToken, cfgPort := readOpenClawConfig(); cfgToken != "" {
			token = cfgToken
			if url == "ws://127.0.0.1:18789" && cfgPort > 0 {
				url = fmt.Sprintf("ws://127.0.0.1:%d", cfgPort)
			}
		}
	}

	// Load device identity for signed auth
	var device *gateway.DeviceIdentity
	identityPath := gateway.DefaultIdentityPath()
	if identityPath != "" {
		if d, err := gateway.LoadDeviceIdentity(identityPath); err == nil {
			device = d
		}
	}

	// Create gateway client
	gw := gateway.NewClient(url, token, password, device)

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

// readOpenClawConfig reads gateway auth token and port from ~/.openclaw/openclaw.json
func readOpenClawConfig() (token string, port int) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG readConfig: homedir err=%v\n", err)
		return "", 0
	}
	cfgPath := filepath.Join(home, ".openclaw", "openclaw.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return "", 0
	}
	// Strip UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}
	var cfg struct {
		Gateway struct {
			Port int    `json:"port"`
			Auth struct {
				Token string `json:"token"`
			} `json:"auth"`
		} `json:"gateway"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", 0
	}
	return cfg.Gateway.Auth.Token, cfg.Gateway.Port
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
        --theme <NAME>       Color theme (ocean, amber, rose, forest, aquarium)
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
