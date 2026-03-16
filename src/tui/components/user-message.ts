import { Container, Spacer, Text } from "@mariozechner/pi-tui";
import { markdownTheme, theme, getCurrentPalette } from "../theme/theme.js";
import { HyperlinkMarkdown } from "./hyperlink-markdown.js";

export class UserMessageComponent extends Container {
  private body: HyperlinkMarkdown;
  private roleLabel: Text;

  constructor(text: string) {
    super();
    const p = getCurrentPalette();

    // User prefix with accent chevron
    const prefix = theme.accentSoft("› ") + theme.bold(theme.userText("You"));
    this.roleLabel = new Text(prefix, 1, 0);

    this.body = new HyperlinkMarkdown(text, 1, 0, markdownTheme, {
      bgColor: (line) => theme.userBg(line),
      color: (line) => theme.userText(line),
    });
    this.addChild(new Spacer(1));
    this.addChild(this.roleLabel);
    this.addChild(this.body);
  }

  setText(text: string) {
    this.body.setText(text);
  }
}
