// Minimal stub — only formatRawAssistantErrorForUi is used by the TUI.

const HTTP_STATUS_PREFIX_RE = /^(\d{3})\s+([\s\S]*)$/;

function parseApiErrorInfo(text: string): { message?: string; httpCode?: string; type?: string; requestId?: string } | null {
  try {
    const jsonStart = text.indexOf("{");
    if (jsonStart < 0) {
      return null;
    }
    const parsed = JSON.parse(text.slice(jsonStart)) as Record<string, unknown>;
    const error = (parsed.error ?? parsed) as Record<string, unknown>;
    const message = typeof error.message === "string" ? error.message : undefined;
    const type = typeof error.type === "string" ? error.type : undefined;
    const requestId = typeof error.request_id === "string" ? error.request_id : undefined;
    if (!message) {
      return null;
    }
    return { message, type, requestId };
  } catch {
    return null;
  }
}

export function formatRawAssistantErrorForUi(raw?: string): string {
  const trimmed = (raw ?? "").trim();
  if (!trimmed) {
    return "LLM request failed with an unknown error.";
  }

  const httpMatch = trimmed.match(HTTP_STATUS_PREFIX_RE);
  if (httpMatch) {
    const rest = httpMatch[2]!.trim();
    if (!rest.startsWith("{")) {
      return `HTTP ${httpMatch[1]}: ${rest}`;
    }
  }

  const info = parseApiErrorInfo(trimmed);
  if (info?.message) {
    const prefix = info.httpCode ? `HTTP ${info.httpCode}` : "LLM error";
    const type = info.type ? ` ${info.type}` : "";
    const requestId = info.requestId ? ` (request_id: ${info.requestId})` : "";
    return `${prefix}${type}: ${info.message}${requestId}`;
  }

  return trimmed.length > 600 ? `${trimmed.slice(0, 600)}…` : trimmed;
}
