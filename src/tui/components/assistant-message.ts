import { Container, Spacer, Text } from "@mariozechner/pi-tui";
import { markdownTheme, theme, getCurrentPalette } from "../theme/theme.js";
import { gradient } from "../theme/gradient.js";
import { HyperlinkMarkdown } from "./hyperlink-markdown.js";

const BRAILLE_FRAMES = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];

export class AssistantMessageComponent extends Container {
  private body: HyperlinkMarkdown;
  private roleLabel: Text;

  constructor(text: string) {
    super();
    const p = getCurrentPalette();

    // Role badge with colored border and claw emoji
    const badge = gradient("┃ 🦞 Assistant", p.primary, p.tertiary);
    this.roleLabel = new Text(badge, 1, 0);

    this.body = new HyperlinkMarkdown(text, 1, 0, markdownTheme, {
      // Keep assistant body text in terminal default foreground so contrast
      // follows the user's terminal theme (dark or light).
      color: (line) => {
        const border = theme.accent("┃");
        return `${border} ${theme.assistantText(line)}`;
      },
    });
    this.addChild(new Spacer(1));
    this.addChild(this.roleLabel);
    this.addChild(this.body);
  }

  setText(text: string) {
    this.body.setText(text);
  }

  /**
   * Get the braille spinner character for the given tick (used during streaming).
   */
  static spinnerFrame(tick: number): string {
    return BRAILLE_FRAMES[tick % BRAILLE_FRAMES.length] ?? BRAILLE_FRAMES[0]!;
  }
}
