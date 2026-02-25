package auth

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrOAuthNotImplemented = errors.New("oauth provider not implemented")
var ErrOAuthInvalidCode = errors.New("oauth invalid code")

// OAuthProviderStub keeps oauth integration points available for later wiring.
type OAuthProviderStub struct {
	provider string
	baseURL  string
}

func NewOAuthProvider(provider string, baseURL string) *OAuthProviderStub {
	if provider == "" {
		provider = "oauth"
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://example.com/oauth/authorize"
	}
	return &OAuthProviderStub{
		provider: provider,
		baseURL:  strings.TrimSpace(baseURL),
	}
}

func (p *OAuthProviderStub) Name() string {
	return p.provider
}

func (p *OAuthProviderStub) BeginAuthURL(state string) string {
	values := url.Values{}
	values.Set("provider", p.provider)
	values.Set("state", state)
	return fmt.Sprintf("%s?%s", p.baseURL, values.Encode())
}

func (p *OAuthProviderStub) ExchangeCode(code string) (User, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return User{}, ErrOAuthInvalidCode
	}

	switch code {
	case "oauth-admin":
		return User{
			ID:                "oauth-admin-user",
			Username:          "oauth-admin",
			Roles:             []string{"admin"},
			AllowedClusters:   []string{"*"},
			AllowedNamespaces: []string{"*"},
		}, nil
	case "oauth-owner":
		return User{
			ID:                "oauth-owner-user",
			Username:          "oauth-owner",
			Roles:             []string{"owner"},
			AllowedClusters:   []string{"*"},
			AllowedNamespaces: []string{"*"},
		}, nil
	case "oauth-viewer":
		return User{
			ID:                "oauth-viewer-user",
			Username:          "oauth-viewer",
			Roles:             []string{"viewer"},
			AllowedClusters:   []string{"*"},
			AllowedNamespaces: []string{"*"},
		}, nil
	default:
		return User{}, ErrOAuthNotImplemented
	}
}
