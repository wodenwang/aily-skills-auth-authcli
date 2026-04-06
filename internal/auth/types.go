package auth

type AuthContext struct {
	UserID  string `json:"user_id"`
	SkillID string `json:"skill_id"`
}

type CheckRequest struct {
	UserID  string         `json:"user_id"`
	SkillID string         `json:"skill_id"`
	Context map[string]any `json:"context,omitempty"`
}

type CheckResponse struct {
	RequestID     string         `json:"request_id"`
	Allowed       bool           `json:"allowed"`
	AccessToken   string         `json:"access_token,omitempty"`
	TokenType     string         `json:"token_type,omitempty"`
	ExpiresIn     int            `json:"expires_in,omitempty"`
	RefreshBefore int            `json:"refresh_before,omitempty"`
	Permissions   []string       `json:"permissions,omitempty"`
	DataScope     map[string]any `json:"data_scope,omitempty"`
	CacheTTL      int            `json:"cache_ttl,omitempty"`
	DenyCode      string         `json:"deny_code,omitempty"`
	DenyMessage   string         `json:"deny_message,omitempty"`
}

type RefreshRequest struct {
	Token string `json:"token"`
}

type RefreshResponse struct {
	AccessToken    string `json:"access_token"`
	ExpiresIn      int    `json:"expires_in"`
	RefreshBefore  int    `json:"refresh_before"`
	OldTokenStatus string `json:"old_token_status"`
	FailureCode    string `json:"failure_code"`
}

type Result struct {
	OK            bool         `json:"ok"`
	RequestID     string       `json:"request_id"`
	Allowed       bool         `json:"allowed"`
	TokenType     string       `json:"token_type,omitempty"`
	AccessToken   string       `json:"access_token,omitempty"`
	ExpiresIn     int          `json:"expires_in,omitempty"`
	RefreshBefore int          `json:"refresh_before,omitempty"`
	CacheHit      bool         `json:"cache_hit,omitempty"`
	DenyCode      string       `json:"deny_code,omitempty"`
	DenyMessage   string       `json:"deny_message,omitempty"`
	AuthContext   *AuthContext `json:"auth_context,omitempty"`
}
