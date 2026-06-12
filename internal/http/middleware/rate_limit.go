package middleware

import (
	"auxstream/internal/cache"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter enforces a fixed-window request limit per client, backed by a
// shared cache (Redis) so the limit holds across server instances.
type RateLimiter struct {
	cache       cache.Cache
	maxRequests int
	window      time.Duration
}

// RateLimitConfig configures the rate limiter. A zero field falls back to the
// defaults applied in NewRateLimiter.
type RateLimitConfig struct {
	MaxRequests int
	Window      time.Duration
}

// NewRateLimiter builds a RateLimiter, defaulting to 100 requests per minute
// when either limit is left zero.
func NewRateLimiter(cache cache.Cache, config RateLimitConfig) *RateLimiter {
	if config.MaxRequests == 0 {
		config.MaxRequests = 100
	}
	if config.Window == 0 {
		config.Window = 1 * time.Minute
	}

	return &RateLimiter{
		cache:       cache,
		maxRequests: config.MaxRequests,
		window:      config.Window,
	}
}

// Middleware enforces the limit per client and sets X-RateLimit-* headers.
// Over-limit requests are aborted with 429 plus a Retry-After header. A cache
// failure fails open (request proceeds) rather than locking users out.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := rl.getClientIdentifier(c)

		allowed, remaining, resetTime, err := rl.checkLimit(c.Request.Context(), identifier)
		if err != nil {
			c.Next()
			return
		}

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

// checkLimit increments the client's window counter and reports whether it is
// still within budget. On the first request of a window (count == 1) it sets
// the key's TTL, which both bounds the window and lets the key self-expire to
// reset the count; the TTL also yields the reset time reported to the client.
func (rl *RateLimiter) checkLimit(ctx context.Context, identifier string) (allowed bool, remaining int, resetTime time.Time, err error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)

	count, err := rl.cache.Incr(ctx, key)
	if err != nil {
		return true, rl.maxRequests, time.Now().Add(rl.window), err
	}

	if count == 1 {
		if err := rl.cache.Expire(ctx, key, rl.window); err != nil {
			// Best-effort: without a TTL the key would never expire, but a
			// failure here shouldn't reject the request.
		}
	}

	ttl, err := rl.cache.TTL(ctx, key)
	if err != nil {
		ttl = rl.window
	}
	resetTime = time.Now().Add(ttl)

	if count > int64(rl.maxRequests) {
		return false, 0, resetTime, nil
	}

	remaining = rl.maxRequests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetTime, nil
}

// getClientIdentifier keys the limit on the authenticated user when present,
// falling back to client IP for anonymous requests.
func (rl *RateLimiter) getClientIdentifier(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}

	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// GetStatus reports the client's current usage without consuming budget. A
// missing key means no requests this window, so the full limit is reported.
func (rl *RateLimiter) GetStatus(ctx context.Context, identifier string) (map[string]any, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)

	countStr, err := rl.cache.GetString(key)
	if err != nil {
		return map[string]any{
			"requests":  0,
			"limit":     rl.maxRequests,
			"remaining": rl.maxRequests,
			"reset_at":  time.Now().Add(rl.window),
		}, nil
	}

	var count int
	fmt.Sscanf(countStr, "%d", &count)

	ttl, err := rl.cache.TTL(ctx, key)
	if err != nil {
		ttl = rl.window
	}

	remaining := rl.maxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return map[string]any{
		"requests":  count,
		"limit":     rl.maxRequests,
		"remaining": remaining,
		"reset_at":  time.Now().Add(ttl),
	}, nil
}

// Reset clears the current window for a client, restoring full budget.
func (rl *RateLimiter) Reset(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	return rl.cache.Del(key)
}

// StreamRateLimiter is a RateLimiter preconfigured for streaming endpoints.
type StreamRateLimiter struct {
	*RateLimiter
}

// NewStreamRateLimiter builds a limiter allowing 50 stream requests per minute.
func NewStreamRateLimiter(cache cache.Cache) *StreamRateLimiter {
	return &StreamRateLimiter{
		RateLimiter: NewRateLimiter(cache, RateLimitConfig{
			MaxRequests: 50,
			Window:      1 * time.Minute,
		}),
	}
}
