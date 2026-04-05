package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	UserID     string `json:"user_id"`
	AgentID    string `json:"agent_id"`
	ChatID     string `json:"chat_id"`
	Format     string `json:"format"`
	IAMBaseURL string `json:"iam_base_url"`
	CachePath  string `json:"cache_path"`
}

type Settings struct {
	FilePath   string
	UserID     string
	AgentID    string
	ChatID     string
	Format     string
	IAMBaseURL string
	CachePath  string
	Timeout    time.Duration
}

func LoadFromEnv() Settings {
	home, _ := os.UserHomeDir()
	configPath := envOrDefault("AUTHCLI_CONFIG_FILE", filepath.Join(home, ".aily-skills-auth", "config.json"))
	timeout := 5 * time.Second
	if raw := os.Getenv("AUTHCLI_TIMEOUT"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			timeout = parsed
		}
	}
	return Settings{
		FilePath:   configPath,
		UserID:     os.Getenv("AUTHCLI_USER_ID"),
		AgentID:    os.Getenv("AUTHCLI_AGENT_ID"),
		ChatID:     os.Getenv("AUTHCLI_CHAT_ID"),
		Format:     os.Getenv("AUTHCLI_FORMAT"),
		IAMBaseURL: os.Getenv("AUTHCLI_IAM_BASE_URL"),
		CachePath:  os.Getenv("AUTHCLI_CACHE_PATH"),
		Timeout:    timeout,
	}
}

func DefaultCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aily-skills-auth", "cache", "tokens.json")
}

func LoadFile(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{}, nil
		}
		return File{}, err
	}
	var out File
	if err := json.Unmarshal(data, &out); err != nil {
		return File{}, err
	}
	return out, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
