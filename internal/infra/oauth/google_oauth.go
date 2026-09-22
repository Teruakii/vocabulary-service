package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	"vocabulary-service/internal/constants"
)

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type GoogleClient struct {
	config *oauth2.Config
}

func NewGoogleClient(clientID, clientSecret, redirectURL string) *GoogleClient {
	return &GoogleClient{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googleoauth.Endpoint,
		},
	}
}

func (g *GoogleClient) AuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (g *GoogleClient) Exchange(ctx context.Context, code string) (*GoogleUserInfo, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", constants.ErrOAuthExchangeFailed, err)
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google userinfo status %d: %s", resp.StatusCode, body)
	}

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
	}
	if !info.EmailVerified {
		return nil, constants.ErrEmailNotVerified
	}
	return &info, nil
}
