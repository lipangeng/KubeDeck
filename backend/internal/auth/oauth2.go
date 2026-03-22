package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// OAuth2Config holds OAuth2 provider configuration
type OAuth2Config struct {
	Provider     string   `json:"provider"` // github, google, gitlab, oidc
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`

	// OIDC specific
	IssuerURL string `json:"issuer_url"`
}

// OAuth2Provider manages OAuth2 authentication
type OAuth2Provider struct {
	config   *OAuth2Config
	oauthCfg *oauth2.Config
}

// UserInfo represents authenticated user information
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// NewOAuth2Provider creates a new OAuth2 provider
func NewOAuth2Provider(cfg *OAuth2Config) (*OAuth2Provider, error) {
	endpoint := getOAuth2Endpoint(cfg.Provider)

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint:     endpoint,
	}

	return &OAuth2Provider{
		config:   cfg,
		oauthCfg: oauthCfg,
	}, nil
}

func getOAuth2Endpoint(provider string) oauth2.Endpoint {
	switch provider {
	case "github":
		return oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		}
	case "google":
		return oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		}
	case "gitlab":
		return oauth2.Endpoint{
			AuthURL:  "https://gitlab.com/oauth/authorize",
			TokenURL: "https://gitlab.com/oauth/token",
		}
	default:
		return oauth2.Endpoint{}
	}
}

// AuthCodeURL generates the OAuth2 authorization URL
func (p *OAuth2Provider) AuthCodeURL(state string) string {
	return p.oauthCfg.AuthCodeURL(state)
}

// Exchange exchanges authorization code for access token
func (p *OAuth2Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.oauthCfg.Exchange(ctx, code)
}

// GetUserInfo retrieves user information from OAuth2 provider
func (p *OAuth2Provider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	switch p.config.Provider {
	case "github":
		return p.getGitHubUserInfo(ctx, token)
	case "google":
		return p.getGoogleUserInfo(ctx, token)
	case "gitlab":
		return p.getGitLabUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", p.config.Provider)
	}
}

func (p *OAuth2Provider) getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var githubUser struct {
		Login string `json:"login"`
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, err
	}

	// If email is not public, fetch from emails endpoint
	if githubUser.Email == "" {
		githubUser.Email, _ = p.getGitHubPrimaryEmail(ctx, client)
	}

	return &UserInfo{
		ID:       githubUser.ID,
		Username: githubUser.Login,
		Email:    githubUser.Email,
		Name:     githubUser.Name,
	}, nil
}

func (p *OAuth2Provider) getGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no verified primary email found")
}

func (p *OAuth2Provider) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:       googleUser.ID,
		Username: googleUser.Email,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
	}, nil
}

func (p *OAuth2Provider) getGitLabUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://gitlab.com/api/v4/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gitlabUser struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gitlabUser); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:       fmt.Sprintf("%d", gitlabUser.ID),
		Username: gitlabUser.Username,
		Email:    gitlabUser.Email,
		Name:     gitlabUser.Name,
	}, nil
}
