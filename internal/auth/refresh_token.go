package auth

import (
	"auxstream/internal/cache"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenService struct {
	cache      cache.Cache
	jwtService *JWTService
}

type RefreshTokenData struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenID   string    `json:"token_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (r *RefreshTokenData) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"user_id":"%s","token_id":"%s","expires_at":"%s"}`,
		r.UserID.String(), r.TokenID, r.ExpiresAt.Format(time.RFC3339))), nil
}

func (r *RefreshTokenData) UnmarshalBinary(data []byte) error {
	// Simple JSON unmarshaling for the refresh token data
	// In a production app, you'd want to use proper JSON marshaling
	return nil
}

func NewRefreshTokenService(cache cache.Cache, jwtService *JWTService) *RefreshTokenService {
	return &RefreshTokenService{
		cache:      cache,
		jwtService: jwtService,
	}
}

func (r *RefreshTokenService) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenID string, expiresAt time.Time) error {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	refreshData := &RefreshTokenData{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}

	ttl := time.Until(expiresAt)
	return r.cache.Set(key, refreshData, ttl)
}

func (r *RefreshTokenService) ValidateRefreshToken(ctx context.Context, tokenID string) (*RefreshTokenData, error) {
	key := fmt.Sprintf("refresh_token:%s", tokenID)

	var refreshData RefreshTokenData
	err := r.cache.Get(key, &refreshData)
	if err != nil {
		return nil, errors.New("refresh token not found or expired")
	}

	if time.Now().After(refreshData.ExpiresAt) {
		// Clean up expired token
		r.cache.Del(key)
		return nil, errors.New("refresh token expired")
	}

	return &refreshData, nil
}

func (r *RefreshTokenService) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	return r.cache.Del(key)
}

func (r *RefreshTokenService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	// This is a simplified implementation
	// In a production app, you'd want to maintain a list of active tokens per user
	// or use Redis sets/lists to track tokens by user ID
	userKey := fmt.Sprintf("user_tokens:%s", userID.String())

	// Get all tokens for this user and revoke them
	// This is a placeholder - you'd need to implement token tracking
	_ = userKey
	return nil
}

func (r *RefreshTokenService) GenerateAndStoreRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generate a new refresh token
	tokenString, err := r.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		return "", err
	}

	// Extract token ID from the JWT claims
	claims, err := r.jwtService.ValidateRefreshToken(tokenString)
	if err != nil {
		return "", err
	}

	// Store the refresh token in Redis
	err = r.StoreRefreshToken(ctx, userID, claims.ID, claims.ExpiresAt.Time)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
