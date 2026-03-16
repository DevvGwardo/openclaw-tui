// Minimal gateway call stubs for standalone TUI.
// Only the functions used by gateway-chat.ts are included.

import type { OpenClawConfig } from "../config/config.js";

export type ExplicitGatewayAuth = {
  token?: string;
  password?: string;
};

export type GatewayConnectionDetails = {
  url: string;
  urlSource: string;
  bindDetail?: string;
  remoteFallbackNote?: string;
  message: string;
};

export function resolveExplicitGatewayAuth(opts?: ExplicitGatewayAuth): ExplicitGatewayAuth {
  const token =
    typeof opts?.token === "string" && opts.token.trim().length > 0 ? opts.token.trim() : undefined;
  const password =
    typeof opts?.password === "string" && opts.password.trim().length > 0
      ? opts.password.trim()
      : undefined;
  return { token, password };
}

export function ensureExplicitGatewayAuth(_params: {
  urlOverride?: string;
  urlOverrideSource?: "cli" | "env";
  explicitAuth?: ExplicitGatewayAuth;
  resolvedAuth?: ExplicitGatewayAuth;
  errorHint: string;
  configPath?: string;
}): void {
  // In standalone mode, we don't enforce the URL-override auth requirement.
  // The user explicitly chose to connect to a URL via CLI args.
}

export function buildGatewayConnectionDetails(
  options: {
    config?: OpenClawConfig;
    url?: string;
    configPath?: string;
    urlSource?: "cli" | "env";
  } = {},
): GatewayConnectionDetails {
  const envUrl =
    process.env.OPENCLAW_GATEWAY_URL?.trim() || undefined;
  const url = options.url ?? envUrl ?? "ws://127.0.0.1:18789";
  return {
    url,
    urlSource: options.url ? "cli" : envUrl ? "env" : "default",
    message: `Connecting to ${url}`,
  };
}
