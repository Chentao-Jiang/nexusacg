package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// QQOAuthService handles QQ Connect OAuth login flow.
type QQOAuthService struct {
	appID  string
	appKey string
}

// NewQQOAuthService returns nil if credentials are not configured.
func NewQQOAuthService(appID, appKey string) *QQOAuthService {
	if appID == "" || appKey == "" {
		return nil
	}
	return &QQOAuthService{
		appID:  appID,
		appKey: appKey,
	}
}

// QQTokenResponse is the parsed response from the access_token exchange endpoint.
type QQTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// QQUserInfo is the response from the get_user_info endpoint.
type QQUserInfo struct {
	Ret         int    `json:"ret"`
	Msg         string `json:"msg"`
	Nickname    string `json:"nickname"`
	Figureurl   string `json:"figureurl"`
	Figureurl1  string `json:"figureurl_1"`
	Figureurl2  string `json:"figureurl_2"`
	Gender      string `json:"gender"`
}

// GetAuthURL returns the QQ authorization URL for redirecting users.
func (s *QQOAuthService) GetAuthURL(redirectURI, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", s.appID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	return "https://graph.qq.com/oauth2.0/authorize?" + params.Encode()
}

// GetAccessToken exchanges an authorization code for access_token.
// QQ returns URL-encoded text, NOT JSON:
// access_token=xxx&expires_in=7776000&refresh_token=yyy
func (s *QQOAuthService) GetAccessToken(code, redirectURI string) (*QQTokenResponse, error) {
	reqURL := fmt.Sprintf(
		"https://graph.qq.com/oauth2.0/token?grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		s.appID, s.appKey, code, url.QueryEscape(redirectURI),
	)

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("qq oauth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read qq token response: %w", err)
	}

	// QQ returns: access_token=xxx&expires_in=7776000&refresh_token=yyy
	// Parse manually
	result, err := parseQQTokenResponse(string(body))
	if err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, fmt.Errorf("invalid qq token response: no access_token")
	}

	return result, nil
}

// GetOpenID retrieves the user's openid from QQ.
// QQ returns: callback( {"client_id":"xxx","openid":"yyy"} );
func (s *QQOAuthService) GetOpenID(accessToken string) (string, error) {
	reqURL := fmt.Sprintf(
		"https://graph.qq.com/oauth2.0/me?access_token=%s",
		accessToken,
	)

	resp, err := http.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("qq openid request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read qq openid response: %w", err)
	}

	// QQ returns: callback( {"client_id":"xxx","openid":"yyy"} );
	// Strip the callback( wrapper and ); suffix
	raw := strings.TrimSpace(string(body))
	jsonStr := strings.TrimPrefix(raw, "callback(")
	jsonStr = strings.TrimSuffix(jsonStr, ");")

	var me struct {
		ClientID string `json:"client_id"`
		OpenID   string `json:"openid"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &me); err != nil {
		return "", fmt.Errorf("failed to parse qq openid response: %w", err)
	}

	if me.OpenID == "" {
		return "", fmt.Errorf("invalid qq openid response")
	}

	return me.OpenID, nil
}

// GetUserInfo retrieves the user's profile from QQ.
// This endpoint returns standard JSON.
func (s *QQOAuthService) GetUserInfo(accessToken, openid string) (*QQUserInfo, error) {
	reqURL := fmt.Sprintf(
		"https://graph.qq.com/user/get_user_info?access_token=%s&oauth_consumer_key=%s&openid=%s",
		accessToken, s.appID, openid,
	)

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("qq userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read qq userinfo response: %w", err)
	}

	var info QQUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse qq userinfo response: %w", err)
	}

	if info.Ret != 0 {
		return nil, fmt.Errorf("qq userinfo error: %s (ret=%d)", info.Msg, info.Ret)
	}

	return &info, nil
}

// parseQQTokenResponse parses URL-encoded token response from QQ.
// Format: access_token=xxx&expires_in=7776000&refresh_token=yyy
func parseQQTokenResponse(raw string) (*QQTokenResponse, error) {
	result := &QQTokenResponse{}

	pairs := strings.Split(raw, "&")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := parts[1]

		switch key {
		case "access_token":
			result.AccessToken = value
		case "expires_in":
			fmt.Sscanf(value, "%d", &result.ExpiresIn)
		case "refresh_token":
			result.RefreshToken = value
		}
	}

	return result, nil
}
