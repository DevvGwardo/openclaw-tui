package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DevvGwardo/openclaw-tui/internal/gateway"
	tea "github.com/charmbracelet/bubbletea"
)

// Msg types for Bubble Tea.
type (
	// GatewayEventMsg wraps a gateway event for the TUI.
	GatewayEventMsg gateway.GatewayEvent

	// TickMsg is the spinner/activity tick.
	TickMsg time.Time

	// ConnectedMsg signals successful connection.
	ConnectedMsg struct{}

	// SendResultMsg is the result of sending a chat message.
	SendResultMsg struct {
		Err error
	}

	// StatusResultMsg is the result of a /status request.
	StatusResultMsg struct {
		Content string
		Err     error
	}

	// ModelInfoMsg carries the model name from a silent status fetch.
	ModelInfoMsg struct {
		Model string
	}
)

// Model is the main Bubble Tea model.
type Model struct {
	gateway     *gateway.Client
	sessionKey  string
	thinking    string
	initMessage string

	// UI components
	header         HeaderModel
	chat           ChatModel
	input          InputModel
	statusBar      StatusBarModel
	activityBar    ActivityBarModel
	background     BackgroundModel
	commandPalette CommandPaletteModel
	theme          Theme

	// State
	width       int
	height      int
	connected   bool
	streaming   bool
	streamBuf   string
	activeRunID string
	ctrlCCount  int
	lastCtrlC   time.Time
	quitting    bool
	err         error
}

// NewModel creates the main TUI model.
func NewModel(gw *gateway.Client, sessionKey, thinking, initMessage string, themeName ThemeName) Model {
	theme := NewTheme(themeName)

	return Model{
		gateway:     gw,
		sessionKey:  sessionKey,
		thinking:    thinking,
		initMessage: initMessage,
		theme:       theme,
		header:      NewHeaderModel(theme, "", "0.1.0"),
		chat:        NewChatModel(theme),
		input:       NewInputModel(theme),
		statusBar:   NewStatusBarModel(theme),
		activityBar: NewActivityBarModel(theme),
		background:     NewBackgroundModel(theme),
		commandPalette: NewCommandPaletteModel(theme),
	}
}

// SetURL sets the gateway URL for display.
func (m *Model) SetURL(url string) {
	m.header.url = url
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.listenGateway(),
		m.tickCmd(),
		m.input.textarea.Focus(),
	)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		return m, nil

	case GatewayEventMsg:
		return m.handleGatewayEvent(gateway.GatewayEvent(msg))

	case TickMsg:
		m.activityBar.Tick()
		m.background.Tick()
		// Sync tasks to aquarium crabs
		if m.streaming {
			m.background.SetTasks([]string{waitingPhrases[m.activityBar.phrase]})
		} else {
			m.background.SetTasks(nil)
		}
		cmds = append(cmds, m.tickCmd())
		return m, tea.Batch(cmds...)

	case ConnectedMsg:
		m.connected = true
		m.header.SetConnected(true)
		m.statusBar.SetConnected(true)
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   "Connected to OpenClaw gateway.",
			Timestamp: time.Now(),
		})

		var cmds []tea.Cmd
		cmds = append(cmds, m.fetchModelInfo())

		// Send initial message if provided
		if m.initMessage != "" {
			msg := m.initMessage
			m.initMessage = ""
			cmds = append(cmds, m.sendMessage(msg))
		}
		return m, tea.Batch(cmds...)

	case SendResultMsg:
		if msg.Err != nil {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleError,
				Content:   fmt.Sprintf("Failed to send: %v", msg.Err),
				Timestamp: time.Now(),
			})
		}
		return m, nil

	case ModelInfoMsg:
		if msg.Model != "" {
			m.statusBar.SetModel(msg.Model)
		} else {
			// Fallback: read default model from OpenClaw config
			if cfgModel := readDefaultModel(); cfgModel != "" {
				m.statusBar.SetModel(cfgModel)
			} else {
				m.statusBar.SetModel("connected")
			}
		}
		return m, nil

	case StatusResultMsg:
		if msg.Err != nil {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleError,
				Content:   fmt.Sprintf("Status error: %v", msg.Err),
				Timestamp: time.Now(),
			})
		} else {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   msg.Content,
				Timestamp: time.Now(),
			})
		}
		return m, nil
	}

	// Pass through to input
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	// Pass through to chat viewport
	var chatCmd tea.Cmd
	m.chat, chatCmd = m.chat.Update(msg)
	cmds = append(cmds, chatCmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return m.theme.Muted.Render("Goodbye! 🦞\n")
	}

	margin := strings.Repeat(" ", sideMargin)
	inner := m.innerWidth()

	header := m.header.View()
	chat := m.chat.View()
	activity := m.activityBar.View()
	input := m.input.View()
	status := m.statusBar.View()

	// Thin separator line between chat and input
	sep := m.theme.Muted.Render(strings.Repeat("─", inner))

	// Apply side margins to each section
	header = addMarginToBlock(header, margin)
	chat = addMarginToBlock(chat, margin)
	sep = margin + sep
	input = addMarginToBlock(input, margin)
	status = addMarginToBlock(status, margin)

	var view string
	if activity != "" {
		activity = addMarginToBlock(activity, margin)
		view = fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", header, chat, sep, activity, input, status)
	} else {
		view = fmt.Sprintf("%s\n%s\n%s\n%s\n%s", header, chat, sep, input, status)
	}

	if m.background.IsActive() {
		view = m.background.ApplyToView(view, m.width, m.height)
	}

	// Overlay command palette
	if m.commandPalette.IsActive() {
		paletteView := m.commandPalette.View(m.width, m.height)
		view = overlayPalette(view, paletteView, m.width, m.height)
	}

	return view
}

// addMarginToBlock prepends a margin string to each line in a block.
func addMarginToBlock(block, margin string) string {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		lines[i] = margin + line
	}
	return strings.Join(lines, "\n")
}

// overlayPalette composites the palette view on top of the main view,
// positioning it near the top of the screen.
func overlayPalette(base, overlay string, width, height int) string {
	if overlay == "" {
		return base
	}
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Position palette starting at row 2 (below header)
	startRow := 2
	for i, ol := range overlayLines {
		row := startRow + i
		if row < len(baseLines) {
			baseLines[row] = ol
		}
	}
	return strings.Join(baseLines, "\n")
}

// sideMargin is the horizontal padding on each side of the content area.
const sideMargin = 2

func (m *Model) innerWidth() int {
	w := m.width - sideMargin*2
	if w < 40 {
		w = 40
	}
	return w
}

func (m *Model) layout() {
	inner := m.innerWidth()
	m.header.SetWidth(inner)
	m.statusBar.SetWidth(inner)
	m.input.SetWidth(inner)
	m.activityBar.SetWidth(inner)
	m.background.SetSize(m.width, m.height)

	// Chat gets remaining height: total - header(1) - input(5) - statusbar(1) - separator(1) - activity(1 if active) - borders
	chatHeight := m.height - 9
	if m.activityBar.IsActive() {
		chatHeight--
	}
	if chatHeight < 5 {
		chatHeight = 5
	}
	m.chat.SetSize(inner, chatHeight)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When command palette is active, route keys to it
	if m.commandPalette.IsActive() {
		return m.handlePaletteKey(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		now := time.Now()
		if now.Sub(m.lastCtrlC) < 500*time.Millisecond {
			m.quitting = true
			return m, tea.Quit
		}
		m.lastCtrlC = now
		m.ctrlCCount++

		// First Ctrl+C: clear input or abort
		if m.input.Value() != "" {
			m.input.Reset()
			return m, nil
		}
		if m.streaming {
			return m, m.abortRun()
		}
		// Show hint
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   "Press Ctrl+C again to exit.",
			Timestamp: time.Now(),
		})
		return m, nil

	case tea.KeyCtrlD:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyEscape:
		if m.streaming {
			return m, m.abortRun()
		}
		return m, nil

	case tea.KeyCtrlL:
		m.chat.Clear()
		return m, nil

	case tea.KeyEnter:
		// During a bracketed paste, insert a newline instead of submitting
		if msg.Paste {
			m.input.InsertNewline()
			return m, nil
		}
		return m.handleSubmit()
	}

	// Pass to input first, then check if we should open palette
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	// Sync palette with input: open when input starts with '/', close otherwise
	val := m.input.Value()
	if strings.HasPrefix(val, "/") {
		filter := val[1:]
		if !m.commandPalette.IsActive() {
			m.commandPalette.Open(filter)
		} else {
			m.commandPalette.SetFilter(filter)
		}
	} else if m.commandPalette.IsActive() {
		m.commandPalette.Close()
	}

	return m, cmd
}

func (m Model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.commandPalette.Close()
		return m, nil

	case tea.KeyUp:
		m.commandPalette.MoveUp()
		return m, nil

	case tea.KeyDown:
		m.commandPalette.MoveDown()
		return m, nil

	case tea.KeyEnter:
		result := m.commandPalette.Selected()
		if result == "" {
			// Sub-option picker just opened, stay in palette
			return m, nil
		}
		m.commandPalette.Close()
		m.input.Reset()
		// Execute the command
		if cmd := ParseCommand(result); cmd != nil {
			return m.handleCommand(cmd)
		}
		return m, nil

	case tea.KeyCtrlC:
		m.commandPalette.Close()
		m.input.Reset()
		return m, nil

	case tea.KeyBackspace:
		// Let input handle backspace, then sync filter
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		val := m.input.Value()
		if !strings.HasPrefix(val, "/") {
			m.commandPalette.Close()
		} else {
			m.commandPalette.SetFilter(val[1:])
		}
		return m, cmd

	default:
		// Let input handle the keystroke, then sync filter
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		val := m.input.Value()
		if strings.HasPrefix(val, "/") {
			m.commandPalette.SetFilter(val[1:])
		} else {
			m.commandPalette.Close()
		}
		return m, cmd
	}
}

func (m Model) handleSubmit() (tea.Model, tea.Cmd) {
	text := m.input.Value()
	if text == "" {
		return m, nil
	}

	m.input.Reset()

	// Check for slash commands
	if cmd := ParseCommand(text); cmd != nil {
		return m.handleCommand(cmd)
	}

	// Send chat message
	m.chat.AddMessage(ChatMsg{
		Role:      RoleUser,
		Content:   text,
		Timestamp: time.Now(),
	})

	return m, m.sendMessage(text)
}

func (m Model) handleCommand(cmd *Command) (tea.Model, tea.Cmd) {
	switch cmd.Name {
	case "help":
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   CommandHelp(m.theme),
			Timestamp: time.Now(),
		})

	case "exit", "quit":
		m.quitting = true
		return m, tea.Quit

	case "clear":
		m.chat.Clear()

	case "theme":
		name := ThemeName(cmd.Args)
		if cmd.Args == "" {
			// Show available themes
			msg := "Available themes: ocean, amber, rose, forest, aquarium\nUsage: /theme <name>"
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   msg,
				Timestamp: time.Now(),
			})
			return m, nil
		}
		m.setTheme(name)
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   fmt.Sprintf("Theme switched to %s.", cmd.Args),
			Timestamp: time.Now(),
		})

	case "bg":
		if cmd.Args == "" {
			mode := m.background.CycleMode()
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("Background: %s", mode),
				Timestamp: time.Now(),
			})
		} else {
			mode := BgMode(cmd.Args)
			valid := false
			for _, bm := range BgModes {
				if bm == mode {
					valid = true
					break
				}
			}
			if !valid {
				m.chat.AddMessage(ChatMsg{
					Role:      RoleError,
					Content:   fmt.Sprintf("Unknown background mode: %s\nAvailable: off, starfield, tunnel, plasma, fire, matrix, ocean, cube, skibidi, sigma, npc, ohio, rizz, gyatt, amogus, bussin, aquarium", cmd.Args),
					Timestamp: time.Now(),
				})
				return m, nil
			}
			m.background.SetMode(mode)
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("Background: %s", mode),
				Timestamp: time.Now(),
			})
		}

	case "think":
		if cmd.Args == "" {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("Current thinking level: %s\nUsage: /think <none|adaptive|full>", m.thinking),
				Timestamp: time.Now(),
			})
			return m, nil
		}
		m.thinking = cmd.Args
		m.statusBar.SetThinking(m.thinking)
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   fmt.Sprintf("Thinking level set to %s.", cmd.Args),
			Timestamp: time.Now(),
		})

	case "status":
		return m, m.requestStatus()

	case "model":
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   "Model switching: use /model <name> (requires gateway support).",
			Timestamp: time.Now(),
		})

	case "session":
		if cmd.Args != "" {
			m.sessionKey = cmd.Args
			m.statusBar.SetSession(m.sessionKey)
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("Session switched to %s.", cmd.Args),
				Timestamp: time.Now(),
			})
		} else {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("Current session: %s\nUsage: /session <key>", m.sessionKey),
				Timestamp: time.Now(),
			})
		}

	case "abort":
		if m.streaming {
			return m, m.abortRun()
		}
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   "No active run to abort.",
			Timestamp: time.Now(),
		})

	case "feed":
		if m.background.Mode() != BgAquarium {
			m.background.SetMode(BgAquarium)
		}
		m.background.DropFood(5)
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   "🐟 Food dropped! Watch the fish swim toward it.",
			Timestamp: time.Now(),
		})

	case "new", "reset":
		m.chat.Clear()
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   "Session reset. Chat history cleared.",
			Timestamp: time.Now(),
		})

	default:
		m.chat.AddMessage(ChatMsg{
			Role:      RoleError,
			Content:   fmt.Sprintf("Unknown command: /%s. Type /help for available commands.", cmd.Name),
			Timestamp: time.Now(),
		})
	}

	return m, nil
}

func (m Model) handleGatewayEvent(evt gateway.GatewayEvent) (tea.Model, tea.Cmd) {
	switch evt.Type {
	case "connected":
		m.connected = true
		m.header.SetConnected(true)
		m.statusBar.SetConnected(true)

	case "disconnected":
		m.connected = false
		m.header.SetConnected(false)
		m.statusBar.SetConnected(false)
		errMsg := "Connection lost."
		if evt.Error != nil {
			errMsg = fmt.Sprintf("Connection lost: %v", evt.Error)
		}
		m.chat.AddMessage(ChatMsg{
			Role:      RoleError,
			Content:   errMsg,
			Timestamp: time.Now(),
		})

	case "chat":
		if evt.Chat != nil {
			m.handleChatEvent(evt.Chat)
		}

	case "agent":
		if evt.Agent != nil {
			m.handleAgentEvent(evt.Agent)
		}

	case "btw":
		if evt.BTW != nil {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("💡 %s", evt.BTW.Message),
				Timestamp: time.Now(),
			})
		}

	case "session.update":
		// Parse session update if needed
		var info gateway.SessionInfo
		if evt.Payload != nil {
			json.Unmarshal(evt.Payload, &info)
			if info.Model != "" {
				m.statusBar.SetModel(info.Model)
			}
		}

	case "error":
		errMsg := "Gateway error"
		if evt.Error != nil {
			errMsg = evt.Error.Error()
		}
		m.chat.AddMessage(ChatMsg{
			Role:      RoleError,
			Content:   errMsg,
			Timestamp: time.Now(),
		})
	}

	return m, m.listenGateway()
}

func (m *Model) handleAgentEvent(agent *gateway.AgentEvent) {
	switch agent.Stream {
	case "assistant":
		var data gateway.AgentAssistantData
		if err := json.Unmarshal(agent.Data, &data); err != nil {
			return
		}
		if !m.streaming {
			m.streaming = true
			m.activeRunID = agent.RunID
			m.streamBuf = ""
			m.activityBar.Start()
			m.chat.AddMessage(ChatMsg{
				Role:      RoleAssistant,
				Content:   "",
				Timestamp: time.Now(),
				Streaming: true,
				RunID:     agent.RunID,
			})
			m.layout()
		}
		m.streamBuf = data.Text
		m.chat.UpdateLastAssistant(m.streamBuf, true)

	case "lifecycle":
		var data gateway.AgentLifecycleData
		if err := json.Unmarshal(agent.Data, &data); err != nil {
			return
		}
		if data.Phase == "end" && m.streaming {
			m.streaming = false
			m.activityBar.Stop()
			m.chat.UpdateLastAssistant(m.streamBuf, false)
			m.streamBuf = ""
			m.activeRunID = ""
			m.layout()
		}
	}
}

func (m *Model) handleChatEvent(chat *gateway.ChatEvent) {
	switch chat.State {
	case gateway.ChatStateDelta:
		if !m.streaming {
			m.streaming = true
			m.activeRunID = chat.RunID
			m.streamBuf = ""
			m.activityBar.Start()
			m.chat.AddMessage(ChatMsg{
				Role:      RoleAssistant,
				Content:   "",
				Timestamp: time.Now(),
				Streaming: true,
				RunID:     chat.RunID,
			})
			m.layout()
		}
		if chat.Message != nil {
			m.streamBuf += chat.Message.Content
			m.chat.UpdateLastAssistant(m.streamBuf, true)
		}

	case gateway.ChatStateFinal:
		m.streaming = false
		m.activityBar.Stop()
		if chat.Message != nil {
			m.chat.UpdateLastAssistant(chat.Message.Content, false)
		} else {
			m.chat.UpdateLastAssistant(m.streamBuf, false)
		}
		m.streamBuf = ""
		m.activeRunID = ""
		m.layout()

	case gateway.ChatStateAborted:
		m.streaming = false
		m.activityBar.Stop()
		m.chat.UpdateLastAssistant(m.streamBuf+" [aborted]", false)
		m.streamBuf = ""
		m.activeRunID = ""
		m.layout()

	case gateway.ChatStateError:
		m.streaming = false
		m.activityBar.Stop()
		errMsg := "Unknown error"
		if chat.ErrorMessage != "" {
			errMsg = chat.ErrorMessage
		}
		m.chat.AddMessage(ChatMsg{
			Role:      RoleError,
			Content:   errMsg,
			Timestamp: time.Now(),
		})
		m.streamBuf = ""
		m.activeRunID = ""
		m.layout()
	}
}

func (m *Model) setTheme(name ThemeName) {
	m.theme = NewTheme(name)
	m.header.SetTheme(m.theme)
	m.chat.SetTheme(m.theme)
	m.input.SetTheme(m.theme)
	m.statusBar.SetTheme(m.theme)
	m.activityBar.SetTheme(m.theme)
	m.background.SetTheme(m.theme)
	m.commandPalette.SetTheme(m.theme)
}

// Tea commands

func (m Model) listenGateway() tea.Cmd {
	return func() tea.Msg {
		evt := <-m.gateway.Events()
		return GatewayEventMsg(evt)
	}
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) sendMessage(text string) tea.Cmd {
	return func() tea.Msg {
		params := gateway.ChatSendParams{
			SessionKey:     m.sessionKey,
			Message:        text,
			Thinking:       m.thinking,
			IdempotencyKey: fmt.Sprintf("%d", time.Now().UnixNano()),
		}
		// Use fire-and-forget — the actual response comes as chat events
		err := m.gateway.RequestAsync(gateway.MethodChatSend, params)
		return SendResultMsg{Err: err}
	}
}

func (m Model) abortRun() tea.Cmd {
	return func() tea.Msg {
		err := m.gateway.RequestAsync(gateway.MethodChatAbort, gateway.ChatAbortParams{
			SessionKey: m.sessionKey,
			RunID:      m.activeRunID,
		})
		if err != nil {
			return SendResultMsg{Err: fmt.Errorf("abort: %w", err)}
		}
		return nil
	}
}

func (m Model) requestStatus() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.gateway.Request(gateway.MethodStatus, nil)
		if err != nil {
			return StatusResultMsg{Err: fmt.Errorf("status: %w", err)}
		}

		var status gateway.StatusPayload
		if resp.Payload != nil {
			json.Unmarshal(resp.Payload, &status)
		}

		content := fmt.Sprintf("Gateway version: %s\nSessions: %d", status.Version, len(status.Sessions))
		for _, s := range status.Sessions {
			content += fmt.Sprintf("\n  • %s", s.Key)
			if s.Model != "" {
				content += fmt.Sprintf(" (model: %s)", s.Model)
			}
		}
		return StatusResultMsg{Content: content}
	}
}

func (m Model) fetchModelInfo() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.gateway.Request(gateway.MethodStatus, nil)
		if err != nil {
			return ModelInfoMsg{}
		}
		var status gateway.StatusPayload
		if resp.Payload != nil {
			json.Unmarshal(resp.Payload, &status)
		}
		for _, s := range status.Sessions {
			if s.Model != "" {
				return ModelInfoMsg{Model: s.Model}
			}
		}
		return ModelInfoMsg{}
	}
}

// readDefaultModel reads the default model from ~/.openclaw/openclaw.json
func readDefaultModel() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	cfgPath := filepath.Join(home, ".openclaw", "openclaw.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return ""
	}
	// Strip UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}
	var cfg struct {
		Models struct {
			Default string `json:"default"`
		} `json:"models"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.Models.Default
}
