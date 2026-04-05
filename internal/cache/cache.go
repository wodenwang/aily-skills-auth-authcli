package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const Version = 1

type Entry struct {
	CacheKey        string    `json:"cache_key"`
	RequestID       string    `json:"request_id"`
	UserID          string    `json:"user_id"`
	SkillID         string    `json:"skill_id"`
	AgentID         string    `json:"agent_id"`
	ChatID          *string   `json:"chat_id"`
	AccessToken     string    `json:"access_token"`
	TokenType       string    `json:"token_type"`
	ExpiresAt       time.Time `json:"expires_at"`
	RefreshBeforeAt time.Time `json:"refresh_before_at"`
	CachedAt        time.Time `json:"cached_at"`
	Source          string    `json:"source"`
}

type File struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

type State int

const (
	StateMiss State = iota
	StateValid
	StateRefresh
	StateExpired
)

type Key struct {
	UserID  string
	SkillID string
	AgentID string
	ChatID  *string
}

func DefaultFile() File {
	return File{Version: Version, Entries: []Entry{}}
}

func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultFile(), nil
		}
		return File{}, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, err
	}
	if f.Version == 0 {
		f.Version = Version
	}
	if f.Entries == nil {
		f.Entries = []Entry{}
	}
	return f, nil
}

func Save(path string, f File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func CacheKey(key Key) string {
	return fmt.Sprintf("%s|%s|%s|%s", key.UserID, key.SkillID, key.AgentID, chatPart(key.ChatID))
}

func Find(f File, key Key) (Entry, int, bool) {
	want := CacheKey(key)
	for i, entry := range f.Entries {
		if entry.CacheKey == want {
			return entry, i, true
		}
	}
	return Entry{}, -1, false
}

func Upsert(f *File, entry Entry) {
	for i := range f.Entries {
		if f.Entries[i].CacheKey == entry.CacheKey {
			f.Entries[i] = entry
			return
		}
	}
	f.Entries = append(f.Entries, entry)
}

func Delete(f *File, key Key) {
	want := CacheKey(key)
	filtered := f.Entries[:0]
	for _, entry := range f.Entries {
		if entry.CacheKey != want {
			filtered = append(filtered, entry)
		}
	}
	f.Entries = filtered
}

func Evaluate(entry Entry, now time.Time) State {
	if now.Before(entry.RefreshBeforeAt) {
		return StateValid
	}
	if now.Before(entry.ExpiresAt) {
		return StateRefresh
	}
	return StateExpired
}

func NewEntry(key Key, requestID, accessToken, tokenType string, expiresIn, refreshBefore int, now time.Time, source string) Entry {
	return Entry{
		CacheKey:        CacheKey(key),
		RequestID:       requestID,
		UserID:          key.UserID,
		SkillID:         key.SkillID,
		AgentID:         key.AgentID,
		ChatID:          key.ChatID,
		AccessToken:     accessToken,
		TokenType:       tokenType,
		ExpiresAt:       now.Add(time.Duration(expiresIn) * time.Second).UTC(),
		RefreshBeforeAt: now.Add(time.Duration(refreshBefore) * time.Second).UTC(),
		CachedAt:        now.UTC(),
		Source:          source,
	}
}

func chatPart(chatID *string) string {
	if chatID == nil || *chatID == "" {
		return "null"
	}
	return *chatID
}
