package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"aily-skills-auth-authcli/internal/config"
	contextfile "aily-skills-auth-authcli/internal/context"
)

var ErrHelp = errors.New("help requested")

type Command struct {
	Name        string
	SkillID     string
	UserID      string
	Format      string
	ContextFile string
}

type Input struct {
	SkillID string
	UserID  string
	Format  string
	Context map[string]any
}

func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, errors.New("missing command")
	}
	cmd := Command{Name: args[0]}
	if cmd.Name != "check" {
		return Command{}, fmt.Errorf("unsupported command: %s", cmd.Name)
	}

	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cmd.SkillID, "skill", "", "skill id")
	fs.StringVar(&cmd.UserID, "user-id", "", "user id")
	fs.StringVar(&cmd.Format, "format", "", "output format")
	fs.StringVar(&cmd.ContextFile, "context-file", "", "runtime context file")
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return cmd, ErrHelp
		}
		return Command{}, err
	}
	if strings.TrimSpace(cmd.SkillID) == "" {
		return Command{}, errors.New("missing required flag: --skill")
	}
	return cmd, nil
}

func ResolveInput(cmd Command, env config.Settings, ctx contextfile.File, file config.File) (Input, error) {
	input := Input{
		SkillID: cmd.SkillID,
		UserID:  firstNonEmpty(cmd.UserID, env.UserID, ctx.UserID, file.UserID),
		Format:  strings.ToLower(firstNonEmpty(cmd.Format, env.Format, file.Format, "json")),
		Context: ctx.Context,
	}

	switch input.Format {
	case "json", "env", "exit-code":
	default:
		return Input{}, fmt.Errorf("unsupported format: %s", input.Format)
	}

	if input.UserID == "" {
		return Input{}, errors.New("user_id is required")
	}
	return input, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
