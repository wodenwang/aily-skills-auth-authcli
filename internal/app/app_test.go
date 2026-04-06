package app

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestCheckRefreshesCachedToken(t *testing.T) {
	var checkCalls int
	var refreshCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/auth/check":
			checkCalls++
			_, _ = w.Write([]byte(`{"request_id":"req_1","allowed":true,"access_token":"tok_1","token_type":"Bearer","expires_in":300,"refresh_before":240}`))
		case "/api/v1/token/refresh":
			refreshCalls++
			_, _ = w.Write([]byte(`{"access_token":"tok_2","expires_in":300,"refresh_before":240,"old_token_status":"refreshed","failure_code":null}`))
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
		Format:  "json",
		Context: map[string]any{},
	}

	if _, err := check(context.Background(), client, cachePath, input, now); err != nil {
		t.Fatalf("initial check() error = %v", err)
	}
	refreshed, err := check(context.Background(), client, cachePath, input, now.Add(250*time.Second))
	if err != nil {
		t.Fatalf("refresh check() error = %v", err)
	}
	if refreshed.AccessToken != "tok_2" {
		t.Fatalf("expected refreshed token, got %+v", refreshed)
	}
	if !refreshed.CacheHit {
		t.Fatalf("expected cache-backed refresh result: %+v", refreshed)
	}
	if checkCalls != 1 || refreshCalls != 1 {
		t.Fatalf("expected one check and one refresh, got check=%d refresh=%d", checkCalls, refreshCalls)
	}
}

func TestCheckRefreshFailureDeletesCacheAndRechecks(t *testing.T) {
	var checkCalls int
	var refreshCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/auth/check":
			checkCalls++
			if checkCalls == 1 {
				_, _ = w.Write([]byte(`{"request_id":"req_1","allowed":true,"access_token":"tok_1","token_type":"Bearer","expires_in":300,"refresh_before":240}`))
				return
			}
			_, _ = w.Write([]byte(`{"request_id":"req_2","allowed":true,"access_token":"tok_2","token_type":"Bearer","expires_in":300,"refresh_before":240}`))
		case "/api/v1/token/refresh":
			refreshCalls++
			_, _ = w.Write([]byte(`{"access_token":null,"expires_in":null,"refresh_before":null,"old_token_status":null,"failure_code":"TOKEN_REVOKED"}`))
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
		Format:  "json",
		Context: map[string]any{},
	}

	if _, err := check(context.Background(), client, cachePath, input, now); err != nil {
		t.Fatalf("initial check() error = %v", err)
	}
	reissued, err := check(context.Background(), client, cachePath, input, now.Add(250*time.Second))
	if err != nil {
		t.Fatalf("recheck after refresh failure error = %v", err)
	}
	if reissued.AccessToken != "tok_2" || reissued.RequestID != "req_2" {
		t.Fatalf("expected reissued token after refresh reset, got %+v", reissued)
	}
	if checkCalls != 2 || refreshCalls != 1 {
		t.Fatalf("expected two checks and one refresh, got check=%d refresh=%d", checkCalls, refreshCalls)
	}
}

func TestRunReturnsUpstreamExitCodeAndStableStderr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cachePath := filepath.Join(t.TempDir(), "tokens.json")
	t.Setenv("AUTHCLI_IAM_BASE_URL", server.URL)
	t.Setenv("AUTHCLI_CACHE_PATH", cachePath)

	stderr := captureStderr(t, func() {
		code := Run([]string{"check", "--skill", "sales-analysis", "--user-id", "ou_abc123", "--format", "exit-code"})
		if code != ExitUpstreamError {
			t.Fatalf("Run() code = %d", code)
		}
	})

	if !strings.Contains(stderr, "AUTHCLI_UPSTREAM_FAILURE:") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestRunReturnsInvalidInputExitCodeAndStableStderr(t *testing.T) {
	stderr := captureStderr(t, func() {
		code := Run([]string{"check", "--skill", "sales-analysis"})
		if code != ExitInvalidInput {
			t.Fatalf("Run() code = %d", code)
		}
	})

	if !strings.Contains(stderr, "AUTHCLI_INVALID_INPUT:") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestRunHelpReturnsZeroAndFormalHelpText(t *testing.T) {
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	code := Run([]string{"check", "--help"})

	_ = writer.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}

	if code != ExitAllowed {
		t.Fatalf("Run() code = %d", code)
	}
	output := buf.String()
	for _, expected := range []string{
		"auth-cli check --skill <skill_id> --user-id <user_id>",
		"Input Priority:",
		"Outputs:",
		"Deny vs Error:",
		"Cache Semantics:",
		"Install And Upgrade:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("missing %q in help output:\n%s", expected, output)
		}
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = writer
	defer func() {
		os.Stderr = original
	}()

	fn()

	_ = writer.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}
	return buf.String()
}
