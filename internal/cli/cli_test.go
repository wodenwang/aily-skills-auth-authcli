package cli

import (
	"testing"

	"aily-skills-auth-authcli/internal/config"
	contextfile "aily-skills-auth-authcli/internal/context"
)

func TestResolveInputPriority(t *testing.T) {
	cmd := Command{
		Name:    "check",
		SkillID: "sales-analysis",
		UserID:  "explicit-user",
		Format:  "env",
	}
	env := config.Settings{
		UserID:  "env-user",
		Format:  "json",
	}
	ctx := contextfile.File{
		UserID:  "ctx-user",
		Context: map[string]any{"requested_action": "read"},
	}
	file := config.File{
		UserID: "file-user",
		Format:  "exit-code",
	}

	input, err := ResolveInput(cmd, env, ctx, file)
	if err != nil {
		t.Fatalf("ResolveInput() error = %v", err)
	}
	if input.UserID != "explicit-user" {
		t.Fatalf("unexpected identity resolution: %+v", input)
	}
	if input.Format != "env" {
		t.Fatalf("unexpected format: %s", input.Format)
	}
	if got := input.Context["requested_action"]; got != "read" {
		t.Fatalf("unexpected context: %#v", input.Context)
	}
}

func TestResolveInputSkipsBlankValues(t *testing.T) {
	cmd := Command{
		Name:    "check",
		SkillID: "sales-analysis",
		UserID:  "   ",
	}
	env := config.Settings{
		UserID:  "env-user",
	}
	file := config.File{
		Format: "json",
	}

	input, err := ResolveInput(cmd, env, contextfile.File{}, file)
	if err != nil {
		t.Fatalf("ResolveInput() error = %v", err)
	}
	if input.UserID != "env-user" {
		t.Fatalf("unexpected fallback resolution: %+v", input)
	}
}
