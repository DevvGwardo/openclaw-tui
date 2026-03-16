package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

func main() {
	token, port := readConfig()
	url := fmt.Sprintf("ws://127.0.0.1:%d", port)
	
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Read challenge
	_, raw, _ := conn.ReadMessage()
	var challenge map[string]interface{}
	json.Unmarshal(raw, &challenge)

	// Send connect — try with NO scopes, let gateway assign them
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
			"auth":   map[string]interface{}{"token": token},
			"role":   "operator",
			"scopes": []string{"operator.admin"},
		},
	}
	conn.WriteJSON(connectFrame)

	// Read connect response — FULL
	_, raw, _ = conn.ReadMessage()
	var prettyJSON map[string]interface{}
	json.Unmarshal(raw, &prettyJSON)
	pretty, _ := json.MarshalIndent(prettyJSON, "", "  ")
	fmt.Printf("HELLO-OK:\n%s\n", string(pretty))
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
	if port == 0 { port = 18789 }
	return cfg.Gateway.Auth.Token, port
}
