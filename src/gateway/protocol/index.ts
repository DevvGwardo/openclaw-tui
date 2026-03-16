// Minimal gateway protocol stubs for standalone TUI.
// Contains only the types and validators actually used by the gateway client and TUI.

export const PROTOCOL_VERSION = 14;

export type ConnectParams = {
  minProtocol: number;
  maxProtocol: number;
  client: {
    id: string;
    displayName?: string;
    version: string;
    platform: string;
    deviceFamily?: string;
    mode: string;
    instanceId?: string;
  };
  caps?: string[];
  commands?: string[];
  permissions?: Record<string, boolean>;
  pathEnv?: string;
  auth?: {
    token?: string;
    deviceToken?: string;
    password?: string;
  };
  role?: string;
  scopes?: string[];
  device?: {
    id: string;
    publicKey: string;
    signature: string;
    signedAt: number;
    nonce: string;
  };
};

export type EventFrame = {
  type: "evt";
  event: string;
  payload?: unknown;
  seq?: number;
};

export type RequestFrame = {
  type: "req";
  id: string;
  method: string;
  params?: unknown;
};

export type ResponseFrame = {
  type: "res";
  id: string;
  ok: boolean;
  payload?: unknown;
  error?: { message?: string };
};

export type HelloOk = {
  ok: boolean;
  protocol?: number;
  auth?: {
    role?: string;
    scopes?: string[];
    deviceToken?: string;
  };
  policy?: {
    tickIntervalMs?: number;
  };
};

export type SessionsListParams = {
  limit?: number;
  activeMinutes?: number;
  includeGlobal?: boolean;
  includeUnknown?: boolean;
  includeDerivedTitles?: boolean;
  includeLastMessage?: boolean;
  agentId?: string;
};

export type SessionsPatchParams = {
  key: string;
  model?: string;
  thinkingLevel?: string;
  verboseLevel?: string;
  reasoningLevel?: string;
  elevatedLevel?: string;
  groupActivation?: string;
  responseUsage?: string | null;
};

export type SessionsPatchResult = {
  key?: string;
  entry?: Record<string, unknown>;
  resolved?: {
    model?: string;
    modelProvider?: string;
  };
};

// Simple runtime validators — accept any frame with the right type field.
export function validateEventFrame(data: unknown): data is EventFrame {
  if (!data || typeof data !== "object") {
    return false;
  }
  return (data as Record<string, unknown>).type === "evt";
}

export function validateRequestFrame(data: unknown): data is RequestFrame {
  if (!data || typeof data !== "object") {
    return false;
  }
  const record = data as Record<string, unknown>;
  return record.type === "req" && typeof record.id === "string" && typeof record.method === "string";
}
// Attach errors array to match AJV-style interface used by gateway client
validateRequestFrame.errors = null as unknown[] | null;

export function validateResponseFrame(
  data: unknown,
): data is ResponseFrame {
  if (!data || typeof data !== "object") {
    return false;
  }
  const record = data as Record<string, unknown>;
  return record.type === "res" && typeof record.id === "string";
}
