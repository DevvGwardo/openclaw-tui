#!/usr/bin/env node
import { runTui } from "./tui/tui.js";
import type { TuiOptions } from "./tui/tui-types.js";

function parseArgs(argv: string[]): TuiOptions {
  const opts: TuiOptions = {};
  let i = 2; // skip node and script
  while (i < argv.length) {
    const arg = argv[i]!;
    const next = argv[i + 1];
    switch (arg) {
      case "--url":
      case "-u":
        opts.url = next;
        i += 2;
        break;
      case "--token":
      case "-t":
        opts.token = next;
        i += 2;
        break;
      case "--password":
      case "-p":
        opts.password = next;
        i += 2;
        break;
      case "--session":
      case "-s":
        opts.session = next;
        i += 2;
        break;
      case "--message":
      case "-m":
        opts.message = next;
        i += 2;
        break;
      case "--thinking":
        opts.thinking = next;
        i += 2;
        break;
      case "--deliver":
        opts.deliver = true;
        i += 1;
        break;
      case "--help":
      case "-h":
        console.log(`openclaw-tui — Terminal UI for OpenClaw gateway

Usage: openclaw-tui [options]

Options:
  --url, -u <url>         Gateway WebSocket URL (default: ws://127.0.0.1:18789)
  --token, -t <token>     Authentication token
  --password, -p <pass>   Authentication password
  --session, -s <key>     Initial session key
  --message, -m <msg>     Send message on connect
  --thinking <level>      Set thinking level
  --deliver               Enable delivery to channels
  --help, -h              Show this help

Environment variables:
  OPENCLAW_GATEWAY_URL    Gateway URL (fallback for --url)
  OPENCLAW_TOKEN          Auth token (fallback for --token)
  OPENCLAW_GATEWAY_PASSWORD  Auth password (fallback for --password)
`);
        process.exit(0);
        break;
      default:
        console.error(`Unknown option: ${arg}`);
        process.exit(1);
    }
  }

  // Fall back to environment variables
  if (!opts.url) {
    opts.url = process.env.OPENCLAW_GATEWAY_URL;
  }
  if (!opts.token) {
    opts.token = process.env.OPENCLAW_TOKEN ?? process.env.OPENCLAW_GATEWAY_TOKEN;
  }
  if (!opts.password) {
    opts.password = process.env.OPENCLAW_GATEWAY_PASSWORD;
  }

  return opts;
}

const opts = parseArgs(process.argv);
void runTui(opts);
