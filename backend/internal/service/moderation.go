package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// ContentModerationService handles AI-based content moderation for posts and products.
// Uses DeepSeek V4 Flash for text moderation and Qwen3.5-Flash for image/video moderation.
type ContentModerationService struct {
	deepseekAPIKey string
	qwenAPIKey     string
}

func NewContentModerationService(deepseekKey, qwenKey string) *ContentModerationService {
	return &ContentModerationService{
		deepseekAPIKey: deepseekKey,
		qwenAPIKey:     qwenKey,
	}
}

// ModerationResult represents the result of content moderation.
type ModerationResult struct {
	Pass   bool     `json:"pass"`
	Labels []string `json:"labels"`
	Reason string   `json:"reason"`
}

// ModerateText checks text content for sensitive/illegal content using DeepSeek V4 Flash.
// Falls back to local keyword filtering when API key is not configured.
func (s *ContentModerationService) ModerateText(ctx context.Context, content string) (*ModerationResult, error) {
	if s.deepseekAPIKey == "" {
		return s.localKeywordFilter(content)
	}
	return s.callDeepSeekModeration(ctx, content)
}

// ModerateImage checks an image URL for inappropriate content using Qwen3-VL-flash.
// Falls back to pass-by-default when API key is not configured.
func (s *ContentModerationService) ModerateImage(ctx context.Context, imageURL string) (*ModerationResult, error) {
	if s.qwenAPIKey == "" {
		return &ModerationResult{Pass: true}, nil
	}
	return s.callQwenVLModeration(ctx, imageURL)
}

// ModerateVideo checks a video URL for inappropriate content using Qwen3-VL-flash.
// Falls back to pass-by-default when API key is not configured.
func (s *ContentModerationService) ModerateVideo(ctx context.Context, videoURL string) (*ModerationResult, error) {
	if s.qwenAPIKey == "" {
		log.Printf("moderation: video fallback (no Qwen API key), passing: %s", videoURL)
		return &ModerationResult{Pass: true}, nil
	}
	log.Printf("moderation: calling Qwen3-VL-flash for video: %s", videoURL)
	return s.callQwenVLModerationVideo(ctx, videoURL)
}

// AutoModeratePost checks a post's content, images, and video before publication.
func (s *ContentModerationService) AutoModeratePost(ctx context.Context, title, content string, images []string, videoURL string) (*ModerationResult, error) {
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

	// Check video
	if videoURL != "" {
		videoResult, err := s.ModerateVideo(ctx, videoURL)
		if err != nil {
			log.Printf("video moderation failed for %s: %v", videoURL, err)
		} else if !videoResult.Pass {
			return videoResult, nil
		}
	}

	return &ModerationResult{Pass: true}, nil
}

// --- DeepSeek text moderation ---

func (s *ContentModerationService) callDeepSeekModeration(ctx context.Context, content string) (*ModerationResult, error) {
	reqBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `你是内容审核助手。请判断给定内容是否包含以下违规类别：
- 色情/性暗示
- 暴力/恐怖
- 政治敏感
- 仇恨/歧视
- 毒品/非法交易
- 赌博
- 广告/垃圾营销
- 辱骂/人身攻击

以 JSON 格式返回：
{"pass": true/false, "categories": ["违规类别列表"], "reason": "简要原因"}`,
			},
			{"role": "user", "content": content},
		},
		"temperature": 0.1,
		"max_tokens":  200,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.deepseekAPIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DeepSeek API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("DeepSeek API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return &ModerationResult{Pass: false, Reason: "no response from moderation API"}, nil
	}

	// Parse JSON from response
	respContent := result.Choices[0].Message.Content
	return parseDeepSeekJSON(respContent)
}

// parseDeepSeekJSON extracts moderation result from the LLM's JSON response.
func parseDeepSeekJSON(raw string) (*ModerationResult, error) {
	start := strings.Index(raw, "{")
	if start == -1 {
		return &ModerationResult{Pass: true, Reason: "no violations detected"}, nil
	}
	end := strings.LastIndex(raw, "}")
	if end == -1 || end < start {
		return &ModerationResult{Pass: true, Reason: "no violations detected"}, nil
	}
	jsonStr := raw[start : end+1]

	var parsed struct {
		Pass       bool     `json:"pass"`
		Categories []string `json:"categories"`
		Reason     string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		log.Printf("failed to parse DeepSeek moderation JSON: %v, raw: %s", err, raw)
		return &ModerationResult{Pass: true, Reason: "moderation parse error, allowing"}, nil
	}

	if !parsed.Pass {
		return &ModerationResult{
			Pass:   false,
			Labels: parsed.Categories,
			Reason: parsed.Reason,
		}, nil
	}
	return &ModerationResult{Pass: true, Reason: parsed.Reason}, nil
}

// --- Qwen3.5-Flash image moderation (DashScope OpenAI-compatible API) ---

func (s *ContentModerationService) callQwenVLModeration(ctx context.Context, imageURL string) (*ModerationResult, error) {
	reqBody := map[string]interface{}{
		"model": "qwen3.5-flash",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "你是内容审核助手。请判断图片是否包含以下违规内容：色情/性暗示、暴力/血腥、政治敏感、仇恨/歧视、毒品/非法交易、赌博、自残/自杀倾向。以JSON格式返回：{\"pass\":true/false,\"categories\":[\"违规类别列表\"],\"reason\":\"简要原因\"}",
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type":     "image_url",
						"image_url": map[string]interface{}{"url": imageURL},
					},
					{
						"type": "text",
						"text": "请审核这张图片是否包含违规内容。",
					},
				},
			},
		},
		"max_tokens": 300,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.qwenAPIKey)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Qwen3-VL API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("Qwen3-VL API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return &ModerationResult{Pass: false, Reason: "no response from moderation API"}, nil
	}

	respContent := result.Choices[0].Message.Content
	return parseQwenJSON(respContent)
}

// callQwenVLModerationVideo moderates video content using Qwen3.5-Flash.
func (s *ContentModerationService) callQwenVLModerationVideo(ctx context.Context, videoURL string) (*ModerationResult, error) {
	reqBody := map[string]interface{}{
		"model": "qwen3.5-flash",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "你是内容审核助手。请判断视频是否包含以下违规内容：色情/性暗示、暴力/血腥、政治敏感、仇恨/歧视、毒品/非法交易、赌博、自残/自杀倾向。以JSON格式返回：{\"pass\":true/false,\"categories\":[\"违规类别列表\"],\"reason\":\"简要原因\"}",
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "video_url",
						"video_url": map[string]interface{}{"url": videoURL},
					},
					{
						"type": "text",
						"text": "请审核这段视频是否包含违规内容。",
					},
				},
			},
		},
		"max_tokens": 300,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.qwenAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Qwen3-VL video API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("Qwen3-VL video API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return &ModerationResult{Pass: false, Reason: "no response from video moderation API"}, nil
	}

	respContent := result.Choices[0].Message.Content
	log.Printf("moderation: Qwen video API raw response: %s", respContent)
	return parseQwenJSON(respContent)
}

func parseQwenJSON(raw string) (*ModerationResult, error) {
	start := strings.Index(raw, "{")
	if start == -1 {
		return &ModerationResult{Pass: true, Reason: "no violations detected"}, nil
	}
	end := strings.LastIndex(raw, "}")
	if end == -1 || end < start {
		return &ModerationResult{Pass: true, Reason: "no violations detected"}, nil
	}
	jsonStr := raw[start : end+1]

	var parsed struct {
		Pass       bool     `json:"pass"`
		Categories []string `json:"categories"`
		Reason     string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		log.Printf("failed to parse Qwen moderation JSON: %v, raw: %s", err, raw)
		return &ModerationResult{Pass: true, Reason: "moderation parse error, allowing"}, nil
	}

	if !parsed.Pass {
		return &ModerationResult{
			Pass:   false,
			Labels: parsed.Categories,
			Reason: parsed.Reason,
		}, nil
	}
	return &ModerationResult{Pass: true, Reason: parsed.Reason}, nil
}

// --- Local keyword filter (fallback) ---

var sensitiveWords = []string{
	"色情", "赌博", "毒品", "暴力恐怖", "涉政", "法轮功",
	"fuck", "shit", "damn",
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
