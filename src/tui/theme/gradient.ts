/**
 * Gradient text utilities for the OpenClaw TUI.
 *
 * Interpolates between two hex colors across characters to produce
 * smooth color transitions in terminal output.
 */

import chalk from "chalk";

/**
 * Parse a hex color string into [r, g, b] (0–255).
 */
function hexToRgb(hex: string): [number, number, number] {
  const h = hex.replace("#", "");
  return [
    parseInt(h.substring(0, 2), 16),
    parseInt(h.substring(2, 4), 16),
    parseInt(h.substring(4, 6), 16),
  ];
}

/**
 * Linearly interpolate between two values.
 */
function lerp(a: number, b: number, t: number): number {
  return Math.round(a + (b - a) * t);
}

/**
 * Render text with a smooth gradient between two hex colors.
 *
 * Each visible character gets a unique color interpolated between
 * `fromHex` and `toHex`. ANSI escape sequences in the input are
 * not expected — pass plain text.
 *
 * @example
 * gradient("OpenClaw", "#5EBED6", "#E87CA0")
 */
export function gradient(text: string, fromHex: string, toHex: string): string {
  if (!text) return "";
  const from = hexToRgb(fromHex);
  const to = hexToRgb(toHex);
  const len = text.length;
  if (len === 1) return chalk.rgb(from[0], from[1], from[2])(text);

  let result = "";
  for (let i = 0; i < len; i++) {
    const t = i / (len - 1);
    const r = lerp(from[0], to[0], t);
    const g = lerp(from[1], to[1], t);
    const b = lerp(from[2], to[2], t);
    result += chalk.rgb(r, g, b)(text[i]);
  }
  return result;
}

/**
 * Render text with a multi-stop gradient.
 *
 * @example
 * multiGradient("Hello World", ["#FF0000", "#00FF00", "#0000FF"])
 */
export function multiGradient(text: string, stops: string[]): string {
  if (!text) return "";
  if (stops.length < 2) return gradient(text, stops[0] ?? "#FFFFFF", stops[0] ?? "#FFFFFF");

  const len = text.length;
  if (len === 1) {
    const rgb = hexToRgb(stops[0]!);
    return chalk.rgb(rgb[0], rgb[1], rgb[2])(text);
  }

  const segments = stops.length - 1;
  let result = "";

  for (let i = 0; i < len; i++) {
    const globalT = i / (len - 1);
    const segIdx = Math.min(Math.floor(globalT * segments), segments - 1);
    const localT = (globalT * segments) - segIdx;
    const from = hexToRgb(stops[segIdx]!);
    const to = hexToRgb(stops[segIdx + 1]!);
    const r = lerp(from[0], to[0], localT);
    const g = lerp(from[1], to[1], localT);
    const b = lerp(from[2], to[2], localT);
    result += chalk.rgb(r, g, b)(text[i]);
  }
  return result;
}
