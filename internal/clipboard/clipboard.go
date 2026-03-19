package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Copy copies text to the system clipboard.
func Copy(text string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try wl-copy first (Wayland), then xclip (X11)
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (install wl-copy, xclip, or xsel)")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
	
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// CopyWithOSC52 copies text using the OSC 52 escape sequence.
// This works in most modern terminals without requiring external tools.
func CopyWithOSC52(text string) string {
	// OSC 52 format: ESC ] 52 ; c ; base64data BEL
	// c = clipboard selection
	encoded := base64Encode(text)
	return fmt.Sprintf("\x1b]52;c;%s\x07", encoded)
}

// base64Encode encodes text to base64 without external packages.
func base64Encode(s string) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	
	var result strings.Builder
	data := []byte(s)
	
	for i := 0; i < len(data); i += 3 {
		b1 := data[i]
		b2 := byte(0)
		b3 := byte(0)
		
		if i+1 < len(data) {
			b2 = data[i+1]
		}
		if i+2 < len(data) {
			b3 = data[i+2]
		}
		
		result.WriteByte(base64Chars[b1>>2])
		result.WriteByte(base64Chars[((b1&0x03)<<4)|(b2>>4)])
		
		if i+1 < len(data) {
			result.WriteByte(base64Chars[((b2&0x0f)<<2)|(b3>>6)])
		} else {
			result.WriteByte('=')
		}
		
		if i+2 < len(data) {
			result.WriteByte(base64Chars[b3&0x3f])
		} else {
			result.WriteByte('=')
		}
	}
	
	return result.String()
}
