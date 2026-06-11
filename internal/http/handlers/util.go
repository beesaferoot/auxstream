package handlers

import "github.com/gin-gonic/gin"

// contextKey is a private type for request-context keys, avoiding collisions
// with keys defined in other packages (a bare string key risks silent clashes).
type contextKey string

// CacheContextKey is where the request-scoped cache client is stored on the
// request context by the cache-injection middleware.
const CacheContextKey contextKey = "cacheClient"

func errorResponse(message string) gin.H {
	return gin.H{
		"error": message,
	}
}
