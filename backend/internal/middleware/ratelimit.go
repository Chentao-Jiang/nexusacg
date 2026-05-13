package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter implements a simple sliding window rate limiter per IP.
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old entries
	var valid []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	// Add the current request first, then check if we exceed the limit
	valid = append(valid, now)
	if len(valid) > rl.limit {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = valid
	return true
}

// Separate rate limiters for auth (strict) and general API (relaxed).
var (
	authLimiter   = newRateLimiter(5, 1*time.Minute)  // 5 requests per minute
	apiLimiter    = newRateLimiter(100, 1*time.Minute) // 100 requests per minute
)

// RateLimit returns a middleware that limits requests per client IP.
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !apiLimiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// AuthRateLimit is a stricter rate limiter for authentication endpoints.
func AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !authLimiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "登录尝试过于频繁，请 1 分钟后再试",
			})
			return
		}
		c.Next()
	}
}
