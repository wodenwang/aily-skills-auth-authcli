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
		DenyCode:    "CHAT_SKILL_DENIED",
		DenyMessage: "denied",
	}

	if err := Write(&buf, "env", result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "AUTH_DENY_CODE=CHAT_SKILL_DENIED") {
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
	}

	if err := Write(&buf, "json", result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"request_id": "req_123"`) {
		t.Fatalf("missing request id: %s", output)
	}
}
