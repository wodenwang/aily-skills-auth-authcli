package contextfile

import (
	"encoding/json"
	"errors"
	"os"
)

type File struct {
	UserID  string         `json:"user_id"`
	AgentID string         `json:"agent_id"`
	ChatID  string         `json:"chat_id"`
	Context map[string]any `json:"context"`
}

func Load(path string) (File, error) {
	if path == "" {
		return File{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{}, err
		}
		return File{}, err
	}
	var out File
	if err := json.Unmarshal(data, &out); err != nil {
		return File{}, err
	}
	if out.Context == nil {
		out.Context = map[string]any{}
	}
	return out, nil
}
