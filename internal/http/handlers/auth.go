package handlers

import (
	"auxstream/internal/auth"
	"auxstream/internal/db"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo       db.UserRepo
	jwtService     *auth.JWTService
	refreshService *auth.RefreshTokenService
}

func NewAuthService(userRepo db.UserRepo, jwtService *auth.JWTService, refreshService *auth.RefreshTokenService) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtService:     jwtService,
		refreshService: refreshService,
	}
}

// accessTokenExpiresIn reports the access-token lifetime in seconds for token
// responses, derived from the JWT service so it can never drift from reality.
func (a *AuthService) accessTokenExpiresIn() int64 {
	return int64(a.jwtService.AccessTokenTTL().Seconds())
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register creates a user from a JSON email/password body (password min 6),
// returning 409 if the email is taken and 201 with a fresh token pair otherwise.
func (a *AuthService) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := a.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Password hash failure: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	user := &db.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		ID:           uuid.New(),
	}

	createdUser, err := a.userRepo.CreateUser(c.Request.Context(), user)
	if err != nil {
		log.Printf("CreateUser error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	accessToken, err := a.jwtService.GenerateAccessToken(createdUser.ID, createdUser.Email)
	if err != nil {
		log.Printf("GenerateAccessToken error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := a.refreshService.GenerateAndStoreRefreshToken(c.Request.Context(), createdUser.ID)
	if err != nil {
		log.Printf("GenerateRefreshToken error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	response := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    a.accessTokenExpiresIn(),
		TokenType:    "Bearer",
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"data":    response,
	})
}

// Login verifies a JSON email/password body and returns a token pair. Unknown
// email and bad password are reported identically as 401 to avoid user enumeration.
func (a *AuthService) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := a.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !cmpHashString(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, err := a.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		log.Printf("GenerateAccessToken error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := a.refreshService.GenerateAndStoreRefreshToken(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("GenerateRefreshToken error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	response := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    a.accessTokenExpiresIn(),
		TokenType:    "Bearer",
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    response,
	})
}

// RefreshToken exchanges a valid, still-stored refresh token (JSON body) for a
// new access token. The refresh token itself is not rotated. Any validation,
// lookup, or unknown-user failure returns 401.
func (a *AuthService) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := a.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
		return
	}

	if _, err = a.refreshService.ValidateRefreshToken(c.Request.Context(), claims.ID); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found or expired"})
		return
	}

	user, err := a.userRepo.GetUserById(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	accessToken, err := a.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		log.Printf("GenerateAccessToken error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	response := TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   a.accessTokenExpiresIn(),
		TokenType:   "Bearer",
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"data":    response,
	})
}

// Logout revokes the refresh token supplied in the JSON body, invalidating it
// server-side; outstanding access tokens remain valid until they expire.
func (a *AuthService) Logout(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := a.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refresh token"})
		return
	}

	err = a.refreshService.RevokeRefreshToken(c.Request.Context(), claims.ID)
	if err != nil {
		log.Printf("RevokeRefreshToken error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func cmpHashString(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
