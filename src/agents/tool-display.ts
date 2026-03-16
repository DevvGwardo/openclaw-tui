// Minimal tool-display stub for standalone TUI.
// In the full codebase this loads JSON config; here we provide sensible defaults.

export type ToolDisplay = {
  name: string;
  emoji: string;
  title: string;
  label: string;
  verb?: string;
  detail?: string;
};

function defaultTitle(name: string): string {
  return name
    .replace(/([a-z])([A-Z])/g, "$1 $2")
    .replace(/[_-]+/g, " ")
    .replace(/^\w/, (c) => c.toUpperCase());
}

function formatArgsDetail(args: unknown): string | undefined {
  if (!args || typeof args !== "object") {
    return undefined;
  }
  const record = args as Record<string, unknown>;
  const parts: string[] = [];
  for (const [key, value] of Object.entries(record)) {
    if (value === undefined || value === null) {
      continue;
    }
    const display = typeof value === "string" ? value : JSON.stringify(value);
    const truncated = display.length > 60 ? `${display.slice(0, 57)}...` : display;
    parts.push(`${key}: ${truncated}`);
    if (parts.length >= 3) {
      break;
    }
  }
  return parts.length > 0 ? parts.join(", ") : undefined;
}

export function resolveToolDisplay(params: {
  name?: string;
  args?: unknown;
  meta?: string;
}): ToolDisplay {
  const name = params.name ?? "unknown";
  const title = defaultTitle(name);
  const detail = formatArgsDetail(params.args);
  return {
    name,
    emoji: "🧩",
    title,
    label: title,
    detail,
  };
}

export function formatToolDetail(display: ToolDisplay): string | undefined {
  return display.detail;
}

export function formatToolSummary(display: ToolDisplay): string {
  const detail = formatToolDetail(display);
  return detail
    ? `${display.emoji} ${display.label}: ${detail}`
    : `${display.emoji} ${display.label}`;
}
