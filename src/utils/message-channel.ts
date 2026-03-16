// Minimal message-channel stub for standalone TUI.
// The TUI only uses GATEWAY_CLIENT_MODES and GATEWAY_CLIENT_NAMES from this module,
// which are actually defined in gateway/protocol/client-info.ts.

export {
  GATEWAY_CLIENT_MODES,
  GATEWAY_CLIENT_NAMES,
  type GatewayClientMode,
  type GatewayClientName,
  normalizeGatewayClientMode,
  normalizeGatewayClientName,
} from "../gateway/protocol/client-info.js";
