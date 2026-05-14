package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/planforever/nexusacg/internal/config"
	"github.com/planforever/nexusacg/internal/middleware"
	"github.com/planforever/nexusacg/internal/service"
)

type AuthHandler struct {
	svc         *service.AuthService
	smsSvc      *service.SMSService
	wechatOauth *service.WeChatOAuthService
	qqOauth     *service.QQOAuthService
	cfg         *config.Config
}

func NewAuthHandler(r *gin.RouterGroup, svc *service.AuthService, smsSvc *service.SMSService, wechatOauth *service.WeChatOAuthService, qqOauth *service.QQOAuthService, cfg *config.Config) {
	h := &AuthHandler{svc: svc, smsSvc: smsSvc, wechatOauth: wechatOauth, qqOauth: qqOauth, cfg: cfg}

	auth := r.Group("/auth")
	auth.Use(middleware.AuthRateLimit())
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/sms/send", h.SMSSendCode)
	auth.POST("/sms/login", h.SMSLogin)
	// WeChat OAuth routes — moderate rate limit to prevent abuse
	wechat := auth.Group("/wechat")
	wechat.Use(middleware.RateLimit())
	wechat.GET("/authorize", h.WeChatAuthorize)
	wechat.GET("/callback", h.WeChatCallback)
	// QQ OAuth routes — moderate rate limit to prevent abuse
	qq := auth.Group("/qq")
	qq.Use(middleware.RateLimit())
	qq.GET("/authorize", h.QQAuthorize)
	qq.GET("/callback", h.QQCallback)

	// Logout requires JWT auth
	logout := r.Group("/auth/logout")
	logout.Use(middleware.JWTAuth(cfg))
	logout.POST("", h.Logout)
}

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"omitempty,max=20"`
	Email    string `json:"email" binding:"omitempty,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
}

// Register godoc
// @Summary Register a new user
// @Description Register with phone/email, password, and nickname
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration info"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	user, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		BadRequest(c, safeErrorMessage(err))
		return
	}

	Success(c, user)
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

// Login godoc
// @Summary User login
// @Description Login with phone/email and password, returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tokens, err := h.svc.Login(c.Request.Context(), service.LoginInput{
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		Unauthorized(c, err.Error())
		return
	}

	Success(c, tokens)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh godoc
// @Summary Refresh access token
// @Description Use a refresh token to get a new access/refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	tokens, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		Unauthorized(c, err.Error())
		return
	}

	Success(c, tokens)
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Logout godoc
// @Summary Logout
// @Description Revoke the given refresh token (requires auth)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh token to revoke"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"status": "logged_out"})
}

type SMSSendCodeRequest struct {
	Phone string `json:"phone" binding:"required,max=20"`
}

// SMSSendCode godoc
// @Summary Send SMS verification code
// @Description Send a 6-digit verification code to the given phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SMSSendCodeRequest true "Phone number"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /auth/sms/send [post]
func (h *AuthHandler) SMSSendCode(c *gin.Context) {
	if h.smsSvc == nil {
		BadRequest(c, "SMS service not configured")
		return
	}

	var req SMSSendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if err := h.smsSvc.SendVerificationCode(req.Phone); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "验证码已发送"})
}

type SMSLoginRequest struct {
	Phone string `json:"phone" binding:"required,max=20"`
	Code  string `json:"code" binding:"required,len=6"`
}

// SMSLogin godoc
// @Summary SMS verification login
// @Description Verify SMS code and login. Auto-registers new users by phone.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SMSLoginRequest true "Phone number and verification code"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /auth/sms/login [post]
func (h *AuthHandler) SMSLogin(c *gin.Context) {
	if h.smsSvc == nil {
		BadRequest(c, "SMS service not configured")
		return
	}

	var req SMSLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if !h.smsSvc.Verify(req.Phone, req.Code) {
		Unauthorized(c, "验证码错误或已过期")
		return
	}

	tokens, err := h.svc.SMSLogin(c.Request.Context(), req.Phone)
	if err != nil {
		Unauthorized(c, err.Error())
		return
	}

	Success(c, tokens)
}

type WeChatAuthorizeResponse struct {
	RedirectURL string `json:"redirect_url"`
}

type QQAuthorizeResponse struct {
	RedirectURL string `json:"redirect_url"`
}

// WeChatAuthorize godoc
// @Summary Get WeChat OAuth authorization URL
// @Description Returns the redirect URL for WeChat Open Platform login.
// @Tags auth
// @Produce json
// @Param state query string false "State parameter (returned in callback)"
// @Success 200 {object} Response{data=WeChatAuthorizeResponse}
// @Failure 503 {object} Response
// @Router /auth/wechat/authorize [get]
func (h *AuthHandler) WeChatAuthorize(c *gin.Context) {
	if h.wechatOauth == nil {
		Error(c, http.StatusServiceUnavailable, "WeChat OAuth not configured")
		return
	}

	redirectURI := h.cfg.BaseURL + "/api/v1/auth/wechat/callback"
	authURL := h.wechatOauth.GetAuthURL(redirectURI, c.DefaultQuery("state", "default"))

	Success(c, WeChatAuthorizeResponse{RedirectURL: authURL})
}

// WeChatCallback godoc
// @Summary WeChat OAuth callback
// @Description Handles the WeChat authorization callback, exchanges code for tokens, and returns JWT tokens.
// @Tags auth
// @Produce json
// @Param code query string true "WeChat authorization code"
// @Param state query string false "State parameter from authorize request"
// @Success 200 {object} Response{data=service.TokenPair}
// @Failure 400 {object} Response
// @Failure 503 {object} Response
// @Router /auth/wechat/callback [get]
func (h *AuthHandler) WeChatCallback(c *gin.Context) {
	if h.wechatOauth == nil {
		Error(c, http.StatusServiceUnavailable, "WeChat OAuth not configured")
		return
	}

	code := c.Query("code")
	if code == "" {
		BadRequest(c, "missing authorization code")
		return
	}

	state := c.Query("state")
	if state == "" {
		BadRequest(c, "missing state parameter")
		return
	}

	user, tokens, err := h.svc.WeChatOAuthLogin(c.Request.Context(), h.wechatOauth, code)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			Error(c, http.StatusServiceUnavailable, err.Error())
		} else {
			BadRequest(c, err.Error())
		}
		return
	}

	_ = user // user is available; tokens are the main response
	Success(c, tokens)
}

// safeErrorMessage returns a user-safe error message, stripping internal details.
func safeErrorMessage(err error) string {
	msg := err.Error()
	for _, prefix := range []string{
		"phone already", "email already", "invalid credentials",
		"user account is not active", "请等待", "今日发送", "验证码",
	} {
		if strings.Contains(msg, prefix) {
			return msg
		}
	}
	return "操作失败，请重试"
}

// QQAuthorize godoc
// @Summary Get QQ OAuth authorization URL
// @Description Returns the redirect URL for QQ Connect login.
// @Tags auth
// @Produce json
// @Param state query string false "State parameter (returned in callback)"
// @Success 200 {object} Response{data=QQAuthorizeResponse}
// @Failure 503 {object} Response
// @Router /auth/qq/authorize [get]
func (h *AuthHandler) QQAuthorize(c *gin.Context) {
	if h.qqOauth == nil {
		Error(c, http.StatusServiceUnavailable, "QQ OAuth not configured")
		return
	}

	redirectURI := h.cfg.BaseURL + "/api/v1/auth/qq/callback"
	authURL := h.qqOauth.GetAuthURL(redirectURI, c.DefaultQuery("state", "default"))

	Success(c, QQAuthorizeResponse{RedirectURL: authURL})
}

// QQCallback godoc
// @Summary QQ OAuth callback
// @Description Handles the QQ authorization callback, exchanges code for tokens, and returns JWT tokens.
// @Tags auth
// @Produce json
// @Param code query string true "QQ authorization code"
// @Param state query string false "State parameter from authorize request"
// @Success 200 {object} Response{data=service.TokenPair}
// @Failure 400 {object} Response
// @Failure 503 {object} Response
// @Router /auth/qq/callback [get]
func (h *AuthHandler) QQCallback(c *gin.Context) {
	if h.qqOauth == nil {
		Error(c, http.StatusServiceUnavailable, "QQ OAuth not configured")
		return
	}

	code := c.Query("code")
	if code == "" {
		BadRequest(c, "missing authorization code")
		return
	}

	state := c.Query("state")
	if state == "" {
		BadRequest(c, "missing state parameter")
		return
	}

	redirectURI := h.cfg.BaseURL + "/api/v1/auth/qq/callback"
	user, tokens, err := h.svc.QQOAuthLogin(c.Request.Context(), h.qqOauth, code, redirectURI)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			Error(c, http.StatusServiceUnavailable, err.Error())
		} else {
			BadRequest(c, err.Error())
		}
		return
	}

	_ = user // user is available; tokens are the main response
	Success(c, tokens)
}
