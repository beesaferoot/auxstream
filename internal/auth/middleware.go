package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const UserContextKey = "user"

// JWTAuthMiddleware validates JWT tokens and sets user claims in context.
func (j *JWTService) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims, err := j.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		ctx := context.WithValue(c.Request.Context(), UserContextKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// OptionalJWTAuthMiddleware sets user claims in context when a valid Bearer token is
// present, but — unlike JWTAuthMiddleware — never aborts. Routes that are readable by
// anonymous visitors (e.g. a public/shared playlist) use this and then decide access
// themselves via GetUserFromContext.
func (j *JWTService) OptionalJWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if authHeader != "" && tokenString != authHeader {
			if claims, err := j.ValidateAccessToken(tokenString); err == nil {
				ctx := context.WithValue(c.Request.Context(), UserContextKey, claims)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		c.Next()
	}
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(c *gin.Context) (*JWTClaims, bool) {
	claims, ok := c.Request.Context().Value(UserContextKey).(*JWTClaims)
	return claims, ok
}
