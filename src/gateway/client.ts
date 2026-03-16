// Simplified gateway client for standalone TUI.
// Removed: device auth, TLS fingerprint pinning, device pairing.
// Kept: WebSocket protocol, token/password auth, reconnection, tick watchdog.

import { randomUUID } from "node:crypto";
import { WebSocket } from "ws";
import {
  GATEWAY_CLIENT_MODES,
  GATEWAY_CLIENT_NAMES,
  type GatewayClientMode,
  type GatewayClientName,
} from "../utils/message-channel.js";
import { VERSION } from "../version.js";
import {
  type ConnectParams,
  type EventFrame,
  type HelloOk,
  PROTOCOL_VERSION,
  type RequestFrame,
  validateEventFrame,
  validateRequestFrame,
  validateResponseFrame,
} from "./protocol/index.js";

type Pending = {
  resolve: (value: unknown) => void;
  reject: (err: unknown) => void;
  expectFinal: boolean;
};

export type GatewayClientOptions = {
  url?: string;
  connectDelayMs?: number;
  tickWatchMinIntervalMs?: number;
  token?: string;
  deviceToken?: string;
  password?: string;
  instanceId?: string;
  clientName?: GatewayClientName;
  clientDisplayName?: string;
  clientVersion?: string;
  platform?: string;
  deviceFamily?: string;
  mode?: GatewayClientMode;
  role?: string;
  scopes?: string[];
  caps?: string[];
  commands?: string[];
  permissions?: Record<string, boolean>;
  pathEnv?: string;
  deviceIdentity?: unknown;
  minProtocol?: number;
  maxProtocol?: number;
  tlsFingerprint?: string;
  onEvent?: (evt: EventFrame) => void;
  onHelloOk?: (hello: HelloOk) => void;
  onConnectError?: (err: Error) => void;
  onClose?: (code: number, reason: string) => void;
  onGap?: (info: { expected: number; received: number }) => void;
};

export const GATEWAY_CLOSE_CODE_HINTS: Readonly<Record<number, string>> = {
  1000: "normal closure",
  1006: "abnormal closure (no close frame)",
  1008: "policy violation",
  1012: "service restart",
};

export function describeGatewayCloseCode(code: number): string | undefined {
  return GATEWAY_CLOSE_CODE_HINTS[code];
}

function rawDataToString(data: unknown): string {
  if (typeof data === "string") {
    return data;
  }
  if (Buffer.isBuffer(data)) {
    return data.toString("utf8");
  }
  if (data instanceof ArrayBuffer) {
    return Buffer.from(data).toString("utf8");
  }
  if (Array.isArray(data)) {
    return Buffer.concat(data.map((chunk) => Buffer.from(chunk))).toString("utf8");
  }
  return String(data);
}

export class GatewayClient {
  private ws: WebSocket | null = null;
  private opts: GatewayClientOptions;
  private pending = new Map<string, Pending>();
  private backoffMs = 1000;
  private closed = false;
  private lastSeq: number | null = null;
  private connectNonce: string | null = null;
  private connectSent = false;
  private connectTimer: NodeJS.Timeout | null = null;
  private lastTick: number | null = null;
  private tickIntervalMs = 30_000;
  private tickTimer: NodeJS.Timeout | null = null;

  constructor(opts: GatewayClientOptions) {
    this.opts = { ...opts };
  }

  start() {
    if (this.closed) {
      return;
    }
    const url = this.opts.url ?? "ws://127.0.0.1:18789";
    this.ws = new WebSocket(url, {
      maxPayload: 25 * 1024 * 1024,
    });

    this.ws.on("open", () => {
      this.queueConnect();
    });
    this.ws.on("message", (data) => this.handleMessage(rawDataToString(data)));
    this.ws.on("close", (code, reason) => {
      const reasonText = rawDataToString(reason);
      this.ws = null;
      this.flushPendingErrors(new Error(`gateway closed (${code}): ${reasonText}`));
      this.scheduleReconnect();
      this.opts.onClose?.(code, reasonText);
    });
    this.ws.on("error", (err) => {
      if (!this.connectSent) {
        this.opts.onConnectError?.(err instanceof Error ? err : new Error(String(err)));
      }
    });
  }

  stop() {
    this.closed = true;
    if (this.tickTimer) {
      clearInterval(this.tickTimer);
      this.tickTimer = null;
    }
    this.ws?.close();
    this.ws = null;
    this.flushPendingErrors(new Error("gateway client stopped"));
  }

  private sendConnect() {
    if (this.connectSent) {
      return;
    }
    const nonce = this.connectNonce?.trim() ?? "";
    if (!nonce) {
      this.opts.onConnectError?.(new Error("gateway connect challenge missing nonce"));
      this.ws?.close(1008, "connect challenge missing nonce");
      return;
    }
    this.connectSent = true;
    if (this.connectTimer) {
      clearTimeout(this.connectTimer);
      this.connectTimer = null;
    }
    const role = this.opts.role ?? "operator";
    const authToken = this.opts.token?.trim() || undefined;
    const authPassword = this.opts.password?.trim() || undefined;
    const auth =
      authToken || authPassword
        ? { token: authToken, password: authPassword }
        : undefined;
    const scopes = this.opts.scopes ?? ["operator.admin"];
    const platform = this.opts.platform ?? process.platform;
    const params: ConnectParams = {
      minProtocol: this.opts.minProtocol ?? PROTOCOL_VERSION,
      maxProtocol: this.opts.maxProtocol ?? PROTOCOL_VERSION,
      client: {
        id: this.opts.clientName ?? GATEWAY_CLIENT_NAMES.GATEWAY_CLIENT,
        displayName: this.opts.clientDisplayName,
        version: this.opts.clientVersion ?? VERSION,
        platform,
        deviceFamily: this.opts.deviceFamily,
        mode: this.opts.mode ?? GATEWAY_CLIENT_MODES.BACKEND,
        instanceId: this.opts.instanceId,
      },
      caps: Array.isArray(this.opts.caps) ? this.opts.caps : [],
      commands: Array.isArray(this.opts.commands) ? this.opts.commands : undefined,
      permissions:
        this.opts.permissions && typeof this.opts.permissions === "object"
          ? this.opts.permissions
          : undefined,
      pathEnv: this.opts.pathEnv,
      auth,
      role,
      scopes,
    };

    void this.request<HelloOk>("connect", params)
      .then((helloOk) => {
        this.backoffMs = 1000;
        this.tickIntervalMs =
          typeof helloOk.policy?.tickIntervalMs === "number"
            ? helloOk.policy.tickIntervalMs
            : 30_000;
        this.lastTick = Date.now();
        this.startTickWatch();
        this.opts.onHelloOk?.(helloOk);
      })
      .catch((err) => {
        this.opts.onConnectError?.(err instanceof Error ? err : new Error(String(err)));
        this.ws?.close(1008, "connect failed");
      });
  }

  private handleMessage(raw: string) {
    try {
      const parsed = JSON.parse(raw);
      if (validateEventFrame(parsed)) {
        const evt = parsed;
        if (evt.event === "connect.challenge") {
          const payload = evt.payload as { nonce?: unknown } | undefined;
          const nonce = payload && typeof payload.nonce === "string" ? payload.nonce : null;
          if (!nonce || nonce.trim().length === 0) {
            this.opts.onConnectError?.(new Error("gateway connect challenge missing nonce"));
            this.ws?.close(1008, "connect challenge missing nonce");
            return;
          }
          this.connectNonce = nonce.trim();
          this.sendConnect();
          return;
        }
        const seq = typeof evt.seq === "number" ? evt.seq : null;
        if (seq !== null) {
          if (this.lastSeq !== null && seq > this.lastSeq + 1) {
            this.opts.onGap?.({ expected: this.lastSeq + 1, received: seq });
          }
          this.lastSeq = seq;
        }
        if (evt.event === "tick") {
          this.lastTick = Date.now();
        }
        this.opts.onEvent?.(evt);
        return;
      }
      if (validateResponseFrame(parsed)) {
        const pending = this.pending.get(parsed.id);
        if (!pending) {
          return;
        }
        const payload = parsed.payload as { status?: unknown } | undefined;
        const status = payload?.status;
        if (pending.expectFinal && status === "accepted") {
          return;
        }
        this.pending.delete(parsed.id);
        if (parsed.ok) {
          pending.resolve(parsed.payload);
        } else {
          pending.reject(new Error(parsed.error?.message ?? "unknown error"));
        }
      }
    } catch {
      // parse error — ignore malformed frames
    }
  }

  private queueConnect() {
    this.connectNonce = null;
    this.connectSent = false;
    const rawConnectDelayMs = this.opts.connectDelayMs;
    const connectChallengeTimeoutMs =
      typeof rawConnectDelayMs === "number" && Number.isFinite(rawConnectDelayMs)
        ? Math.max(250, Math.min(10_000, rawConnectDelayMs))
        : 2_000;
    if (this.connectTimer) {
      clearTimeout(this.connectTimer);
    }
    this.connectTimer = setTimeout(() => {
      if (this.connectSent || this.ws?.readyState !== WebSocket.OPEN) {
        return;
      }
      this.opts.onConnectError?.(new Error("gateway connect challenge timeout"));
      this.ws?.close(1008, "connect challenge timeout");
    }, connectChallengeTimeoutMs);
  }

  private scheduleReconnect() {
    if (this.closed) {
      return;
    }
    if (this.tickTimer) {
      clearInterval(this.tickTimer);
      this.tickTimer = null;
    }
    const delay = this.backoffMs;
    this.backoffMs = Math.min(this.backoffMs * 2, 30_000);
    setTimeout(() => this.start(), delay).unref();
  }

  private flushPendingErrors(err: Error) {
    for (const [, p] of this.pending) {
      p.reject(err);
    }
    this.pending.clear();
  }

  private startTickWatch() {
    if (this.tickTimer) {
      clearInterval(this.tickTimer);
    }
    const rawMinInterval = this.opts.tickWatchMinIntervalMs;
    const minInterval =
      typeof rawMinInterval === "number" && Number.isFinite(rawMinInterval)
        ? Math.max(1, Math.min(30_000, rawMinInterval))
        : 1000;
    const interval = Math.max(this.tickIntervalMs, minInterval);
    this.tickTimer = setInterval(() => {
      if (this.closed) {
        return;
      }
      if (!this.lastTick) {
        return;
      }
      const gap = Date.now() - this.lastTick;
      if (gap > this.tickIntervalMs * 2) {
        this.ws?.close(4000, "tick timeout");
      }
    }, interval);
  }

  async request<T = Record<string, unknown>>(
    method: string,
    params?: unknown,
    opts?: { expectFinal?: boolean },
  ): Promise<T> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error("gateway not connected");
    }
    const id = randomUUID();
    const frame: RequestFrame = { type: "req", id, method, params };
    if (!validateRequestFrame(frame)) {
      throw new Error(
        `invalid request frame: ${JSON.stringify(validateRequestFrame.errors, null, 2)}`,
      );
    }
    const expectFinal = opts?.expectFinal === true;
    const p = new Promise<T>((resolve, reject) => {
      this.pending.set(id, {
        resolve: (value) => resolve(value as T),
        reject,
        expectFinal,
      });
    });
    this.ws.send(JSON.stringify(frame));
    return p;
  }
}
