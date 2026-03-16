// Minimal commands-registry stub for standalone TUI.
// In the full codebase this loads commands from config + data files.
// The TUI only calls listChatCommands / listChatCommandsForConfig to merge
// gateway-defined commands into the slash-command autocomplete list.

import type { OpenClawConfig } from "../config/types.js";

export type ChatCommandDefinition = {
  key: string;
  description: string;
  textAliases: string[];
};

export function listChatCommands(): ChatCommandDefinition[] {
  return [];
}

export function listChatCommandsForConfig(_cfg: OpenClawConfig): ChatCommandDefinition[] {
  return [];
}
