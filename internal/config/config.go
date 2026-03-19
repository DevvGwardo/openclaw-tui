package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds user preferences and state.
type Config struct {
	Theme         string   `json:"theme"`
	Session       string   `json:"session"`
	Thinking      string   `json:"thinking"`
	Background    string   `json:"background"`
	History       []string `json:"history"`       // Recent commands/messages
	MaxHistory    int      `json:"max_history"`   // Max history items to keep
	ShowTimestamps bool    `json:"show_timestamps"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Theme:          "ocean",
		Session:        "agent:main:main",
		Thinking:       "adaptive",
		Background:     "off",
		History:        []string{},
		MaxHistory:     1000,
		ShowTimestamps: true,
	}
}

// ConfigPath returns the path to the config file.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "openclaw-tui", "config.json")
}

// Load reads the config from disk.
func Load() (Config, error) {
	cfg := DefaultConfig()
	
	path := ConfigPath()
	if path == "" {
		return cfg, fmt.Errorf("could not determine config path")
	}
	
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return cfg, err
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults for new config
			return cfg, nil
		}
		return cfg, err
	}
	
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	
	return cfg, nil
}

// Save writes the config to disk.
func (c Config) Save() error {
	path := ConfigPath()
	if path == "" {
		return fmt.Errorf("could not determine config path")
	}
	
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// AddHistory adds an entry to command history, preventing duplicates at the end.
func (c *Config) AddHistory(entry string) {
	if entry == "" {
		return
	}
	
	// Don't add if same as last entry
	if len(c.History) > 0 && c.History[len(c.History)-1] == entry {
		return
	}
	
	c.History = append(c.History, entry)
	
	// Trim to max size
	if len(c.History) > c.MaxHistory {
		c.History = c.History[len(c.History)-c.MaxHistory:]
	}
}

// GetHistory returns history entries matching a prefix (for autocomplete).
func (c Config) GetHistory(prefix string, limit int) []string {
	var matches []string
	seen := make(map[string]bool)
	
	// Search in reverse (most recent first)
	for i := len(c.History) - 1; i >= 0 && len(matches) < limit; i-- {
		entry := c.History[i]
		if seen[entry] {
			continue
		}
		if prefix == "" || len(entry) >= len(prefix) && entry[:len(prefix)] == prefix {
			matches = append(matches, entry)
			seen[entry] = true
		}
	}
	
	return matches
}
