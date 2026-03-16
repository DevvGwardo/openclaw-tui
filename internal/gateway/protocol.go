package gateway

import "encoding/json"

// Frame types — gateway sends long-form, clients send short-form
const (
	FrameTypeEvent    = "event"
	FrameTypeEventAlt = "evt"
	FrameTypeRequest  = "req"
	FrameTypeResponse = "response"
	FrameTypeResponseAlt = "res"
)

// Event names
const (
	EventConnectChallenge = "connect.challenge"
	EventTick             = "tick"
	EventChat             = "chat"
	EventAgent            = "agent"
	EventSessionUpdate    = "session.update"
	EventBTW              = "btw"
)

// Request methods
const (
	MethodConnect      = "connect"
	MethodChatSend     = "chat.send"
	MethodChatAbort    = "chat.abort"
	MethodSessionsList = "sessions.list"
	MethodSessionsPatch = "sessions.patch"
	MethodAgentsList   = "agents.list"
	MethodStatus       = "status"
)

// Chat states
const (
	ChatStateDelta   = "delta"
	ChatStateFinal   = "final"
	ChatStateAborted = "aborted"
	ChatStateError   = "error"
)

// Frame is a raw WebSocket frame that can be any type.
type Frame struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Event   string          `json:"event,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Seq     int             `json:"seq,omitempty"`
	OK      *bool           `json:"ok,omitempty"`
	Error   *FrameError     `json:"error,omitempty"`
}

type FrameError struct {
	Message string `json:"message"`
}

// EventFrame is a parsed event.
type EventFrame struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
	Seq     int             `json:"seq"`
}

// RequestFrame is an outgoing request.
type RequestFrame struct {
	Type   string      `json:"type"`
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// ResponseFrame is a parsed response.
type ResponseFrame struct {
	ID      string          `json:"id"`
	OK      bool            `json:"ok"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   *FrameError     `json:"error,omitempty"`
}

// ConnectChallenge is the payload of connect.challenge event.
type ConnectChallenge struct {
	Nonce string `json:"nonce"`
}

// ClientInfo identifies this client to the gateway.
type ClientInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Platform    string `json:"platform"`
	Mode        string `json:"mode"`
}

// AuthInfo holds authentication credentials.
type AuthInfo struct {
	Token    string `json:"token,omitempty"`
	Password string `json:"password,omitempty"`
}

// ConnectParams is sent as the params of the connect request.
type ConnectParams struct {
	MinProtocol int         `json:"minProtocol"`
	MaxProtocol int         `json:"maxProtocol"`
	Client      ClientInfo  `json:"client"`
	Caps        []string    `json:"caps"`
	Auth        AuthInfo    `json:"auth"`
	Role        string      `json:"role"`
	Scopes      []string    `json:"scopes"`
	Device      *DeviceInfo `json:"device,omitempty"`
}

// DeviceInfo is the signed device identity included in connect params.
type DeviceInfo struct {
	ID        string `json:"id"`
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
	SignedAt  int64  `json:"signedAt"`
	Nonce     string `json:"nonce"`
}

// ChatSendParams is sent when the user sends a message.
type ChatSendParams struct {
	SessionKey string `json:"sessionKey"`
	Message    string `json:"message"`
	Thinking   string `json:"thinking,omitempty"`
}

// ChatAbortParams aborts an active run.
type ChatAbortParams struct {
	SessionKey string `json:"sessionKey"`
	RunID      string `json:"runId,omitempty"`
}

// ChatEvent is the payload of a chat event.
type ChatEvent struct {
	RunID        string       `json:"runId"`
	SessionKey   string       `json:"sessionKey"`
	State        string       `json:"state"`
	Message      *ChatMessage `json:"message,omitempty"`
	ErrorMessage string       `json:"errorMessage,omitempty"`
}

// ChatMessage is a message within a chat event.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TickPayload is the payload of a tick event.
type TickPayload struct {
	Interval int `json:"interval,omitempty"`
}

// SessionInfo represents a session from sessions.list.
type SessionInfo struct {
	Key   string `json:"key"`
	Model string `json:"model,omitempty"`
	Agent string `json:"agent,omitempty"`
}

// StatusPayload is the response to a status request.
type StatusPayload struct {
	Sessions []SessionInfo `json:"sessions,omitempty"`
	Version  string        `json:"version,omitempty"`
}

// BTWPayload is a "by the way" informational message.
type BTWPayload struct {
	Message string `json:"message"`
}
