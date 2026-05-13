package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ContentModerationService handles AI-based content moderation for posts and products.
// Integrates with external content safety APIs (e.g., Alibaba Cloud Content Security).
type ContentModerationService struct {
	apiKey    string
	apiSecret string
	apiURL    string
	db        *gorm.DB
}

func NewContentModerationService(db *gorm.DB, apiKey, apiSecret string) *ContentModerationService {
	return &ContentModerationService{
		db:        db,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		apiURL:    "https://green.cn-shanghai.aliyuncs.com/green/text/scan",
	}
}

// ModerationResult represents the result of content moderation.
type ModerationResult struct {
	Pass    bool     `json:"pass"`
	Labels  []string `json:"labels"`
	Reason  string   `json:"reason"`
}

// ModerateText checks text content for sensitive/illegal content using AI moderation API.
// When API credentials are not configured, falls back to local keyword filtering.
func (s *ContentModerationService) ModerateText(ctx context.Context, content string) (*ModerationResult, error) {
	// Fallback: local keyword filter when API not configured
	if s.apiKey == "" || s.apiSecret == "" {
		return s.localKeywordFilter(content)
	}

	// Call Alibaba Cloud Content Security API
	return s.callAlibabaAPI(ctx, content)
}

// ModerateImage checks an image URL for inappropriate content.
func (s *ContentModerationService) ModerateImage(ctx context.Context, imageURL string) (*ModerationResult, error) {
	if s.apiKey == "" || s.apiSecret == "" {
		// No API configured, pass by default
		return &ModerationResult{Pass: true}, nil
	}
	// Call Alibaba Cloud Image Moderation API
	return s.callAlibabaImageAPI(ctx, imageURL)
}

// AutoModeratePost checks a post's content and images before publication.
func (s *ContentModerationService) AutoModeratePost(ctx context.Context, title, content string, images []string) (*ModerationResult, error) {
	// Check text
	textResult, err := s.ModerateText(ctx, title+" "+content)
	if err != nil {
		return nil, fmt.Errorf("text moderation failed: %w", err)
	}
	if !textResult.Pass {
		return textResult, nil
	}

	// Check images
	for _, imgURL := range images {
		imgResult, err := s.ModerateImage(ctx, imgURL)
		if err != nil {
			log.Printf("image moderation failed for %s: %v", imgURL, err)
			continue
		}
		if !imgResult.Pass {
			return imgResult, nil
		}
	}

	return &ModerationResult{Pass: true}, nil
}

// --- Local keyword filter (fallback) ---

// Common sensitive words for Chinese ACG platforms
var sensitiveWords = []string{
	"色情", "赌博", "毒品", "暴力恐怖", "涉政", "法轮功",
	"fuck", "shit", "damn", // add more as needed
}

func (s *ContentModerationService) localKeywordFilter(content string) (*ModerationResult, error) {
	lower := strings.ToLower(content)
	for _, word := range sensitiveWords {
		if strings.Contains(lower, strings.ToLower(word)) {
			return &ModerationResult{
				Pass:   false,
				Labels: []string{"sensitive_word"},
				Reason: fmt.Sprintf("contains sensitive word: %s", word),
			}, nil
		}
	}
	return &ModerationResult{Pass: true}, nil
}

// --- Alibaba Cloud API calls ---

type alibabaTextTask struct {
	Content string `json:"content"`
}

type alibabaScanRequest struct {
	Tasks  []alibabaTextTask `json:"tasks"`
	Scenes []string          `json:"scenes"`
}

type alibabaScanResponse struct {
	Code int `json:"code"`
	Data []struct {
		Results []struct {
			Label  string  `json:"label"`
			Score  float64 `json:"rate"`
			Scene  string  `json:"scene"`
			Suggestion string `json:"suggestion"`
		} `json:"results"`
	} `json:"data"`
}

func (s *ContentModerationService) callAlibabaAPI(ctx context.Context, content string) (*ModerationResult, error) {
	reqBody, err := json.Marshal(alibabaScanRequest{
		Tasks:  []alibabaTextTask{{Content: content}},
		Scenes: []string{"antispam"},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	signature := s.buildAuthHeader(string(reqBody), timestamp)

	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-content-sha256", fmt.Sprintf("%x", sha256.Sum256(reqBody)))
	req.Header.Set("x-acs-date", timestamp)
	req.Header.Set("Authorization", signature)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result alibabaScanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Code != 200 || len(result.Data) == 0 {
		return &ModerationResult{Pass: false, Reason: "API error"}, nil
	}

	for _, r := range result.Data[0].Results {
		if r.Suggestion == "block" || r.Suggestion == "review" {
			return &ModerationResult{
				Pass:   false,
				Labels: []string{r.Label},
				Reason: fmt.Sprintf("%s (score: %.2f)", r.Label, r.Score),
			}, nil
		}
	}

	return &ModerationResult{Pass: true}, nil
}

func (s *ContentModerationService) callAlibabaImageAPI(ctx context.Context, imageURL string) (*ModerationResult, error) {
	type imageTask struct {
		DataID string `json:"dataId"`
		URL    string `json:"url"`
	}
	reqBody := struct {
		Tasks  []imageTask `json:"tasks"`
		Scenes []string    `json:"scenes"`
	}{
		Tasks:  []imageTask{{URL: imageURL}},
		Scenes: []string{"porn"},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	signature := s.buildAuthHeader(string(body), timestamp)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://green.cn-shanghai.aliyuncs.com/green/image/scan", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-content-sha256", fmt.Sprintf("%x", sha256.Sum256(body)))
	req.Header.Set("x-acs-date", timestamp)
	req.Header.Set("Authorization", signature)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("image API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result alibabaScanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Code != 200 || len(result.Data) == 0 {
		return &ModerationResult{Pass: false, Reason: "image API error"}, nil
	}

	for _, r := range result.Data[0].Results {
		if r.Suggestion == "block" || r.Suggestion == "review" {
			return &ModerationResult{
				Pass:   false,
				Labels: []string{r.Label},
				Reason: fmt.Sprintf("%s (score: %.2f)", r.Label, r.Score),
			}, nil
		}
	}

	return &ModerationResult{Pass: true}, nil
}

func (s *ContentModerationService) buildAuthHeader(body, timestamp string) string {
	// HMAC-SHA256 signature: sign(timestamp + body) with API secret
	message := timestamp + body
	mac := hmac.New(sha256.New, []byte(s.apiSecret))
	mac.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("HMAC-SHA256 Credential=%s, Signature=%s", s.apiKey, signature)
}
