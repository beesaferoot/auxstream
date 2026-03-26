package auth

import (
	"auxstream/internal/db"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type OAuthService struct {
	config   *oauth2.Config
	userRepo db.UserRepo
}

func NewOAuthService(clientID, clientSecret, redirectURL string, userRepo db.UserRepo) *OAuthService {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &OAuthService{
		config:   config,
		userRepo: userRepo,
	}
}

func (o *OAuthService) GetAuthURL(state string) string {
	return o.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (o *OAuthService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return o.config.Exchange(ctx, code)
}

func (o *OAuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := o.config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// FindOrCreateUser locates a user by Google ID or email, creating one if neither match.
func (o *OAuthService) FindOrCreateUser(ctx context.Context, googleUser *GoogleUserInfo) (*db.User, error) {
	user, err := o.userRepo.GetUserByGoogleID(ctx, googleUser.ID)
	if err == nil {
		return user, nil
	}

	user, err = o.userRepo.GetUserByEmail(ctx, googleUser.Email)
	if err == nil {
		user.GoogleID = googleUser.ID
		user.Provider = "google"
		return o.userRepo.UpdateUser(ctx, user)
	}

	return o.userRepo.CreateUser(ctx, &db.User{
		Email:    googleUser.Email,
		GoogleID: googleUser.ID,
		Provider: "google",
	})
}
