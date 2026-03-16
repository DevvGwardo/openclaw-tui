package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

func main() {
	token, port := readConfig()
	url := fmt.Sprintf("ws://127.0.0.1:%d", port)
	fmt.Printf("Connecting to %s with token=%s...\n", url, token[:8]+"...")

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Read challenge
	_, raw, _ := conn.ReadMessage()
	fmt.Printf("RECV: %s\n\n", string(raw))

	// Send connect
	connectID := uuid.New().String()
	connectFrame := map[string]interface{}{
		"type": "req",
		"id":   connectID,
		"method": "connect",
		"params": map[string]interface{}{
			"minProtocol": 3,
			"maxProtocol": 3,
			"client": map[string]interface{}{
				"id":          "gateway-client",
				"displayName": "debug-tui",
				"version":     "0.1.0",
				"platform":    "windows",
				"mode":        "backend",
			},
			"caps":   []string{},
			"auth":   map[string]interface{}{"token": token, "password": token},
			"role":   "operator",
			"scopes": []string{"operator.admin", "operator.read", "operator.write", "operator.approvals"},
		},
	}
	conn.WriteJSON(connectFrame)

	// Read connect response
	_, raw, _ = conn.ReadMessage()
	fmt.Printf("RECV connect response: %s\n\n", truncate(string(raw), 500))

	// Send a chat message
	chatID := uuid.New().String()
	chatFrame := map[string]interface{}{
		"type":   "req",
		"id":     chatID,
		"method": "chat.send",
		"params": map[string]interface{}{
			"sessionKey": "agent:main:main",
			"message":    "Say hello in one word",
			"thinking":   "none",
		},
	}
	fmt.Printf("SEND chat.send: %s\n\n", chatID)
	conn.WriteJSON(chatFrame)

	// Read all frames for 30 seconds
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	for i := 0; i < 50; i++ {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("READ ERROR: %v\n", err)
			break
		}
		var frame map[string]interface{}
		json.Unmarshal(raw, &frame)
		frameType, _ := frame["type"].(string)
		event, _ := frame["event"].(string)
		method, _ := frame["method"].(string)
		
		if frameType == "event" && event == "tick" {
			fmt.Printf("RECV [tick]\n")
			continue
		}
		
		fmt.Printf("RECV [%s/%s/%s] %s\n\n", frameType, event, method, truncate(string(raw), 800))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func readConfig() (string, int) {
	home, _ := os.UserHomeDir()
	data, _ := os.ReadFile(filepath.Join(home, ".openclaw", "openclaw.json"))
	var cfg struct {
		Gateway struct {
			Port int `json:"port"`
			Auth struct {
				Token string `json:"token"`
			} `json:"auth"`
		} `json:"gateway"`
	}
	json.Unmarshal(data, &cfg)
	port := cfg.Gateway.Port
	if port == 0 {
		port = 18789
	}
	return cfg.Gateway.Auth.Token, port
}
