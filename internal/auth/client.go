package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client interface {
	Check(ctx context.Context, req CheckRequest) (CheckResponse, error)
	Refresh(ctx context.Context, token string) (RefreshResponse, error)
}

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string, client *http.Client) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

func (c *HTTPClient) Check(ctx context.Context, req CheckRequest) (CheckResponse, error) {
	var out CheckResponse
	if err := c.post(ctx, "/api/v1/auth/check", req, &out); err != nil {
		return CheckResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) Refresh(ctx context.Context, token string) (RefreshResponse, error) {
	var out RefreshResponse
	if err := c.post(ctx, "/api/v1/token/refresh", RefreshRequest{Token: token}, &out); err != nil {
		return RefreshResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) post(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return UpstreamError{Op: path, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return UpstreamError{Op: path, Err: fmt.Errorf("status %d", resp.StatusCode)}
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return decodeAPIError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id"`
}

func (e APIError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("api error status=%d", e.StatusCode)
	}
	return fmt.Sprintf("api error status=%d code=%s", e.StatusCode, e.Code)
}

type UpstreamError struct {
	Op  string
	Err error
}

func (e UpstreamError) Error() string {
	return fmt.Sprintf("upstream %s: %v", e.Op, e.Err)
}

func (e UpstreamError) Unwrap() error {
	return e.Err
}

func decodeAPIError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UpstreamError{Op: "decode_error", Err: fmt.Errorf("status %d", resp.StatusCode)}
	}

	var apiErr APIError
	apiErr.StatusCode = resp.StatusCode
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return UpstreamError{Op: "decode_error", Err: fmt.Errorf("status %d", resp.StatusCode)}
	}
	if apiErr.Code == "" {
		var authCheckDeny CheckResponse
		if err := json.Unmarshal(body, &authCheckDeny); err == nil && !authCheckDeny.Allowed && authCheckDeny.DenyCode != "" {
			return APIError{
				StatusCode: resp.StatusCode,
				Code:       authCheckDeny.DenyCode,
				Message:    authCheckDeny.DenyMessage,
				RequestID:  authCheckDeny.RequestID,
			}
		}
	}
	return apiErr
}
