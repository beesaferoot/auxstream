package handlers

import (
	"auxstream/internal/auth"
	"auxstream/internal/cache"
	"auxstream/internal/db"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// oauthStateTTL bounds how long a pending OAuth login may stay open before the
// generated state token is rejected as stale.
const oauthStateTTL = 10 * time.Minute

type AuthService struct {
	userRepo       db.UserRepo
	jwtService     *auth.JWTService
	refreshService *auth.RefreshTokenService
	oauthService   *auth.OAuthService
	cache          cache.Cache
}

func NewAuthService(userRepo db.UserRepo, jwtService *auth.JWTService, refreshService *auth.RefreshTokenService, oauthService *auth.OAuthService, cache cache.Cache) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtService:     jwtService,
		refreshService: refreshService,
		oauthService:   oauthService,
		cache:          cache,
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

// Register handles user registration with email and password.
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
		Provider:     "local",
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

// Login handles user login with email and password.
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

// RefreshToken handles token refresh requests.
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

// Logout handles user logout and token revocation.
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

// GoogleAuth initiates the Google OAuth flow.
func (a *AuthService) GoogleAuth(c *gin.Context) {
	state := uuid.New().String()

	// Persist the state so the callback can prove this request originated here,
	// protecting the OAuth flow against CSRF.
	if err := a.cache.Set(oauthStateKey(state), true, oauthStateTTL); err != nil {
		log.Printf("store oauth state error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start authentication"})
		return
	}

	authURL := a.oauthService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GoogleCallback handles the Google OAuth callback.
func (a *AuthService) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	// Verify the state issued by GoogleAuth before trusting this callback.
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter not provided"})
		return
	}
	exists, err := a.cache.Exists(c.Request.Context(), oauthStateKey(state))
	if err != nil || !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired state"})
		return
	}
	// State is single-use; consume it so it cannot be replayed.
	_ = a.cache.Del(oauthStateKey(state))

	token, err := a.oauthService.ExchangeCodeForToken(c.Request.Context(), code)
	if err != nil {
		log.Printf("ExchangeCodeForToken error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	googleUser, err := a.oauthService.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		log.Printf("GetUserInfo error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info from Google"})
		return
	}

	user, err := a.oauthService.FindOrCreateUser(c.Request.Context(), googleUser)
	if err != nil {
		log.Printf("FindOrCreateUser error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create or find user"})
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
		"message": "Google authentication successful",
		"data":    response,
	})
}

func oauthStateKey(state string) string {
	return fmt.Sprintf("oauth_state:%s", state)
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
