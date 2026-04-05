package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"aily-skills-auth-authcli/internal/auth"
	"aily-skills-auth-authcli/internal/cache"
	"aily-skills-auth-authcli/internal/cli"
)

func TestCheckUsesCacheAfterInitialAuth(t *testing.T) {
	var checkCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/check":
			checkCalls++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"request_id":"req_1","allowed":true,"access_token":"tok_1","token_type":"Bearer","expires_in":300,"refresh_before":240}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := auth.NewHTTPClient(server.URL, server.Client())
	cachePath := filepath.Join(t.TempDir(), "tokens.json")
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	input := cli.Input{
		SkillID: "sales-analysis",
		UserID:  "ou_abc123",
		AgentID: "agent_1",
		Format:  "json",
		Context: map[string]any{},
	}

	first, err := check(context.Background(), client, cachePath, input, now)
	if err != nil {
		t.Fatalf("first check() error = %v", err)
	}
	if !first.Allowed || first.CacheHit {
		t.Fatalf("unexpected first result: %+v", first)
	}

	second, err := check(context.Background(), client, cachePath, input, now.Add(10*time.Second))
	if err != nil {
		t.Fatalf("second check() error = %v", err)
	}
	if !second.CacheHit {
		t.Fatalf("expected cache hit: %+v", second)
	}
	if checkCalls != 1 {
		t.Fatalf("expected 1 auth call, got %d", checkCalls)
	}

	cacheFile, err := cache.Load(cachePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cacheFile.Entries) != 1 {
		t.Fatalf("expected 1 cache entry, got %d", len(cacheFile.Entries))
	}
}
