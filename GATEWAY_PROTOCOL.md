# OpenClaw Gateway WebSocket Protocol

## Connection Flow
1. Client connects to gateway via WebSocket (e.g. ws://127.0.0.1:18789)
2. Gateway sends event frame: `{"type":"evt","event":"connect.challenge","payload":{"nonce":"<uuid>"}}`
3. Client sends request: `{"type":"req","id":"<uuid>","method":"connect","params":{...connectParams}}`
4. Gateway responds: `{"type":"res","id":"<uuid>","ok":true,"payload":{...helloOk}}`
5. After connect, gateway sends periodic tick events and chat events

## Frame Types

### Event Frame
```json
{"type":"evt","event":"<event-name>","payload":<any>,"seq":<number>}
```
Events: connect.challenge, tick, chat, agent, session.update, btw, etc.

### Request Frame  
```json
{"type":"req","id":"<uuid>","method":"<method>","params":<any>}
```
Methods: connect, chat.send, chat.abort, sessions.list, sessions.patch, agents.list, status, etc.

### Response Frame
```json
{"type":"res","id":"<uuid>","ok":true|false,"payload":<any>,"error":{"message":"..."}}
```

## Connect Params
```json
{
  "minProtocol": 14,
  "maxProtocol": 14,
  "client": {
    "id": "gateway-client",
    "displayName": "openclaw-tui",
    "version": "0.1.0",
    "platform": "windows",
    "mode": "backend"
  },
  "caps": [],
  "auth": {
    "token": "<optional>",
    "password": "<optional>"
  },
  "role": "operator",
  "scopes": ["operator.admin"]
}
```

## Chat Events
Chat events come as: `{"type":"evt","event":"chat","payload":{...chatEvent}}`

ChatEvent payload:
```json
{
  "runId": "<uuid>",
  "sessionKey": "<key>",
  "state": "delta|final|aborted|error",
  "message": { "role": "assistant", "content": "..." },
  "errorMessage": "..."
}
```

For streaming, multiple "delta" events arrive with partial content, then a "final" event.

## Sending Messages
```json
{"type":"req","id":"<uuid>","method":"chat.send","params":{
  "sessionKey": "agent:main:main",
  "message": "Hello!",
  "thinking": "adaptive"
}}
```

## Session Management
- sessions.list → list active sessions
- sessions.patch → change model, thinking level, etc.
- agents.list → list available agents

## Tick Watchdog
Gateway sends tick events at regular intervals (default 30s).
If no tick received for 2x the interval, assume connection is stale and reconnect.

## Reconnection
On disconnect, use exponential backoff: 1s → 2s → 4s → ... → 30s max.
