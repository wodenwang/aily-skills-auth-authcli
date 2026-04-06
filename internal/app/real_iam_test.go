package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aily-skills-auth-authcli/internal/auth"
	"aily-skills-auth-authcli/internal/cache"
	"aily-skills-auth-authcli/internal/cli"
)

func TestRealIAMAllowJSON(t *testing.T) {
	run := newRealIAMRun(t)
	result := run.run([]string{
		"check",
		"--skill", "sales-analysis",
		"--user-id", "ou_abc123",
		"--format", "json",
		"--context-file", filepath.Join(run.repoRoot, "examples", "context-private.json"),
	})

	if result.exitCode != ExitAllowed {
		t.Fatalf("exitCode = %d stderr=%s", result.exitCode, result.stderr)
	}
	var body auth.Result
	mustJSON(t, result.stdout, &body)
	if !body.OK || !body.Allowed || body.AccessToken == "" {
		t.Fatalf("unexpected allow result: %+v", body)
	}
	if body.AuthContext == nil || body.AuthContext.UserID != "ou_abc123" || body.AuthContext.SkillID != "sales-analysis" {
		t.Fatalf("unexpected auth_context: %+v", body.AuthContext)
	}
}

func TestRealIAMAllowEnvOutput(t *testing.T) {
	run := newRealIAMRun(t)
	result := run.run([]string{
		"check",
		"--skill", "sales-analysis",
		"--user-id", "ou_abc123",
		"--format", "env",
		"--context-file", filepath.Join(run.repoRoot, "examples", "context-private.json"),
	})

	if result.exitCode != ExitAllowed {
		t.Fatalf("exitCode = %d stderr=%s", result.exitCode, result.stderr)
	}
	for _, expected := range []string{
		"AUTH_OK=true",
		"AUTH_ALLOWED=true",
		"AUTH_TOKEN_TYPE=Bearer",
		"AUTH_USER_ID=ou_abc123",
		"AUTH_SKILL_ID=sales-analysis",
	} {
		if !strings.Contains(result.stdout, expected) {
			t.Fatalf("missing %q in env output:\n%s", expected, result.stdout)
		}
	}
	if strings.Contains(result.stdout, "AUTH_AGENT_ID=") || strings.Contains(result.stdout, "AUTH_CHAT_ID=") {
		t.Fatalf("unexpected legacy identity fields:\n%s", result.stdout)
	}
}

func TestRealIAMDenyJSON(t *testing.T) {
	run := newRealIAMRun(t)
	result := run.run([]string{
		"check",
		"--skill", "sales-analysis",
		"--user-id", "ou_left999",
		"--format", "json",
		"--context-file", filepath.Join(run.repoRoot, "examples", "context-private.json"),
	})

	if result.exitCode != ExitDenied {
		t.Fatalf("exitCode = %d stdout=%s stderr=%s", result.exitCode, result.stdout, result.stderr)
	}
	var body auth.Result
	mustJSON(t, result.stdout, &body)
	if body.Allowed || body.DenyCode == "" {
		t.Fatalf("unexpected deny result: %+v", body)
	}
	if body.AuthContext != nil {
		t.Fatalf("deny response should not include auth_context: %+v", body.AuthContext)
	}
}

func TestRealIAMRefreshPath(t *testing.T) {
	client := newRealIAMClient(t)
	cachePath := filepath.Join(t.TempDir(), "tokens.json")
	key := cache.Key{
		UserID:  "ou_abc123",
		SkillID: "sales-analysis",
	}
	now := time.Now().UTC()

	first, err := check(context.Background(), client, cachePath, cli.Input{
		SkillID: "sales-analysis",
		UserID:  "ou_abc123",
		Format:  "json",
		Context: map[string]any{"requested_action": "read"},
	}, now)
	if err != nil {
		t.Fatalf("initial check() error = %v", err)
	}

	cacheFile, err := cache.Load(cachePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	entry, _, found := cache.Find(cacheFile, key)
	if !found {
		t.Fatal("expected cache entry")
	}

	entry.RefreshBeforeAt = now.Add(-1 * time.Second)
	entry.ExpiresAt = now.Add(30 * time.Second)
	cache.Upsert(&cacheFile, entry)
	if err := cache.Save(cachePath, cacheFile); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	second, err := check(context.Background(), client, cachePath, cli.Input{
		SkillID: "sales-analysis",
		UserID:  "ou_abc123",
		Format:  "json",
		Context: map[string]any{"requested_action": "read"},
	}, now)
	if err != nil {
		t.Fatalf("refresh check() error = %v", err)
	}
	if second.AccessToken == "" || second.AccessToken == first.AccessToken {
		t.Fatalf("expected refreshed token, got first=%q second=%q", first.AccessToken, second.AccessToken)
	}
}

func TestRealIAMCacheReuseBeforeRefreshWindow(t *testing.T) {
	run := newRealIAMRun(t)
	args := []string{
		"check",
		"--skill", "sales-analysis",
		"--user-id", "ou_abc123",
		"--format", "json",
		"--context-file", filepath.Join(run.repoRoot, "examples", "context-private.json"),
	}

	first := run.run(args)
	second := run.run(args)
	if first.exitCode != ExitAllowed || second.exitCode != ExitAllowed {
		t.Fatalf("unexpected exit codes first=%d second=%d", first.exitCode, second.exitCode)
	}

	var firstBody auth.Result
	var secondBody auth.Result
	mustJSON(t, first.stdout, &firstBody)
	mustJSON(t, second.stdout, &secondBody)

	if secondBody.CacheHit != true {
		t.Fatalf("expected cache hit on second call: %+v", secondBody)
	}
	if firstBody.AccessToken != secondBody.AccessToken {
		t.Fatalf("expected cached token reuse")
	}
}

func TestRealIAMUpstreamFailClosed(t *testing.T) {
	run := newRealIAMRunWithBaseURL(t, "http://127.0.0.1:18000")
	result := run.run([]string{
		"check",
		"--skill", "sales-analysis",
		"--user-id", "ou_abc123",
		"--format", "exit-code",
	})

	if result.exitCode != ExitUpstreamError {
		t.Fatalf("exitCode = %d stderr=%s", result.exitCode, result.stderr)
	}
	if !strings.Contains(result.stderr, "AUTHCLI_UPSTREAM_FAILURE:") {
		t.Fatalf("unexpected stderr: %s", result.stderr)
	}
	if strings.Contains(result.stderr, "eyJ") {
		t.Fatalf("stderr must not contain token-like data: %s", result.stderr)
	}
}

func newRealIAMClient(t *testing.T) auth.Client {
	t.Helper()

	baseURL := os.Getenv("AUTHCLI_REAL_IAM_BASE_URL")
	if baseURL == "" {
		t.Skip("AUTHCLI_REAL_IAM_BASE_URL is not set")
	}
	return auth.NewHTTPClient(baseURL, &http.Client{Timeout: 5 * time.Second})
}

func isDenied(err error) bool {
	return errors.Is(err, errDenied)
}

type realIAMRun struct {
	repoRoot  string
	cachePath string
	baseURL   string
}

type cliRunResult struct {
	exitCode int
	stdout   string
	stderr   string
}

func newRealIAMRun(t *testing.T) realIAMRun {
	t.Helper()

	baseURL := os.Getenv("AUTHCLI_REAL_IAM_BASE_URL")
	if baseURL == "" {
		t.Skip("AUTHCLI_REAL_IAM_BASE_URL is not set")
	}
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	repoRoot = filepath.Clean(filepath.Join(repoRoot, "..", ".."))
	return realIAMRun{
		repoRoot:  repoRoot,
		cachePath: filepath.Join(t.TempDir(), "tokens.json"),
		baseURL:   baseURL,
	}
}

func newRealIAMRunWithBaseURL(t *testing.T, baseURL string) realIAMRun {
	t.Helper()

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	repoRoot = filepath.Clean(filepath.Join(repoRoot, "..", ".."))
	return realIAMRun{
		repoRoot:  repoRoot,
		cachePath: filepath.Join(t.TempDir(), "tokens.json"),
		baseURL:   baseURL,
	}
}

func (r realIAMRun) run(args []string) cliRunResult {
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	originalBaseURL := os.Getenv("AUTHCLI_IAM_BASE_URL")
	originalCachePath := os.Getenv("AUTHCLI_CACHE_PATH")

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	_ = os.Setenv("AUTHCLI_IAM_BASE_URL", r.baseURL)
	_ = os.Setenv("AUTHCLI_CACHE_PATH", r.cachePath)

	exitCode := Run(args)

	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr
	_ = os.Setenv("AUTHCLI_IAM_BASE_URL", originalBaseURL)
	_ = os.Setenv("AUTHCLI_CACHE_PATH", originalCachePath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	_, _ = io.Copy(&stdout, stdoutReader)
	_, _ = io.Copy(&stderr, stderrReader)

	return cliRunResult{
		exitCode: exitCode,
		stdout:   strings.TrimSpace(stdout.String()),
		stderr:   strings.TrimSpace(stderr.String()),
	}
}

func mustJSON(t *testing.T, raw string, out any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\nraw=%s", err, raw)
	}
}
