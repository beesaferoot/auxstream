package auth

import (
	"auxstream/internal/cache"
	"context"
	"encoding/json"
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
	return json.Marshal(r)
}

func (r *RefreshTokenData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

func NewRefreshTokenService(cache cache.Cache, jwtService *JWTService) *RefreshTokenService {
	return &RefreshTokenService{
		cache:      cache,
		jwtService: jwtService,
	}
}

// StoreRefreshToken saves the refresh token with TTL matching its expiry.
func (r *RefreshTokenService) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenID string, expiresAt time.Time) error {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	refreshData := &RefreshTokenData{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("invalid expiration time for refresh token")
	}

	// Save token data
	if err := r.cache.Set(key, refreshData, ttl); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Track token under the user's list for easy revocation
	userKey := fmt.Sprintf("user_tokens:%s", userID.String())
	return r.cache.SAdd(ctx, userKey, tokenID)
}

// ValidateRefreshToken checks if the token is valid and not expired.
func (r *RefreshTokenService) ValidateRefreshToken(ctx context.Context, tokenID string) (*RefreshTokenData, error) {
	key := fmt.Sprintf("refresh_token:%s", tokenID)

	var refreshData RefreshTokenData
	err := r.cache.Get(key, &refreshData)
	if err != nil {
		return nil, errors.New("refresh token not found or expired")
	}

	if time.Now().After(refreshData.ExpiresAt) {
		// Clean up expired token
		_ = r.cache.Del(key)
		return nil, errors.New("refresh token expired")
	}

	return &refreshData, nil
}

// RevokeRefreshToken deletes a specific refresh token from the cache.
func (r *RefreshTokenService) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	return r.cache.Del(key)
}

// RevokeAllUserTokens removes all refresh tokens for a user.
func (r *RefreshTokenService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	userKey := fmt.Sprintf("user_tokens:%s", userID.String())

	tokenIDs, err := r.cache.SMembers(ctx, userKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve user tokens: %w", err)
	}

	for _, tokenID := range tokenIDs {
		_ = r.cache.Del(fmt.Sprintf("refresh_token:%s", tokenID))
	}

	// Delete the user's token tracking key
	return r.cache.Del(userKey)
}

// GenerateAndStoreRefreshToken creates a new refresh token and stores it in the cache.
func (r *RefreshTokenService) GenerateAndStoreRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generate a new refresh token
	tokenString, err := r.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Validate token to extract claims (ID, Expiry)
	claims, err := r.jwtService.ValidateRefreshToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Store the refresh token in cache
	err = r.StoreRefreshToken(ctx, userID, claims.ID, claims.ExpiresAt.Time)
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return tokenString, nil
}
