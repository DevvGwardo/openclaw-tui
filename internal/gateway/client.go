package gateway

import (
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	defaultTickInterval = 30 * time.Second
	maxBackoff          = 30 * time.Second
	protocolVersion     = 3
	clientVersion       = "0.1.0"
)

// GatewayEvent is sent to the TUI via a channel.
type GatewayEvent struct {
	Type    string // "connected", "disconnected", "chat", "tick", "session.update", "btw", "error"
	Chat    *ChatEvent
	BTW     *BTWPayload
	Error   error
	Payload json.RawMessage
}

// Client manages the WebSocket connection to the OpenClaw gateway.
type Client struct {
	url      string
	token    string
	password string

	conn     *websocket.Conn
	connMu   sync.Mutex
	events   chan GatewayEvent
	pending  map[string]chan *ResponseFrame
	pendMu   sync.Mutex
	done     chan struct{}
	closed   bool

	tickInterval time.Duration
	lastTick     time.Time
	tickMu       sync.Mutex

	reconnecting bool
	reconnectMu  sync.Mutex
	backoff      int
}

// NewClient creates a new gateway client.
func NewClient(url, token, password string) *Client {
	return &Client{
		url:          url,
		token:        token,
		password:     password,
		events:       make(chan GatewayEvent, 64),
		pending:      make(map[string]chan *ResponseFrame),
		done:         make(chan struct{}),
		tickInterval: defaultTickInterval,
	}
}

// Events returns the channel of gateway events for the TUI to consume.
func (c *Client) Events() <-chan GatewayEvent {
	return c.events
}

// Connect establishes the WebSocket connection and performs auth handshake.
func (c *Client) Connect() error {
	c.connMu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(c.url, nil)
	if err != nil {
		c.connMu.Unlock()
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn
	c.connMu.Unlock()

	// Read the challenge - read raw first to handle any frame shape
	_, rawMsg, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read challenge: %w", err)
	}
	var frame Frame
	if err := json.Unmarshal(rawMsg, &frame); err != nil {
		return fmt.Errorf("parse challenge frame: %w", err)
	}
	if !isEventFrame(frame.Type) || frame.Event != EventConnectChallenge {
		return fmt.Errorf("expected connect.challenge, got type=%q event=%q", frame.Type, frame.Event)
	}

	// Send connect request directly (readLoop not started yet)
	connectID := uuid.New().String()
	connectFrame := RequestFrame{
		Type:   FrameTypeRequest,
		ID:     connectID,
		Method: MethodConnect,
		Params: ConnectParams{
			MinProtocol: protocolVersion,
			MaxProtocol: protocolVersion,
			Client: ClientInfo{
				ID:          "gateway-client",
				DisplayName: "openclaw-tui",
				Version:     clientVersion,
				Platform:    runtime.GOOS,
				Mode:        "backend",
			},
			Caps: []string{},
			Auth: AuthInfo{
				Token:    c.token,
				Password: c.password,
			},
			Role:   "operator",
			Scopes: []string{"operator.admin", "operator.read", "operator.write", "operator.approvals"},
		},
	}
	if err := conn.WriteJSON(connectFrame); err != nil {
		return fmt.Errorf("write connect: %w", err)
	}

	// Read connect response directly
	_, rawResp, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read connect response: %w", err)
	}
	var respFrame Frame
	if err := json.Unmarshal(rawResp, &respFrame); err != nil {
		return fmt.Errorf("parse connect response: %w", err)
	}
	if !isResponseFrame(respFrame.Type) {
		return fmt.Errorf("expected response, got type=%q", respFrame.Type)
	}
	if respFrame.OK == nil || !*respFrame.OK {
		msg := "unknown error"
		if respFrame.Error != nil {
			msg = respFrame.Error.Message
		}
		return fmt.Errorf("connect rejected: %s", msg)
	}

	c.backoff = 0
	c.tickMu.Lock()
	c.lastTick = time.Now()
	c.tickMu.Unlock()

	c.emit(GatewayEvent{Type: "connected"})

	go c.readLoop()
	go c.watchdogLoop()

	return nil
}

// Request sends a request and waits for a correlated response.
func (c *Client) Request(method string, params interface{}) (*ResponseFrame, error) {
	id := uuid.New().String()
	ch := make(chan *ResponseFrame, 1)

	c.pendMu.Lock()
	c.pending[id] = ch
	c.pendMu.Unlock()

	defer func() {
		c.pendMu.Lock()
		delete(c.pending, id)
		c.pendMu.Unlock()
	}()

	frame := RequestFrame{
		Type:   FrameTypeRequest,
		ID:     id,
		Method: method,
		Params: params,
	}

	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	if err := conn.WriteJSON(frame); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("request timeout")
	case <-c.done:
		return nil, fmt.Errorf("client closed")
	}
}

// RequestAsync sends a request without waiting for a response.
func (c *Client) RequestAsync(method string, params interface{}) error {
	id := uuid.New().String()
	frame := RequestFrame{
		Type:   FrameTypeRequest,
		ID:     id,
		Method: method,
		Params: params,
	}

	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	return conn.WriteJSON(frame)
}

// Close shuts down the client.
func (c *Client) Close() {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.closed {
		return
	}
	c.closed = true
	close(c.done)

	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) emit(evt GatewayEvent) {
	select {
	case c.events <- evt:
	default:
		// Drop if buffer full
	}
}

func (c *Client) readLoop() {
	for {
		c.connMu.Lock()
		conn := c.conn
		c.connMu.Unlock()

		if conn == nil {
			return
		}

		var frame Frame
		if err := conn.ReadJSON(&frame); err != nil {
			select {
			case <-c.done:
				return
			default:
			}
			c.emit(GatewayEvent{Type: "disconnected", Error: err})
			c.scheduleReconnect()
			return
		}

		switch {
		case isEventFrame(frame.Type):
			c.handleEvent(frame)
		case isResponseFrame(frame.Type):
			c.handleResponse(frame)
		}
	}
}

func (c *Client) handleEvent(frame Frame) {
	switch frame.Event {
	case EventTick:
		c.tickMu.Lock()
		c.lastTick = time.Now()
		c.tickMu.Unlock()

		var tick TickPayload
		if frame.Payload != nil {
			json.Unmarshal(frame.Payload, &tick)
			if tick.Interval > 0 {
				c.tickMu.Lock()
				c.tickInterval = time.Duration(tick.Interval) * time.Second
				c.tickMu.Unlock()
			}
		}

	case EventChat:
		var chat ChatEvent
		if err := json.Unmarshal(frame.Payload, &chat); err == nil {
			c.emit(GatewayEvent{Type: "chat", Chat: &chat})
		}

	case EventSessionUpdate:
		c.emit(GatewayEvent{Type: "session.update", Payload: frame.Payload})

	case EventBTW:
		var btw BTWPayload
		if err := json.Unmarshal(frame.Payload, &btw); err == nil {
			c.emit(GatewayEvent{Type: "btw", BTW: &btw})
		}
	}
}

func (c *Client) handleResponse(frame Frame) {
	if frame.ID == "" {
		return
	}

	c.pendMu.Lock()
	ch, ok := c.pending[frame.ID]
	c.pendMu.Unlock()

	if ok {
		okVal := false
		if frame.OK != nil {
			okVal = *frame.OK
		}
		resp := &ResponseFrame{
			ID:      frame.ID,
			OK:      okVal,
			Payload: frame.Payload,
			Error:   frame.Error,
		}
		select {
		case ch <- resp:
		default:
		}
	}
}

func (c *Client) watchdogLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.tickMu.Lock()
			deadline := c.lastTick.Add(c.tickInterval * 2)
			c.tickMu.Unlock()

			if time.Now().After(deadline) {
				c.emit(GatewayEvent{Type: "disconnected", Error: fmt.Errorf("tick watchdog timeout")})
				c.scheduleReconnect()
				return
			}
		case <-c.done:
			return
		}
	}
}

func isEventFrame(t string) bool {
	return t == FrameTypeEvent || t == FrameTypeEventAlt
}

func isResponseFrame(t string) bool {
	return t == FrameTypeResponse || t == FrameTypeResponseAlt
}

func (c *Client) scheduleReconnect() {
	c.reconnectMu.Lock()
	if c.reconnecting {
		c.reconnectMu.Unlock()
		return
	}
	c.reconnecting = true
	c.reconnectMu.Unlock()

	go func() {
		defer func() {
			c.reconnectMu.Lock()
			c.reconnecting = false
			c.reconnectMu.Unlock()
		}()

		for {
			select {
			case <-c.done:
				return
			default:
			}

			delay := time.Duration(math.Min(
				float64(time.Second)*math.Pow(2, float64(c.backoff)),
				float64(maxBackoff),
			))
			c.backoff++

			time.Sleep(delay)

			if err := c.Connect(); err != nil {
				c.emit(GatewayEvent{Type: "error", Error: fmt.Errorf("reconnect: %w", err)})
				continue
			}
			return
		}
	}()
}
