package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"aily-skills-auth-authcli/internal/auth"
)

func Write(w io.Writer, format string, result auth.Result) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	case "env":
		return writeEnv(w, result)
	case "exit-code":
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func writeEnv(w io.Writer, result auth.Result) error {
	lines := []string{
		fmt.Sprintf("AUTH_OK=%t", result.OK),
		fmt.Sprintf("AUTH_ALLOWED=%t", result.Allowed),
		fmt.Sprintf("AUTH_REQUEST_ID=%s", result.RequestID),
	}
	if result.Allowed {
		authContext := auth.AuthContext{}
		if result.AuthContext != nil {
			authContext = *result.AuthContext
		}
		lines = append(lines,
			fmt.Sprintf("AUTH_TOKEN_TYPE=%s", result.TokenType),
			fmt.Sprintf("AUTH_ACCESS_TOKEN=%s", result.AccessToken),
			fmt.Sprintf("AUTH_EXPIRES_IN=%d", result.ExpiresIn),
			fmt.Sprintf("AUTH_REFRESH_BEFORE=%d", result.RefreshBefore),
			fmt.Sprintf("AUTH_USER_ID=%s", authContext.UserID),
			fmt.Sprintf("AUTH_SKILL_ID=%s", authContext.SkillID),
		)
	} else {
		lines = append(lines,
			fmt.Sprintf("AUTH_DENY_CODE=%s", result.DenyCode),
			fmt.Sprintf("AUTH_DENY_MESSAGE=%s", result.DenyMessage),
		)
	}
	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}
