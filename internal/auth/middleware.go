package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const UserContextKey = "user"

// JWTAuthMiddleware validates JWT tokens and sets user context
func (j *JWTService) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := j.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user context
		ctx := context.WithValue(c.Request.Context(), UserContextKey, claims)
		c.Request = c.Request.WithContext(ctx)

		// Continue to next handler
		c.Next()
	}
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(c *gin.Context) (*JWTClaims, bool) {
	claims, ok := c.Request.Context().Value(UserContextKey).(*JWTClaims)
	return claims, ok
}
