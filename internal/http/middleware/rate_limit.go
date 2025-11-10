package middleware

import (
	"auxstream/internal/cache"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket rate limiting using Redis
type RateLimiter struct {
	cache       cache.Cache
	maxRequests int           // Maximum requests allowed
	window      time.Duration // Time window
}

// RateLimitConfig configures the rate limiter
type RateLimitConfig struct {
	MaxRequests int           // Maximum requests per window
	Window      time.Duration // Time window
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cache cache.Cache, config RateLimitConfig) *RateLimiter {
	if config.MaxRequests == 0 {
		config.MaxRequests = 100 // Default: 100 requests
	}
	if config.Window == 0 {
		config.Window = 1 * time.Minute // Default: per minute
	}

	return &RateLimiter{
		cache:       cache,
		maxRequests: config.MaxRequests,
		window:      config.Window,
	}
}

// Middleware returns a Gin middleware that enforces rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address or user ID)
		identifier := rl.getClientIdentifier(c)

		// Check rate limit
		allowed, remaining, resetTime, err := rl.checkLimit(c.Request.Context(), identifier)
		if err != nil {
			// Log error but don't block request
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkLimit checks if the client is within rate limits
func (rl *RateLimiter) checkLimit(ctx context.Context, identifier string) (allowed bool, remaining int, resetTime time.Time, err error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)

	// Try to get current count
	count, err := rl.cache.Incr(ctx, key)
	if err != nil {
		return true, rl.maxRequests, time.Now().Add(rl.window), err
	}

	// If this is the first request, set the expiration
	if count == 1 {
		if err := rl.cache.Expire(ctx, key, rl.window); err != nil {
			// Log error but continue
		}
	}

	// Get TTL to determine reset time
	ttl, err := rl.cache.TTL(ctx, key)
	if err != nil {
		ttl = rl.window
	}
	resetTime = time.Now().Add(ttl)

	// Check if limit exceeded
	if count > int64(rl.maxRequests) {
		return false, 0, resetTime, nil
	}

	remaining = rl.maxRequests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetTime, nil
}

// getClientIdentifier extracts a unique identifier for the client
func (rl *RateLimiter) getClientIdentifier(c *gin.Context) string {
	// Try to get user ID from context (if authenticated)
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}

	// Fall back to IP address
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// GetStatus returns the current rate limit status for a client
func (rl *RateLimiter) GetStatus(ctx context.Context, identifier string) (map[string]interface{}, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)

	// Get current count
	countStr, err := rl.cache.GetString(key)
	if err != nil {
		// No entries found, client is clean
		return map[string]interface{}{
			"requests":  0,
			"limit":     rl.maxRequests,
			"remaining": rl.maxRequests,
			"reset_at":  time.Now().Add(rl.window),
		}, nil
	}

	var count int
	fmt.Sscanf(countStr, "%d", &count)

	// Get TTL
	ttl, err := rl.cache.TTL(ctx, key)
	if err != nil {
		ttl = rl.window
	}

	remaining := rl.maxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return map[string]interface{}{
		"requests":  count,
		"limit":     rl.maxRequests,
		"remaining": remaining,
		"reset_at":  time.Now().Add(ttl),
	}, nil
}

// Reset clears the rate limit for a specific client
func (rl *RateLimiter) Reset(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	return rl.cache.Del(key)
}

// StreamRateLimiter is a specialized rate limiter for streaming endpoints
type StreamRateLimiter struct {
	*RateLimiter
}

// NewStreamRateLimiter creates a rate limiter specifically for streaming
func NewStreamRateLimiter(cache cache.Cache) *StreamRateLimiter {
	return &StreamRateLimiter{
		RateLimiter: NewRateLimiter(cache, RateLimitConfig{
			MaxRequests: 50,              // 50 stream requests
			Window:      1 * time.Minute, // per minute
		}),
	}
}
