/**
 * Theme palette definitions for the OpenClaw TUI.
 *
 * Each palette provides a full set of semantic colors for rendering
 * the chat interface. The active palette is selected via the
 * OPENCLAW_TUI_THEME env var or the /theme slash command.
 */

export interface ThemePalette {
  /** Display name shown in /theme picker */
  name: string;

  // ── Accent colors ──────────────────────────────────────────────
  primary: string;
  secondary: string;
  tertiary: string;

  // ── Surface / background colors ────────────────────────────────
  surface: string;
  surfaceAlt: string;
  surfaceElevated: string;

  // ── On-surface contrast text ───────────────────────────────────
  text: string;
  onPrimary: string;
  onSecondary: string;

  // ── Semantic text ──────────────────────────────────────────────
  dim: string;
  systemText: string;
  border: string;

  // ── Status ─────────────────────────────────────────────────────
  info: string;
  warning: string;
  success: string;
  error: string;

  // ── User messages ──────────────────────────────────────────────
  userBg: string;
  userText: string;

  // ── Tool execution ─────────────────────────────────────────────
  toolPendingBg: string;
  toolSuccessBg: string;
  toolErrorBg: string;
  toolTitle: string;
  toolOutput: string;

  // ── Markdown / code ────────────────────────────────────────────
  quote: string;
  quoteBorder: string;
  code: string;
  codeBlock: string;
  codeBorder: string;
  link: string;
}

// ── Ocean (default) ──────────────────────────────────────────────
// Cool teal-blue primary with warm coral accents. Professional and calm.
export const ocean: ThemePalette = {
  name: "Ocean",
  primary: "#5EBED6",
  secondary: "#F0A870",
  tertiary: "#C490E4",
  surface: "#1A1E26",
  surfaceAlt: "#212630",
  surfaceElevated: "#2A303C",
  text: "#E0DDD5",
  onPrimary: "#1A1E26",
  onSecondary: "#1A1E26",
  dim: "#6B7280",
  systemText: "#8B95A8",
  border: "#3A4050",
  info: "#5EBED6",
  warning: "#F0A870",
  success: "#6CCB8E",
  error: "#F07068",
  userBg: "#1E2530",
  userText: "#E0DDD5",
  toolPendingBg: "#1A2530",
  toolSuccessBg: "#1A2D22",
  toolErrorBg: "#2D1A1A",
  toolTitle: "#5EBED6",
  toolOutput: "#D0CDC5",
  quote: "#8CC8FF",
  quoteBorder: "#3B5070",
  code: "#E8C878",
  codeBlock: "#181C24",
  codeBorder: "#303848",
  link: "#6CCB8E",
};

// ── Amber (legacy/warm) ─────────────────────────────────────────
// Warm gold tones matching the original theme. Cozy and familiar.
export const amber: ThemePalette = {
  name: "Amber",
  primary: "#F6C453",
  secondary: "#F2A65A",
  tertiary: "#E88B6A",
  surface: "#1C1F26",
  surfaceAlt: "#22252E",
  surfaceElevated: "#2B2F38",
  text: "#E8E3D5",
  onPrimary: "#1C1F26",
  onSecondary: "#1C1F26",
  dim: "#7B7F87",
  systemText: "#9BA3B2",
  border: "#3C414B",
  info: "#8CC8FF",
  warning: "#F6C453",
  success: "#7DD3A5",
  error: "#F97066",
  userBg: "#2B2F36",
  userText: "#F3EEE0",
  toolPendingBg: "#1F2A2F",
  toolSuccessBg: "#1E2D23",
  toolErrorBg: "#2F1F1F",
  toolTitle: "#F6C453",
  toolOutput: "#E1DACB",
  quote: "#8CC8FF",
  quoteBorder: "#3B4D6B",
  code: "#F0C987",
  codeBlock: "#1E232A",
  codeBorder: "#343A45",
  link: "#7DD3A5",
};

// ── Rose ─────────────────────────────────────────────────────────
// Soft pink-magenta primary with lavender accents. Elegant and modern.
export const rose: ThemePalette = {
  name: "Rose",
  primary: "#E87CA0",
  secondary: "#C8A0E0",
  tertiary: "#80C0D8",
  surface: "#1C1A22",
  surfaceAlt: "#221F2A",
  surfaceElevated: "#2C2834",
  text: "#E4DDE8",
  onPrimary: "#1C1A22",
  onSecondary: "#1C1A22",
  dim: "#7A7084",
  systemText: "#9990A8",
  border: "#403850",
  info: "#80C0D8",
  warning: "#E8C070",
  success: "#78CC90",
  error: "#E86868",
  userBg: "#261E2C",
  userText: "#E4DDE8",
  toolPendingBg: "#201A2A",
  toolSuccessBg: "#1A2820",
  toolErrorBg: "#2C1A1A",
  toolTitle: "#E87CA0",
  toolOutput: "#D8D0DE",
  quote: "#C8A0E0",
  quoteBorder: "#503870",
  code: "#E8C070",
  codeBlock: "#1A1820",
  codeBorder: "#38304A",
  link: "#78CC90",
};

// ── Forest ───────────────────────────────────────────────────────
// Rich green primary with earth-tone accents. Natural and grounded.
export const forest: ThemePalette = {
  name: "Forest",
  primary: "#6CC890",
  secondary: "#D8B870",
  tertiary: "#70A8D0",
  surface: "#1A2018",
  surfaceAlt: "#1E261C",
  surfaceElevated: "#263024",
  text: "#D8E0D0",
  onPrimary: "#1A2018",
  onSecondary: "#1A2018",
  dim: "#6A7868",
  systemText: "#8A9888",
  border: "#354830",
  info: "#70A8D0",
  warning: "#D8B870",
  success: "#6CC890",
  error: "#D87060",
  userBg: "#1E2A1E",
  userText: "#D8E0D0",
  toolPendingBg: "#1A261E",
  toolSuccessBg: "#1A2E1E",
  toolErrorBg: "#2A1A18",
  toolTitle: "#6CC890",
  toolOutput: "#C8D0C0",
  quote: "#70A8D0",
  quoteBorder: "#305040",
  code: "#D8B870",
  codeBlock: "#182018",
  codeBorder: "#2E3E28",
  link: "#70A8D0",
};

// ── Palette registry ─────────────────────────────────────────────

export const palettes: Record<string, ThemePalette> = {
  ocean,
  amber,
  rose,
  forest,
};

export const paletteNames = Object.keys(palettes) as string[];

export const DEFAULT_PALETTE_NAME = "ocean";

/**
 * Resolve the active palette name from environment or fallback.
 */
export function resolveThemeName(envOverride?: string): string {
  const name = (envOverride ?? process.env.OPENCLAW_TUI_THEME ?? "").trim().toLowerCase();
  if (name && palettes[name]) {
    return name;
  }
  return DEFAULT_PALETTE_NAME;
}

/**
 * Get the active palette, respecting OPENCLAW_TUI_THEME env var.
 */
export function getActivePalette(envOverride?: string): ThemePalette {
  return palettes[resolveThemeName(envOverride)] ?? ocean;
}

/**
 * WCAG AA contrast ratio check between two hex colors.
 * Returns true if contrast ratio >= 4.5:1 (normal text) or >= 3:1 (large text).
 */
export function checkWcagAA(
  foreground: string,
  background: string,
  largeText = false,
): { passes: boolean; ratio: number } {
  const fgLum = relativeLuminance(hexToRgb(foreground));
  const bgLum = relativeLuminance(hexToRgb(background));
  const lighter = Math.max(fgLum, bgLum);
  const darker = Math.min(fgLum, bgLum);
  const ratio = (lighter + 0.05) / (darker + 0.05);
  const threshold = largeText ? 3 : 4.5;
  return { passes: ratio >= threshold, ratio: Math.round(ratio * 100) / 100 };
}

function hexToRgb(hex: string): [number, number, number] {
  const h = hex.replace("#", "");
  return [
    parseInt(h.substring(0, 2), 16) / 255,
    parseInt(h.substring(2, 4), 16) / 255,
    parseInt(h.substring(4, 6), 16) / 255,
  ];
}

function relativeLuminance(rgb: [number, number, number]): number {
  const [r, g, b] = rgb.map((c) =>
    c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4),
  );
  return 0.2126 * r! + 0.7152 * g! + 0.0722 * b!;
}
