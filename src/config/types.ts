// Minimal config type stubs for standalone TUI.
// Only the fields actually read by the TUI are included.

export type OpenClawConfig = {
  session?: {
    scope?: string;
    mainKey?: string;
  };
  gateway?: {
    mode?: string;
    bind?: string;
    port?: number;
    auth?: {
      token?: string;
    };
    remote?: {
      url?: string;
      token?: string;
      password?: string;
    };
    tls?: {
      enabled?: boolean;
    };
  };
  agents?: {
    list?: Array<{
      id: string;
      name?: string;
      default?: boolean;
      model?: unknown;
      skills?: unknown;
      memorySearch?: unknown;
      humanDelay?: unknown;
      heartbeat?: unknown;
      identity?: unknown;
      groupChat?: unknown;
      subagents?: unknown;
      sandbox?: unknown;
      tools?: unknown;
    }>;
  };
  models?: {
    providers?: Record<
      string,
      {
        models?: Array<{
          id: string;
          cost?: {
            input: number;
            output: number;
            cacheRead: number;
            cacheWrite: number;
          };
        }>;
      }
    >;
  };
  commands?: Record<string, unknown>;
};
