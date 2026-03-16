# @openclaw/tui

Standalone terminal UI client for OpenClaw gateway.

## Install

```bash
npm install -g @openclaw/tui
```

## Usage

```bash
# Connect to a local gateway
openclaw-tui --url ws://127.0.0.1:18789

# Connect with authentication
openclaw-tui --url wss://gateway.example.com --token YOUR_TOKEN

# Or use environment variables
export OPENCLAW_GATEWAY_URL=ws://127.0.0.1:18789
export OPENCLAW_TOKEN=your-token
openclaw-tui
```

## CLI Options

- `--url` / `-u` — Gateway WebSocket URL (default: `ws://127.0.0.1:18789`)
- `--token` / `-t` — Authentication token
- `--password` / `-p` — Authentication password
- `--session` / `-s` — Initial session key
- `--message` / `-m` — Send a message on connect
- `--thinking` — Set thinking level
- `--deliver` — Enable message delivery to channels

## Keyboard Shortcuts

- `Ctrl+C` — Clear input / exit (press twice)
- `Ctrl+D` — Exit
- `Ctrl+L` — Switch model
- `Ctrl+G` — Switch agent
- `Ctrl+P` — Switch session
- `Ctrl+T` — Toggle thinking display
- `Ctrl+O` — Toggle tool output expansion
- `Escape` — Abort active run

## Slash Commands

- `/help` — Show commands
- `/status` — Gateway status
- `/model` — Switch model
- `/agent` — Switch agent
- `/session` — Switch session
- `/think <level>` — Set thinking level
- `/theme <name>` — Switch theme (ocean, amber, rose, forest)
- `/new` / `/reset` — Reset session
- `/exit` — Exit

## Development

```bash
pnpm install
pnpm exec tsc --noEmit   # Type check
pnpm build                # Build
node dist/index.js --url ws://127.0.0.1:18789
```
