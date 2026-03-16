// Minimal config loader for standalone TUI.
// Returns empty config — the TUI connects via explicit URL/token args.
export type { OpenClawConfig } from "./types.js";
import type { OpenClawConfig } from "./types.js";

let cachedConfig: OpenClawConfig | null = null;

export function loadConfig(): OpenClawConfig {
  if (cachedConfig) {
    return cachedConfig;
  }
  cachedConfig = {};
  return cachedConfig;
}

export function resolveGatewayPort(config?: OpenClawConfig): number {
  return config?.gateway?.port ?? 18789;
}

export function resolveConfigPath(..._args: unknown[]): string {
  return "";
}

export function resolveStateDir(..._args: unknown[]): string {
  return "";
}
