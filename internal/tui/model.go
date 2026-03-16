package tui

import (
	"encoding/json"
	"fmt"
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
)

// Model is the main Bubble Tea model.
type Model struct {
	gateway     *gateway.Client
	sessionKey  string
	thinking    string
	initMessage string

	// UI components
	header      HeaderModel
	chat        ChatModel
	input       InputModel
	statusBar   StatusBarModel
	activityBar ActivityBarModel
	background  BackgroundModel
	theme       Theme

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
		background:  NewBackgroundModel(theme),
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

		// Send initial message if provided
		if m.initMessage != "" {
			msg := m.initMessage
			m.initMessage = ""
			return m, m.sendMessage(msg)
		}
		return m, nil

	case SendResultMsg:
		if msg.Err != nil {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleError,
				Content:   fmt.Sprintf("Failed to send: %v", msg.Err),
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

	header := m.header.View()
	chat := m.chat.View()
	activity := m.activityBar.View()
	input := m.input.View()
	status := m.statusBar.View()

	var view string
	if activity != "" {
		view = fmt.Sprintf("%s\n%s\n%s\n%s\n%s", header, chat, activity, input, status)
	} else {
		view = fmt.Sprintf("%s\n%s\n%s\n%s", header, chat, input, status)
	}

	if m.background.IsActive() {
		view = m.background.ApplyToView(view, m.width, m.height)
	}

	return view
}

func (m *Model) layout() {
	m.header.SetWidth(m.width)
	m.statusBar.SetWidth(m.width)
	m.input.SetWidth(m.width)
	m.activityBar.SetWidth(m.width)
	m.background.SetSize(m.width, m.height)

	// Chat gets remaining height: total - header(1) - input(5) - statusbar(1) - activity(1 if active) - borders
	chatHeight := m.height - 8
	if m.activityBar.IsActive() {
		chatHeight--
	}
	if chatHeight < 5 {
		chatHeight = 5
	}
	m.chat.SetSize(m.width, chatHeight)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.handleSubmit()
	}

	// Pass to input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
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
			msg := "Available themes: ocean, amber, rose, forest\nUsage: /theme <name>"
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
			for _, m := range BgModes {
				if m == mode {
					valid = true
					break
				}
			}
			if !valid {
				m.chat.AddMessage(ChatMsg{
					Role:      RoleError,
					Content:   fmt.Sprintf("Unknown background mode: %s\nAvailable: off, wave, matrix, aurora, rain, particles, pulse", cmd.Args),
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
			SessionKey: m.sessionKey,
			Message:    text,
			Thinking:   m.thinking,
		}
		_, err := m.gateway.Request(gateway.MethodChatSend, params)
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
			return SendResultMsg{Err: fmt.Errorf("status: %w", err)}
		}

		var status gateway.StatusPayload
		if resp.Payload != nil {
			json.Unmarshal(resp.Payload, &status)
		}

		content := fmt.Sprintf("Gateway version: %s\nSessions: %d", status.Version, len(status.Sessions))
		// We can't directly add a message from a Cmd, so we'll use SendResult for errors
		// For now, return nil and handle status display differently
		_ = content
		return nil
	}
}
