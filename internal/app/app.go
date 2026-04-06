package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"aily-skills-auth-authcli/internal/auth"
	"aily-skills-auth-authcli/internal/cache"
	"aily-skills-auth-authcli/internal/cli"
	"aily-skills-auth-authcli/internal/config"
	contextfile "aily-skills-auth-authcli/internal/context"
	"aily-skills-auth-authcli/internal/output"
)

const (
	ExitAllowed       = 0
	ExitDenied        = 10
	ExitInvalidInput  = 20
	ExitCacheFailure  = 30
	ExitUpstreamError = 40
	ExitInternalError = 50
)

var (
	errDenied       = errors.New("denied")
	errInvalidInput = errors.New("invalid input")
	errCacheFailure = errors.New("cache failure")
	errUpstream     = errors.New("upstream failure")
	errInternal     = errors.New("internal error")
)

func Run(args []string) int {
	if len(args) >= 2 && args[0] == "check" && (args[1] == "--help" || args[1] == "-h") {
		_, _ = fmt.Fprint(os.Stdout, helpText())
		return ExitAllowed
	}
	if err := run(args); err != nil {
		switch {
		case errors.Is(err, errDenied):
			return ExitDenied
		case errors.Is(err, errInvalidInput):
			_, _ = fmt.Fprintln(os.Stderr, renderError(err))
			return ExitInvalidInput
		case errors.Is(err, errCacheFailure):
			_, _ = fmt.Fprintln(os.Stderr, renderError(err))
			return ExitCacheFailure
		case errors.Is(err, errUpstream):
			_, _ = fmt.Fprintln(os.Stderr, renderError(err))
			return ExitUpstreamError
		default:
			_, _ = fmt.Fprintln(os.Stderr, renderError(err))
			return ExitInternalError
		}
	}
	return ExitAllowed
}

func run(args []string) error {
	command, err := cli.Parse(args)
	if err != nil {
		if errors.Is(err, cli.ErrHelp) {
			_, _ = fmt.Fprint(os.Stdout, helpText())
			return nil
		}
		return fmt.Errorf("%w: %v", errInvalidInput, err)
	}

	settings := config.LoadFromEnv()
	fileConfig, err := config.LoadFile(settings.FilePath)
	if err != nil {
		return fmt.Errorf("%w: load config: %v", errInvalidInput, err)
	}
	runtimeContext, err := contextfile.Load(command.ContextFile)
	if err != nil {
		return fmt.Errorf("%w: load context file: %v", errInvalidInput, err)
	}
	input, err := cli.ResolveInput(command, settings, runtimeContext, fileConfig)
	if err != nil {
		return fmt.Errorf("%w: %v", errInvalidInput, err)
	}

	baseURL := firstNonEmpty(settings.IAMBaseURL, fileConfig.IAMBaseURL, "http://127.0.0.1:8000")
	cachePath := firstNonEmpty(settings.CachePath, fileConfig.CachePath, config.DefaultCachePath())

	httpClient := &http.Client{Timeout: settings.Timeout}
	client := auth.NewHTTPClient(baseURL, httpClient)
	now := time.Now().UTC()

	result, err := check(context.Background(), client, cachePath, input, now)
	if err != nil {
		if errors.Is(err, errDenied) {
			return err
		}
		return err
	}
	if err := output.Write(os.Stdout, input.Format, result); err != nil {
		return fmt.Errorf("%w: write output: %v", errInternal, err)
	}
	return nil
}

func check(ctx context.Context, client auth.Client, cachePath string, input cli.Input, now time.Time) (auth.Result, error) {
	key := cache.Key{
		UserID:  input.UserID,
		SkillID: input.SkillID,
	}

	cacheFile, err := cache.Load(cachePath)
	if err != nil {
		return auth.Result{}, fmt.Errorf("%w: load cache: %v", errCacheFailure, err)
	}

	if entry, _, found := cache.Find(cacheFile, key); found {
		switch cache.Evaluate(entry, now) {
		case cache.StateValid:
			return allowedResultFromEntry(entry, key, true), nil
		case cache.StateRefresh:
			refreshResp, refreshErr := client.Refresh(ctx, entry.AccessToken)
			if refreshErr == nil {
				if refreshResp.FailureCode == "" {
					updated := cache.NewEntry(key, entry.RequestID, refreshResp.AccessToken, entry.TokenType, refreshResp.ExpiresIn, refreshResp.RefreshBefore, now, "token_refresh")
					cache.Upsert(&cacheFile, updated)
					if saveErr := cache.Save(cachePath, cacheFile); saveErr != nil {
						return auth.Result{}, fmt.Errorf("%w: save cache after refresh: %v", errCacheFailure, saveErr)
					}
					return allowedResultFromEntry(updated, key, true), nil
				}
				if isRefreshResetCode(refreshResp.FailureCode) {
					cache.Delete(&cacheFile, key)
					if saveErr := cache.Save(cachePath, cacheFile); saveErr != nil {
						return auth.Result{}, fmt.Errorf("%w: delete invalid cache: %v", errCacheFailure, saveErr)
					}
					break
				}
				return auth.Result{}, fmt.Errorf("%w: refresh failure_code=%s", errUpstream, refreshResp.FailureCode)
			}
			var apiErr auth.APIError
			if errors.As(refreshErr, &apiErr) && isRefreshResetCode(apiErr.Code) {
				cache.Delete(&cacheFile, key)
				if saveErr := cache.Save(cachePath, cacheFile); saveErr != nil {
					return auth.Result{}, fmt.Errorf("%w: delete invalid cache: %v", errCacheFailure, saveErr)
				}
			} else {
				return auth.Result{}, fmt.Errorf("%w: %v", errUpstream, refreshErr)
			}
		case cache.StateExpired:
			cache.Delete(&cacheFile, key)
			if saveErr := cache.Save(cachePath, cacheFile); saveErr != nil {
				return auth.Result{}, fmt.Errorf("%w: delete expired cache: %v", errCacheFailure, saveErr)
			}
		}
	}

	resp, err := client.Check(ctx, auth.CheckRequest{
		UserID:  input.UserID,
		SkillID: input.SkillID,
		Context: input.Context,
	})
	if err != nil {
		return auth.Result{}, fmt.Errorf("%w: %v", errUpstream, err)
	}

	if !resp.Allowed {
		result := auth.Result{
			OK:          false,
			RequestID:   resp.RequestID,
			Allowed:     false,
			DenyCode:    resp.DenyCode,
			DenyMessage: resp.DenyMessage,
		}
		if writeErr := output.Write(os.Stdout, input.Format, result); writeErr != nil {
			return auth.Result{}, fmt.Errorf("%w: write deny output: %v", errInternal, writeErr)
		}
		return auth.Result{}, errDenied
	}

	entry := cache.NewEntry(key, resp.RequestID, resp.AccessToken, resp.TokenType, resp.ExpiresIn, resp.RefreshBefore, now, "auth_check")
	cache.Upsert(&cacheFile, entry)
	if err := cache.Save(cachePath, cacheFile); err != nil {
		return auth.Result{}, fmt.Errorf("%w: save cache: %v", errCacheFailure, err)
	}
	return auth.Result{
		OK:            true,
		RequestID:     resp.RequestID,
		Allowed:       true,
		TokenType:     resp.TokenType,
		AccessToken:   resp.AccessToken,
		ExpiresIn:     resp.ExpiresIn,
		RefreshBefore: resp.RefreshBefore,
		CacheHit:      false,
		AuthContext: &auth.AuthContext{
			UserID:  input.UserID,
			SkillID: input.SkillID,
		},
	}, nil
}

func allowedResultFromEntry(entry cache.Entry, key cache.Key, cacheHit bool) auth.Result {
	expiresIn := int(time.Until(entry.ExpiresAt).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}
	refreshBefore := int(time.Until(entry.RefreshBeforeAt).Seconds())
	if refreshBefore < 0 {
		refreshBefore = 0
	}
	return auth.Result{
		OK:            true,
		RequestID:     entry.RequestID,
		Allowed:       true,
		TokenType:     entry.TokenType,
		AccessToken:   entry.AccessToken,
		ExpiresIn:     expiresIn,
		RefreshBefore: refreshBefore,
		CacheHit:      cacheHit,
		AuthContext: &auth.AuthContext{
			UserID:  key.UserID,
			SkillID: key.SkillID,
		},
	}
}

func isRefreshResetCode(code string) bool {
	switch code {
	case "TOKEN_REVOKED", "TOKEN_INVALID", "TOKEN_EXPIRED":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func renderError(err error) string {
	switch {
	case errors.Is(err, errInvalidInput):
		return fmt.Sprintf("AUTHCLI_INVALID_INPUT: %s", errorDetails(err, errInvalidInput))
	case errors.Is(err, errCacheFailure):
		return fmt.Sprintf("AUTHCLI_CACHE_FAILURE: %s", errorDetails(err, errCacheFailure))
	case errors.Is(err, errUpstream):
		return fmt.Sprintf("AUTHCLI_UPSTREAM_FAILURE: %s", renderUpstreamDetails(err))
	default:
		return fmt.Sprintf("AUTHCLI_INTERNAL_ERROR: %s", errorDetails(err, errInternal))
	}
}

func errorDetails(err error, marker error) string {
	text := err.Error()
	prefix := marker.Error() + ": "
	if len(text) >= len(prefix) && text[:len(prefix)] == prefix {
		return text[len(prefix):]
	}
	return text
}

func renderUpstreamDetails(err error) string {
	var apiErr auth.APIError
	if errors.As(err, &apiErr) {
		if apiErr.RequestID != "" {
			return fmt.Sprintf("request_id=%s code=%s message=%s", apiErr.RequestID, apiErr.Code, apiErr.Message)
		}
		if apiErr.Code != "" {
			return fmt.Sprintf("code=%s message=%s", apiErr.Code, apiErr.Message)
		}
	}
	return errorDetails(err, errUpstream)
}

func helpText() string {
	lines := []string{
		"Usage:",
		"  auth-cli check --skill <skill_id> --user-id <user_id> [--format <json|env|exit-code>] [--context-file <path>]",
		"",
		"Purpose:",
		"  Run local Skill authorization against IAM using the frozen 0.2.0 minimum input model: user_id + skill_id.",
		"",
		"Input Priority:",
		"  1. explicit flags",
		"  2. environment variables",
		"  3. runtime context file",
		"  4. local config file",
		"",
		"Required Inputs:",
		"  --skill",
		"  --user-id",
		"",
		"Optional Inputs:",
		"  --format",
		"  --context-file",
		"",
		"Environment:",
		"  AUTHCLI_USER_ID",
		"  AUTHCLI_FORMAT",
		"  AUTHCLI_IAM_BASE_URL",
		"  AUTHCLI_TIMEOUT",
		"  AUTHCLI_CACHE_PATH",
		"  AUTHCLI_CONFIG_FILE",
		"",
		"Outputs:",
		"  json      default; prints the frozen JSON contract",
		"  env       prints AUTH_* lines; token fields only on allow",
		"  exit-code prints no body; rely on process exit status",
		"",
		"Exit Codes:",
		"  0  allowed",
		"  10 denied by policy",
		"  20 invalid input",
		"  30 cache read/write failure",
		"  40 upstream unavailable or timeout",
		"  50 unexpected internal error",
		"",
		"Deny vs Error:",
		"  deny is a valid authorization decision and returns exit code 10.",
		"  upstream failures, cache failures, and invalid input are errors and always fail closed.",
		"",
		"Cache Semantics:",
		"  cache key is user_id + skill_id.",
		"  before refresh_before_at: reuse cached token.",
		"  from refresh_before_at until expires_at: call /api/v1/token/refresh first.",
		"  at or after expires_at, or after refresh reset codes: delete cache entry and call /api/v1/auth/check.",
		"",
		"Install And Upgrade:",
		"  use the release install script for tagged versions.",
		"  curl -fsSL https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.2.0/install-authcli.sh | sh -s -- --version v0.2.0 --install-dir /usr/local/bin",
		"",
		"Examples:",
		"  auth-cli check --skill sales-analysis --user-id ou_abc123 --format json",
		"  auth-cli check --skill <skill_id> --user-id <user_id> --format json",
	}
	return strings.Join(lines, "\n") + "\n"
}
