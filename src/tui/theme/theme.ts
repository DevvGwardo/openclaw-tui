import type {
  EditorTheme,
  MarkdownTheme,
  SelectListTheme,
  SettingsListTheme,
} from "@mariozechner/pi-tui";
import chalk from "chalk";
import { highlight, supportsLanguage } from "cli-highlight";
import type { SearchableSelectListTheme } from "../components/searchable-select-list.js";
import { getActivePalette, type ThemePalette } from "./palettes.js";
import { createSyntaxTheme } from "./syntax-theme.js";

// ── Palette resolution ──────────────────────────────────────────
// The active palette is determined once at import time from
// OPENCLAW_TUI_THEME env var. It can be swapped at runtime via
// `setActivePalette()` for the /theme command.

let activePalette: ThemePalette = getActivePalette();

/**
 * Returns the currently active palette — useful for components that
 * need raw hex values (e.g. gradient helpers).
 */
export function getCurrentPalette(): ThemePalette {
  return activePalette;
}

/**
 * Hot-swap the active palette. Rebuilds all derived theme objects.
 * Used by the /theme command to cycle themes without restarting.
 */
export function setActivePalette(palette: ThemePalette): void {
  activePalette = palette;
  rebuildTheme();
}

// ── NO_COLOR support ─────────────────────────────────────────────
const noColor = Boolean(process.env.NO_COLOR);

// ── Color helpers ────────────────────────────────────────────────
const fg = (hex: string) => (text: string) => (noColor ? text : chalk.hex(hex)(text));
const bg = (hex: string) => (text: string) => (noColor ? text : chalk.bgHex(hex)(text));
const bold = (text: string) => (noColor ? text : chalk.bold(text));
const italic = (text: string) => (noColor ? text : chalk.italic(text));

// ── Syntax highlighting ──────────────────────────────────────────
let syntaxTheme = createSyntaxTheme(fg(activePalette.code));

function highlightCode(code: string, lang?: string): string[] {
  try {
    const language = lang && supportsLanguage(lang) ? lang : undefined;
    const highlighted = highlight(code, {
      language,
      theme: syntaxTheme,
      ignoreIllegals: true,
    });
    return highlighted.split("\n");
  } catch {
    return code.split("\n").map((line) => fg(activePalette.code)(line));
  }
}

// ── Main theme object ────────────────────────────────────────────
// All entries are functions (string) => string so consumers don't
// need to know about the palette directly.

export const theme = {
  fg: fg(activePalette.text),
  assistantText: (text: string) => text,
  dim: fg(activePalette.dim),
  accent: fg(activePalette.primary),
  accentSoft: fg(activePalette.secondary),
  tertiary: fg(activePalette.tertiary),
  success: fg(activePalette.success),
  error: fg(activePalette.error),
  info: fg(activePalette.info),
  warning: fg(activePalette.warning),
  header: (text: string) => bold(fg(activePalette.primary)(text)),
  system: fg(activePalette.systemText),
  userBg: bg(activePalette.userBg),
  userText: fg(activePalette.userText),
  toolTitle: fg(activePalette.toolTitle),
  toolOutput: fg(activePalette.toolOutput),
  toolPendingBg: bg(activePalette.toolPendingBg),
  toolSuccessBg: bg(activePalette.toolSuccessBg),
  toolErrorBg: bg(activePalette.toolErrorBg),
  border: fg(activePalette.border),
  surfaceBg: bg(activePalette.surface),
  surfaceAltBg: bg(activePalette.surfaceAlt),
  surfaceElevatedBg: bg(activePalette.surfaceElevated),
  bold,
  italic,
};

// ── Markdown theme ───────────────────────────────────────────────

export const markdownTheme: MarkdownTheme = {
  heading: (text) => bold(fg(activePalette.primary)(text)),
  link: (text) => fg(activePalette.link)(text),
  linkUrl: (text) => (noColor ? text : chalk.dim(text)),
  code: (text) => fg(activePalette.code)(text),
  codeBlock: (text) => fg(activePalette.code)(text),
  codeBlockBorder: (text) => fg(activePalette.codeBorder)(text),
  quote: (text) => fg(activePalette.quote)(text),
  quoteBorder: (text) => fg(activePalette.quoteBorder)(text),
  hr: (text) => fg(activePalette.border)(text),
  listBullet: (text) => fg(activePalette.secondary)(text),
  bold,
  italic,
  strikethrough: (text) => (noColor ? text : chalk.strikethrough(text)),
  underline: (text) => (noColor ? text : chalk.underline(text)),
  highlightCode,
};

// ── Select list themes ───────────────────────────────────────────

const baseSelectListTheme: SelectListTheme = {
  selectedPrefix: (text) => fg(activePalette.primary)(text),
  selectedText: (text) => bold(fg(activePalette.primary)(text)),
  description: (text) => fg(activePalette.dim)(text),
  scrollInfo: (text) => fg(activePalette.dim)(text),
  noMatch: (text) => fg(activePalette.dim)(text),
};

export const selectListTheme: SelectListTheme = baseSelectListTheme;

export const filterableSelectListTheme = {
  ...baseSelectListTheme,
  filterLabel: (text: string) => fg(activePalette.dim)(text),
};

export const settingsListTheme: SettingsListTheme = {
  label: (text, selected) =>
    selected ? bold(fg(activePalette.primary)(text)) : fg(activePalette.text)(text),
  value: (text, selected) =>
    selected ? fg(activePalette.secondary)(text) : fg(activePalette.dim)(text),
  description: (text) => fg(activePalette.systemText)(text),
  cursor: fg(activePalette.primary)("→ "),
  hint: (text) => fg(activePalette.dim)(text),
};

export const editorTheme: EditorTheme = {
  borderColor: (text) => fg(activePalette.border)(text),
  selectList: selectListTheme,
};

export const searchableSelectListTheme: SearchableSelectListTheme = {
  ...baseSelectListTheme,
  searchPrompt: (text) => fg(activePalette.secondary)(text),
  searchInput: (text) => fg(activePalette.text)(text),
  matchHighlight: (text) => bold(fg(activePalette.primary)(text)),
};

// ── Runtime rebuild ──────────────────────────────────────────────
// Called when the palette changes at runtime (/theme command).

function rebuildTheme() {
  const p = activePalette;
  syntaxTheme = createSyntaxTheme(fg(p.code));

  theme.fg = fg(p.text);
  theme.dim = fg(p.dim);
  theme.accent = fg(p.primary);
  theme.accentSoft = fg(p.secondary);
  theme.tertiary = fg(p.tertiary);
  theme.success = fg(p.success);
  theme.error = fg(p.error);
  theme.info = fg(p.info);
  theme.warning = fg(p.warning);
  theme.header = (text: string) => bold(fg(p.primary)(text));
  theme.system = fg(p.systemText);
  theme.userBg = bg(p.userBg);
  theme.userText = fg(p.userText);
  theme.toolTitle = fg(p.toolTitle);
  theme.toolOutput = fg(p.toolOutput);
  theme.toolPendingBg = bg(p.toolPendingBg);
  theme.toolSuccessBg = bg(p.toolSuccessBg);
  theme.toolErrorBg = bg(p.toolErrorBg);
  theme.border = fg(p.border);
  theme.surfaceBg = bg(p.surface);
  theme.surfaceAltBg = bg(p.surfaceAlt);
  theme.surfaceElevatedBg = bg(p.surfaceElevated);

  // Rebuild markdown theme entries
  markdownTheme.heading = (text) => bold(fg(p.primary)(text));
  markdownTheme.link = (text) => fg(p.link)(text);
  markdownTheme.code = (text) => fg(p.code)(text);
  markdownTheme.codeBlock = (text) => fg(p.code)(text);
  markdownTheme.codeBlockBorder = (text) => fg(p.codeBorder)(text);
  markdownTheme.quote = (text) => fg(p.quote)(text);
  markdownTheme.quoteBorder = (text) => fg(p.quoteBorder)(text);
  markdownTheme.hr = (text) => fg(p.border)(text);
  markdownTheme.listBullet = (text) => fg(p.secondary)(text);

  // Rebuild select list themes
  baseSelectListTheme.selectedPrefix = (text) => fg(p.primary)(text);
  baseSelectListTheme.selectedText = (text) => bold(fg(p.primary)(text));
  baseSelectListTheme.description = (text) => fg(p.dim)(text);
  baseSelectListTheme.scrollInfo = (text) => fg(p.dim)(text);
  baseSelectListTheme.noMatch = (text) => fg(p.dim)(text);

  settingsListTheme.label = (text, selected) =>
    selected ? bold(fg(p.primary)(text)) : fg(p.text)(text);
  settingsListTheme.value = (text, selected) =>
    selected ? fg(p.secondary)(text) : fg(p.dim)(text);
  settingsListTheme.description = (text) => fg(p.systemText)(text);
  settingsListTheme.cursor = fg(p.primary)("→ ");
  settingsListTheme.hint = (text) => fg(p.dim)(text);

  editorTheme.borderColor = (text) => fg(p.border)(text);

  searchableSelectListTheme.searchPrompt = (text) => fg(p.secondary)(text);
  searchableSelectListTheme.searchInput = (text) => fg(p.text)(text);
  searchableSelectListTheme.matchHighlight = (text) => bold(fg(p.primary)(text));
}
