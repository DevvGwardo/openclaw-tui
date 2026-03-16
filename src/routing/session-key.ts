// Adapted from the full codebase — inlined dependencies to avoid deep imports.

// Inlined from channels/chat-type.ts
export type ChatType = "direct" | "group" | "channel";

// Inlined from sessions/session-key-utils.ts
export type ParsedAgentSessionKey = {
  agentId: string;
  rest: string;
};

export function parseAgentSessionKey(
  sessionKey: string | undefined | null,
): ParsedAgentSessionKey | null {
  const raw = (sessionKey ?? "").trim().toLowerCase();
  if (!raw) {
    return null;
  }
  const parts = raw.split(":").filter(Boolean);
  if (parts.length < 3) {
    return null;
  }
  if (parts[0] !== "agent") {
    return null;
  }
  const agentId = parts[1]?.trim();
  const rest = parts.slice(2).join(":");
  if (!agentId || !rest) {
    return null;
  }
  return { agentId, rest };
}

export function getSubagentDepth(sessionKey: string | undefined | null): number {
  const raw = (sessionKey ?? "").trim().toLowerCase();
  if (!raw) {
    return 0;
  }
  return raw.split(":subagent:").length - 1;
}

export function isCronSessionKey(sessionKey: string | undefined | null): boolean {
  const parsed = parseAgentSessionKey(sessionKey);
  if (!parsed) {
    return false;
  }
  return parsed.rest.toLowerCase().startsWith("cron:");
}

export function isAcpSessionKey(sessionKey: string | undefined | null): boolean {
  const raw = (sessionKey ?? "").trim();
  if (!raw) {
    return false;
  }
  const normalized = raw.toLowerCase();
  if (normalized.startsWith("acp:")) {
    return true;
  }
  const parsed = parseAgentSessionKey(raw);
  return Boolean((parsed?.rest ?? "").toLowerCase().startsWith("acp:"));
}

export function isSubagentSessionKey(sessionKey: string | undefined | null): boolean {
  const raw = (sessionKey ?? "").trim();
  if (!raw) {
    return false;
  }
  if (raw.toLowerCase().startsWith("subagent:")) {
    return true;
  }
  const parsed = parseAgentSessionKey(raw);
  return Boolean((parsed?.rest ?? "").toLowerCase().startsWith("subagent:"));
}

// Inlined from routing/account-id.ts
export const DEFAULT_ACCOUNT_ID = "default";

const VALID_ACCOUNT_ID_RE = /^[a-z0-9][a-z0-9_-]{0,63}$/i;
const INVALID_CHARS_RE = /[^a-z0-9_-]+/g;
const LEADING_DASH_RE = /^-+/;
const TRAILING_DASH_RE = /-+$/;

export function normalizeAccountId(value: string | undefined | null): string {
  const trimmed = (value ?? "").trim();
  if (!trimmed) {
    return DEFAULT_ACCOUNT_ID;
  }
  if (VALID_ACCOUNT_ID_RE.test(trimmed)) {
    return trimmed.toLowerCase();
  }
  const canonical = trimmed
    .toLowerCase()
    .replace(INVALID_CHARS_RE, "-")
    .replace(LEADING_DASH_RE, "")
    .replace(TRAILING_DASH_RE, "")
    .slice(0, 64);
  return canonical || DEFAULT_ACCOUNT_ID;
}

export function normalizeOptionalAccountId(value: string | undefined | null): string | undefined {
  const trimmed = (value ?? "").trim();
  if (!trimmed) {
    return undefined;
  }
  if (VALID_ACCOUNT_ID_RE.test(trimmed)) {
    return trimmed.toLowerCase();
  }
  const canonical = trimmed
    .toLowerCase()
    .replace(INVALID_CHARS_RE, "-")
    .replace(LEADING_DASH_RE, "")
    .replace(TRAILING_DASH_RE, "")
    .slice(0, 64);
  return canonical || undefined;
}

// Main exports
export const DEFAULT_AGENT_ID = "main";
export const DEFAULT_MAIN_KEY = "main";
export type SessionKeyShape = "missing" | "agent" | "legacy_or_alias" | "malformed_agent";

const VALID_ID_RE = /^[a-z0-9][a-z0-9_-]{0,63}$/i;
const AGENT_INVALID_CHARS_RE = /[^a-z0-9_-]+/g;
const AGENT_LEADING_DASH_RE = /^-+/;
const AGENT_TRAILING_DASH_RE = /-+$/;

export function scopedHeartbeatWakeOptions<T extends object>(
  sessionKey: string,
  wakeOptions: T,
): T | (T & { sessionKey: string }) {
  return parseAgentSessionKey(sessionKey) ? { ...wakeOptions, sessionKey } : wakeOptions;
}

export function normalizeMainKey(value: string | undefined | null): string {
  const trimmed = (value ?? "").trim();
  return trimmed ? trimmed.toLowerCase() : DEFAULT_MAIN_KEY;
}

export function toAgentRequestSessionKey(storeKey: string | undefined | null): string | undefined {
  const raw = (storeKey ?? "").trim();
  if (!raw) {
    return undefined;
  }
  return parseAgentSessionKey(raw)?.rest ?? raw;
}

export function toAgentStoreSessionKey(params: {
  agentId: string;
  requestKey: string | undefined | null;
  mainKey?: string | undefined;
}): string {
  const raw = (params.requestKey ?? "").trim();
  if (!raw || raw.toLowerCase() === DEFAULT_MAIN_KEY) {
    return buildAgentMainSessionKey({ agentId: params.agentId, mainKey: params.mainKey });
  }
  const parsed = parseAgentSessionKey(raw);
  if (parsed) {
    return `agent:${parsed.agentId}:${parsed.rest}`;
  }
  const lowered = raw.toLowerCase();
  if (lowered.startsWith("agent:")) {
    return lowered;
  }
  return `agent:${normalizeAgentId(params.agentId)}:${lowered}`;
}

export function resolveAgentIdFromSessionKey(sessionKey: string | undefined | null): string {
  const parsed = parseAgentSessionKey(sessionKey);
  return normalizeAgentId(parsed?.agentId ?? DEFAULT_AGENT_ID);
}

export function classifySessionKeyShape(sessionKey: string | undefined | null): SessionKeyShape {
  const raw = (sessionKey ?? "").trim();
  if (!raw) {
    return "missing";
  }
  if (parseAgentSessionKey(raw)) {
    return "agent";
  }
  return raw.toLowerCase().startsWith("agent:") ? "malformed_agent" : "legacy_or_alias";
}

export function normalizeAgentId(value: string | undefined | null): string {
  const trimmed = (value ?? "").trim();
  if (!trimmed) {
    return DEFAULT_AGENT_ID;
  }
  if (VALID_ID_RE.test(trimmed)) {
    return trimmed.toLowerCase();
  }
  return (
    trimmed
      .toLowerCase()
      .replace(AGENT_INVALID_CHARS_RE, "-")
      .replace(AGENT_LEADING_DASH_RE, "")
      .replace(AGENT_TRAILING_DASH_RE, "")
      .slice(0, 64) || DEFAULT_AGENT_ID
  );
}

export function isValidAgentId(value: string | undefined | null): boolean {
  const trimmed = (value ?? "").trim();
  return Boolean(trimmed) && VALID_ID_RE.test(trimmed);
}

export function sanitizeAgentId(value: string | undefined | null): string {
  return normalizeAgentId(value);
}

export function buildAgentMainSessionKey(params: {
  agentId: string;
  mainKey?: string | undefined;
}): string {
  const agentId = normalizeAgentId(params.agentId);
  const mainKey = normalizeMainKey(params.mainKey);
  return `agent:${agentId}:${mainKey}`;
}

export function buildAgentPeerSessionKey(params: {
  agentId: string;
  mainKey?: string | undefined;
  channel: string;
  accountId?: string | null;
  peerKind?: ChatType | null;
  peerId?: string | null;
  identityLinks?: Record<string, string[]>;
  dmScope?: "main" | "per-peer" | "per-channel-peer" | "per-account-channel-peer";
}): string {
  const peerKind = params.peerKind ?? "direct";
  if (peerKind === "direct") {
    const dmScope = params.dmScope ?? "main";
    let peerId = (params.peerId ?? "").trim().toLowerCase();
    if (dmScope === "per-account-channel-peer" && peerId) {
      const channel = (params.channel ?? "").trim().toLowerCase() || "unknown";
      const accountId = normalizeAccountId(params.accountId);
      return `agent:${normalizeAgentId(params.agentId)}:${channel}:${accountId}:direct:${peerId}`;
    }
    if (dmScope === "per-channel-peer" && peerId) {
      const channel = (params.channel ?? "").trim().toLowerCase() || "unknown";
      return `agent:${normalizeAgentId(params.agentId)}:${channel}:direct:${peerId}`;
    }
    if (dmScope === "per-peer" && peerId) {
      return `agent:${normalizeAgentId(params.agentId)}:direct:${peerId}`;
    }
    return buildAgentMainSessionKey({
      agentId: params.agentId,
      mainKey: params.mainKey,
    });
  }
  const channel = (params.channel ?? "").trim().toLowerCase() || "unknown";
  const peerId = ((params.peerId ?? "").trim() || "unknown").toLowerCase();
  return `agent:${normalizeAgentId(params.agentId)}:${channel}:${peerKind}:${peerId}`;
}

export function buildGroupHistoryKey(params: {
  channel: string;
  accountId?: string | null;
  peerKind: "group" | "channel";
  peerId: string;
}): string {
  const channel = (params.channel ?? "").trim().toLowerCase() || "unknown";
  const accountId = normalizeAccountId(params.accountId);
  const peerId = params.peerId.trim().toLowerCase() || "unknown";
  return `${channel}:${accountId}:${params.peerKind}:${peerId}`;
}

export function resolveThreadSessionKeys(params: {
  baseSessionKey: string;
  threadId?: string | null;
  parentSessionKey?: string;
  useSuffix?: boolean;
  normalizeThreadId?: (threadId: string) => string;
}): { sessionKey: string; parentSessionKey?: string } {
  const threadId = (params.threadId ?? "").trim();
  if (!threadId) {
    return { sessionKey: params.baseSessionKey, parentSessionKey: undefined };
  }
  const normalizedThreadId = (params.normalizeThreadId ?? ((value: string) => value.toLowerCase()))(
    threadId,
  );
  const useSuffix = params.useSuffix ?? true;
  const sessionKey = useSuffix
    ? `${params.baseSessionKey}:thread:${normalizedThreadId}`
    : params.baseSessionKey;
  return { sessionKey, parentSessionKey: params.parentSessionKey };
}
