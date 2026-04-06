package output

import (
	"bytes"
	"strings"
	"testing"

	"aily-skills-auth-authcli/internal/auth"
)

func TestWriteEnvDenied(t *testing.T) {
	var buf bytes.Buffer
	result := auth.Result{
		OK:          false,
		RequestID:   "req_123",
		Allowed:     false,
		DenyCode:    "GRANT_NOT_ACTIVE",
		DenyMessage: "denied",
	}

	if err := Write(&buf, "env", result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "AUTH_DENY_CODE=GRANT_NOT_ACTIVE") {
		t.Fatalf("missing deny code: %s", output)
	}
	if strings.Contains(output, "AUTH_ACCESS_TOKEN=") {
		t.Fatalf("unexpected token field: %s", output)
	}
}

func TestWriteJSONSuccessIncludesRequestID(t *testing.T) {
	var buf bytes.Buffer
	result := auth.Result{
		OK:          true,
		RequestID:   "req_123",
		Allowed:     true,
		TokenType:   "Bearer",
		AccessToken: "tok_123",
		AuthContext: &auth.AuthContext{
			UserID:  "ou_abc123",
			SkillID: "sales-analysis",
		},
	}

	if err := Write(&buf, "json", result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"request_id": "req_123"`) {
		t.Fatalf("missing request id: %s", output)
	}
}

func TestWriteEnvAllowDoesNotExposeLegacyIdentityFields(t *testing.T) {
	var buf bytes.Buffer
	result := auth.Result{
		OK:            true,
		RequestID:     "req_123",
		Allowed:       true,
		TokenType:     "Bearer",
		AccessToken:   "tok_123",
		ExpiresIn:     300,
		RefreshBefore: 240,
		AuthContext: &auth.AuthContext{
			UserID:  "ou_abc123",
			SkillID: "sales-analysis",
		},
	}

	if err := Write(&buf, "env", result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if strings.Contains(output, "AUTH_AGENT_ID=") || strings.Contains(output, "AUTH_CHAT_ID=") {
		t.Fatalf("unexpected legacy identity fields: %s", output)
	}
}
