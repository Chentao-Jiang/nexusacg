package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// WeChatOAuthService handles WeChat Open Platform OAuth login flow.
type WeChatOAuthService struct {
	appID     string
	appSecret string
}

// NewWeChatOAuthService returns nil if credentials are not configured.
func NewWeChatOAuthService(appID, appSecret string) *WeChatOAuthService {
	if appID == "" || appSecret == "" {
		return nil
	}
	return &WeChatOAuthService{
		appID:     appID,
		appSecret: appSecret,
	}
}

// WeChatTokenResponse is the response from the access_token exchange endpoint.
type WeChatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// WeChatUserInfo is the response from the userinfo endpoint.
type WeChatUserInfo struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadimgURL string `json:"headimgurl"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// GetAuthURL returns the WeChat authorization URL for redirecting users.
func (s *WeChatOAuthService) GetAuthURL(redirectURI, state string) string {
	params := url.Values{}
	params.Set("appid", s.appID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "snsapi_login")
	params.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + params.Encode() + "#wechat_redirect"
}

// GetAccessToken exchanges an authorization code for access_token and openid.
func (s *WeChatOAuthService) GetAccessToken(code string) (*WeChatTokenResponse, error) {
	reqURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		s.appID, s.appSecret, code,
	)

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("wechat oauth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read wechat token response: %w", err)
	}

	var tokenResp WeChatTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse wechat token response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("wechat token error: %s (code=%d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	if tokenResp.AccessToken == "" || tokenResp.OpenID == "" {
		return nil, fmt.Errorf("invalid wechat token response")
	}

	return &tokenResp, nil
}

// GetUserInfo retrieves the user's profile from WeChat.
func (s *WeChatOAuthService) GetUserInfo(accessToken, openid string) (*WeChatUserInfo, error) {
	reqURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		accessToken, openid,
	)

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("wechat userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read wechat userinfo response: %w", err)
	}

	var info WeChatUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse wechat userinfo response: %w", err)
	}

	if info.ErrCode != 0 {
		return nil, fmt.Errorf("wechat userinfo error: %s (code=%d)", info.ErrMsg, info.ErrCode)
	}

	if info.OpenID == "" {
		return nil, fmt.Errorf("invalid wechat userinfo response")
	}

	return &info, nil
}
