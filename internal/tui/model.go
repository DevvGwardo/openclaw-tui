package tui

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
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
	width              int
	height             int
	connected          bool
	streaming          bool
	streamBuf          string
	activeRunID        string
	ctrlCCount         int
	lastCtrlC          time.Time
	quitting           bool
	mouseMode          bool
	mouseFilter        mouseFilter // suppresses leaked mouse escape sequence fragments
	pendingAttachments []gateway.Attachment
	err                error
}

// NewModel creates the main TUI model.
func NewModel(gw *gateway.Client, sessionKey, thinking, initMessage string, themeName ThemeName) Model {
	theme := NewTheme(themeName)

	return Model{
		gateway:     gw,
		sessionKey:  sessionKey,
		thinking:    thinking,
		initMessage: initMessage,
		mouseMode:   true,
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
	case tea.MouseMsg:
		// Record that we received a real mouse event so the filter can
		// suppress any escape sequence fragments that follow.
		m.mouseFilter.onMouseEvent()
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.chat.ScrollUp(3)
			return m, nil
		case tea.MouseButtonWheelDown:
			m.chat.ScrollDown(3)
			return m, nil
		}
		return m, nil

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
			cmds = append(cmds, m.sendMessage(msg, nil))
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

	// Pass through to input only — do NOT forward to chat viewport here.
	// The viewport has its own key handlers (Up/Down/PgUp/PgDown) that
	// conflict with the textarea, causing unwanted scrolling when the
	// user types. Chat scrolling is handled explicitly via keybindings
	// above (PgUp, PgDown, Home, End, Ctrl+Up, Ctrl+Down).
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

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

	// Overlay crab task labels on top of everything (so they appear above chat UI)
	if m.background.Mode() == BgAquarium {
		view = m.background.OverlayCrabLabelsOnView(view, m.width, m.height)
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

// mouseFilter suppresses mouse escape sequence fragments that leak through
// Bubble Tea's input parser during scroll wheel events.
//
// SGR mouse format: ESC [ < Btn ; X ; Y M/m
// When Bubble Tea consumes the ESC but fails to parse the rest, fragments like
// "[<64;66;36M" arrive as individual KeyRunes. This filter catches them using:
//  1. Time-based: after a real tea.MouseMsg, suppress mouse-like chars for 200ms
//  2. Pattern-based: multi-char arrivals matching mouse sequence patterns
//  3. Bracket tracking: tentatively eat lone '[' and confirm/replay based on next char
type mouseFilter struct {
	lastMouseEvent time.Time // when we last saw a real tea.MouseMsg
	pendingBracket bool      // true if we ate a '[' that might start a mouse seq
	bracketTime    time.Time // when the pending bracket was received
}

// mouseFilterChars is the set of characters found in SGR mouse escape sequences.
var mouseFilterChars = func() [256]bool {
	var t [256]bool
	for _, c := range []byte("[]<>0123456789;Mm") {
		t[c] = true
	}
	return t
}()

// isMouseFragment reports whether s consists entirely of mouse-sequence characters.
func isMouseFragment(s string) bool {
	for _, r := range s {
		if r > 255 || !mouseFilterChars[byte(r)] {
			return false
		}
	}
	return len(s) > 0
}

// onMouseEvent records that a real mouse event was received from Bubble Tea.
func (f *mouseFilter) onMouseEvent() {
	f.lastMouseEvent = time.Now()
}

// shouldSuppress returns true if msg looks like a mouse escape fragment.
// If a previously eaten '[' turns out to be real input, it calls replayBracket
// to re-inject it into the input.
func (f *mouseFilter) shouldSuppress(msg tea.KeyMsg, replayBracket func()) bool {
	if msg.Type != tea.KeyRunes {
		// Non-rune key: if we had a pending bracket, replay it
		if f.pendingBracket {
			replayBracket()
			f.pendingBracket = false
		}
		return false
	}

	s := string(msg.Runes)
	if len(s) == 0 {
		return false
	}
	now := time.Now()

	// --- Multi-char pattern detection ---
	// Catches complete or large fragments arriving at once (e.g. "[<64;66;36M")
	if len(s) > 2 && isMouseFragment(s) {
		return true
	}
	if strings.Contains(s, "[<") || strings.Contains(s, "[M") {
		return true
	}

	// --- Time-based suppression ---
	// Within 200ms of a real mouse event, suppress any mouse-like characters.
	// This catches the remaining digits/semicolons/M/m that follow a partially
	// parsed sequence. Nobody types digits/semicolons during a scroll.
	recentMouse := now.Sub(f.lastMouseEvent) < 200*time.Millisecond
	if recentMouse && isMouseFragment(s) {
		return true
	}

	// --- Bracket tracking ---
	// Handle pending bracket from a previous call
	if f.pendingBracket {
		if now.Sub(f.bracketTime) > 100*time.Millisecond {
			// Timeout: the '[' was a real keystroke, replay it
			replayBracket()
			f.pendingBracket = false
			// Fall through to process current char normally
		} else if s == "<" {
			// Confirmed mouse sequence start: [<
			// Set lastMouseEvent so the time-based filter catches the rest
			f.pendingBracket = false
			f.lastMouseEvent = now
			return true
		} else {
			// Not a mouse sequence, replay the '[' and process current char
			replayBracket()
			f.pendingBracket = false
			// Fall through
		}
	}

	// A lone '[' might start a mouse sequence — eat it tentatively.
	// Only do this when there's no recent mouse event (if recent, the
	// time-based filter above already handled it).
	if s == "[" && !recentMouse {
		f.pendingBracket = true
		f.bracketTime = now
		return true
	}

	return false
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Suppress stray mouse escape sequence fragments.
	if m.mouseFilter.shouldSuppress(msg, func() {
		m.input.InsertRune('[')
	}) {
		return m, nil
	}

	// When command palette is active, route keys to it
	if m.commandPalette.IsActive() {
		return m.handlePaletteKey(msg)
	}

	// Alt+Up / Alt+Down for reliable scrolling (Ctrl+Arrow often captured by OS)
	if msg.Alt {
		switch msg.Type {
		case tea.KeyUp:
			m.chat.ScrollUp(1)
			return m, nil
		case tea.KeyDown:
			m.chat.ScrollDown(1)
			return m, nil
		}
	}

	// Alt+M toggles mouse mode: cell motion (scroll) ↔ all motion (full tracking).
	// Mouse is never fully disabled to avoid raw escape sequences leaking into
	// the textarea. To select text for copy, hold Shift while clicking/dragging
	// (supported by most terminals: iTerm2, Terminal.app, Alacritty, WezTerm).
	if msg.Type == tea.KeyRunes && msg.Alt && string(msg.Runes) == "m" {
		m.mouseMode = !m.mouseMode
		m.statusBar.SetMouseMode(m.mouseMode)
		if m.mouseMode {
			return m, tea.EnableMouseCellMotion
		}
		return m, tea.EnableMouseAllMotion
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

	case tea.KeyPgUp:
		m.chat.ScrollUp(m.chat.Height() / 2)
		return m, nil

	case tea.KeyPgDown:
		m.chat.ScrollDown(m.chat.Height() / 2)
		return m, nil

	case tea.KeyHome:
		m.chat.ScrollToTop()
		return m, nil

	case tea.KeyEnd:
		m.chat.ScrollToBottom()
		return m, nil

	case tea.KeyCtrlUp:
		m.chat.ScrollUp(1)
		return m, nil

	case tea.KeyCtrlDown:
		m.chat.ScrollDown(1)
		return m, nil

	case tea.KeyEnter:
		// Alt+Enter inserts a newline instead of submitting
		if msg.Alt {
			m.input.InsertNewline()
			return m, nil
		}
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

	// Build display content with attachment indicators
	displayContent := text
	if len(m.pendingAttachments) > 0 {
		var names []string
		for _, a := range m.pendingAttachments {
			names = append(names, a.Name)
		}
		displayContent = fmt.Sprintf("[%s]\n%s", strings.Join(names, ", "), text)
	}

	// Send chat message
	m.chat.AddMessage(ChatMsg{
		Role:      RoleUser,
		Content:   displayContent,
		Timestamp: time.Now(),
	})

	// Capture and clear pending attachments
	attachments := m.pendingAttachments
	m.pendingAttachments = nil
	m.input.SetAttachmentCount(0)

	return m, m.sendMessage(text, attachments)
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
		if cmd.Args == "" {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("Current model: %s\nUsage: /model <name>", m.statusBar.Model()),
				Timestamp: time.Now(),
			})
			return m, nil
		}
		return m, m.setModel(cmd.Args)

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

	case "attach", "image":
		if cmd.Args == "" {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("Usage: /attach <file-path>\nSupported: PNG, JPG, JPEG, GIF, WEBP\nPending attachments: %d", len(m.pendingAttachments)),
				Timestamp: time.Now(),
			})
			return m, nil
		}
		att, err := loadAttachment(cmd.Args)
		if err != nil {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleError,
				Content:   fmt.Sprintf("Failed to attach: %v", err),
				Timestamp: time.Now(),
			})
			return m, nil
		}
		m.pendingAttachments = append(m.pendingAttachments, att)
		m.input.SetAttachmentCount(len(m.pendingAttachments))
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   fmt.Sprintf("Attached %s (%s). Send a message to include it.", att.Name, att.Type),
			Timestamp: time.Now(),
		})

	case "unattach", "detach":
		if len(m.pendingAttachments) == 0 {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   "No pending attachments to remove.",
				Timestamp: time.Now(),
			})
			return m, nil
		}
		if cmd.Args == "" || cmd.Args == "all" {
			m.pendingAttachments = nil
			m.input.SetAttachmentCount(0)
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   "All attachments removed.",
				Timestamp: time.Now(),
			})
		} else {
			// Remove by filename
			var kept []gateway.Attachment
			removed := false
			for _, a := range m.pendingAttachments {
				if a.Name == cmd.Args && !removed {
					removed = true
					continue
				}
				kept = append(kept, a)
			}
			if !removed {
				m.chat.AddMessage(ChatMsg{
					Role:      RoleError,
					Content:   fmt.Sprintf("No attachment named %q found.", cmd.Args),
					Timestamp: time.Now(),
				})
				return m, nil
			}
			m.pendingAttachments = kept
			m.input.SetAttachmentCount(len(m.pendingAttachments))
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   fmt.Sprintf("Removed %s. %d attachment(s) remaining.", cmd.Args, len(m.pendingAttachments)),
				Timestamp: time.Now(),
			})
		}

	case "new", "reset":
		m.chat.Clear()
		m.pendingAttachments = nil
		m.input.SetAttachmentCount(0)
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
	// Debug: log all gateway events to file
	if df, err := os.OpenFile("/tmp/openclaw-tui-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(df, "[%s] event type=%q", time.Now().Format("15:04:05"), evt.Type)
		if evt.Chat != nil {
			fmt.Fprintf(df, " chat.state=%q msg_nil=%v", evt.Chat.State, evt.Chat.Message == nil)
			if evt.Chat.Message != nil {
				fmt.Fprintf(df, " role=%q content_len=%d content=%q", evt.Chat.Message.Role, len(evt.Chat.Message.Content), evt.Chat.Message.Content)
			}
		}
		if evt.Agent != nil {
			fmt.Fprintf(df, " agent.stream=%q data=%s", evt.Agent.Stream, string(evt.Agent.Data))
		}
		if evt.Payload != nil {
			if evt.Type == "chat" || evt.Type == "agent" {
				fmt.Fprintf(df, " raw_payload=%s", string(evt.Payload))
			}
		}
		fmt.Fprintln(df)
		df.Close()
	}

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
	interval := 500 * time.Millisecond
	if m.streaming || m.background.IsActive() {
		interval = 100 * time.Millisecond
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) sendMessage(text string, attachments []gateway.Attachment) tea.Cmd {
	return func() tea.Msg {
		params := gateway.ChatSendParams{
			SessionKey:     m.sessionKey,
			Message:        text,
			Thinking:       m.thinking,
			IdempotencyKey: fmt.Sprintf("%d", time.Now().UnixNano()),
			Attachments:    attachments,
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

func (m Model) setModel(modelName string) tea.Cmd {
	return func() tea.Msg {
		params := gateway.SessionsPatchParams{
			SessionKey: m.sessionKey,
		}
		params.Patch.Model = modelName
		
		resp, err := m.gateway.Request(gateway.MethodSessionsPatch, params)
		if err != nil {
			return StatusResultMsg{Err: fmt.Errorf("set model: %w", err)}
		}
		
		if !resp.OK {
			errMsg := "failed to set model"
			if resp.Error != nil {
				errMsg = resp.Error.Message
			}
			return StatusResultMsg{Err: fmt.Errorf("set model: %s", errMsg)}
		}
		
		// Update local display
		m.statusBar.SetModel(modelName)
		return StatusResultMsg{Content: fmt.Sprintf("Model switched to %s.", modelName)}
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

// supportedImageExts lists file extensions we accept as image attachments.
var supportedImageExts = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// maxAttachmentSize is the maximum file size for an attachment (10 MB).
const maxAttachmentSize = 10 * 1024 * 1024

// loadAttachment reads a file from disk and returns a base64-encoded Attachment.
func loadAttachment(path string) (gateway.Attachment, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return gateway.Attachment{}, fmt.Errorf("invalid path: %w", err)
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(absPath))
	mimeType, ok := supportedImageExts[ext]
	if !ok {
		// Try mime package as fallback
		mimeType = mime.TypeByExtension(ext)
		if mimeType == "" || !strings.HasPrefix(mimeType, "image/") {
			return gateway.Attachment{}, fmt.Errorf("unsupported file type %q (supported: png, jpg, gif, webp)", ext)
		}
	}

	// Read file
	info, err := os.Stat(absPath)
	if err != nil {
		return gateway.Attachment{}, fmt.Errorf("file not found: %w", err)
	}
	if info.Size() > maxAttachmentSize {
		return gateway.Attachment{}, fmt.Errorf("file too large (%d MB, max 10 MB)", info.Size()/(1024*1024))
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return gateway.Attachment{}, fmt.Errorf("read file: %w", err)
	}

	return gateway.Attachment{
		Type: mimeType,
		Name: filepath.Base(absPath),
		Data: base64.StdEncoding.EncodeToString(data),
	}, nil
}
