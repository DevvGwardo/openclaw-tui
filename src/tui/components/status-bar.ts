/**
 * Status bar component for the OpenClaw TUI.
 *
 * Renders a persistent bottom bar showing session, model, token usage,
 * cost, and thinking state with color-coded indicators.
 */

import { Container, Text } from "@mariozechner/pi-tui";
import { theme, getCurrentPalette } from "../theme/theme.js";

export interface StatusBarInfo {
  sessionLabel?: string;
  modelLabel?: string;
  inputTokens?: number | null;
  outputTokens?: number | null;
  totalTokens?: number | null;
  contextTokens?: number | null;
  thinkingLevel?: string;
  agentLabel?: string;
}

/** Format a token count as a compact string (e.g. "12.3k"). */
function compactTokens(n: number | null | undefined): string {
  if (n == null) return "–";
  if (n < 1000) return String(n);
  if (n < 100_000) return `${(n / 1000).toFixed(1)}k`;
  return `${Math.round(n / 1000)}k`;
}

/**
 * Color-code a token count based on how full the context window is.
 * Green (< 50%) → Yellow (50-80%) → Red (> 80%).
 */
function colorTokenUsage(total: number | null | undefined, context: number | null | undefined): string {
  const label = compactTokens(total);
  if (total == null || context == null || context === 0) {
    return theme.dim(label);
  }
  const pct = total / context;
  const p = getCurrentPalette();
  if (pct < 0.5) return theme.success(label);
  if (pct < 0.8) return theme.warning(label);
  return theme.error(label);
}

export class StatusBar extends Container {
  private bar: Text;
  private info: StatusBarInfo = {};

  constructor() {
    super();
    this.bar = new Text("", 1, 0);
    this.addChild(this.bar);
  }

  update(info: StatusBarInfo) {
    this.info = { ...this.info, ...info };
    this.refresh();
  }

  private refresh() {
    const i = this.info;
    const p = getCurrentPalette();

    const parts: string[] = [];

    if (i.agentLabel) {
      parts.push(theme.dim("agent ") + theme.accent(i.agentLabel));
    }

    if (i.sessionLabel) {
      parts.push(theme.dim("session ") + theme.accent(i.sessionLabel));
    }

    if (i.modelLabel) {
      parts.push(theme.accentSoft(i.modelLabel));
    }

    // Token usage with color coding
    const tokensUsed = colorTokenUsage(i.totalTokens, i.contextTokens);
    const ctxLabel = compactTokens(i.contextTokens);
    parts.push(theme.dim("tokens ") + tokensUsed + theme.dim(`/${ctxLabel}`));

    if (i.thinkingLevel && i.thinkingLevel !== "off") {
      parts.push(theme.tertiary(`think ${i.thinkingLevel}`));
    }

    const separator = theme.dim(" │ ");
    this.bar.setText(theme.dim("─ ") + parts.join(separator) + theme.dim(" ─"));
  }
}
