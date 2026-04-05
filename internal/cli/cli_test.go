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
		AgentID: "explicit-agent",
		ChatID:  "explicit-chat",
		Format:  "env",
	}
	env := config.Settings{
		UserID:  "env-user",
		AgentID: "env-agent",
		ChatID:  "env-chat",
		Format:  "json",
	}
	ctx := contextfile.File{
		UserID:  "ctx-user",
		AgentID: "ctx-agent",
		ChatID:  "ctx-chat",
		Context: map[string]any{"requested_action": "read"},
	}
	file := config.File{
		UserID:  "file-user",
		AgentID: "file-agent",
		ChatID:  "file-chat",
		Format:  "exit-code",
	}

	input, err := ResolveInput(cmd, env, ctx, file)
	if err != nil {
		t.Fatalf("ResolveInput() error = %v", err)
	}
	if input.UserID != "explicit-user" || input.AgentID != "explicit-agent" {
		t.Fatalf("unexpected identity resolution: %+v", input)
	}
	if input.ChatID == nil || *input.ChatID != "explicit-chat" {
		t.Fatalf("unexpected chat resolution: %+v", input.ChatID)
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
		AgentID: " ",
		ChatID:  " ",
	}
	env := config.Settings{
		UserID:  "env-user",
		AgentID: "env-agent",
		ChatID:  "env-chat",
	}
	file := config.File{
		Format: "json",
	}

	input, err := ResolveInput(cmd, env, contextfile.File{}, file)
	if err != nil {
		t.Fatalf("ResolveInput() error = %v", err)
	}
	if input.UserID != "env-user" || input.AgentID != "env-agent" {
		t.Fatalf("unexpected fallback resolution: %+v", input)
	}
	if input.ChatID == nil || *input.ChatID != "env-chat" {
		t.Fatalf("unexpected chat fallback: %+v", input.ChatID)
	}
}
