package auth

import "errors"

var ErrOAuthNotImplemented = errors.New("oauth provider not implemented")

// OAuthProviderStub keeps oauth integration points available for later wiring.
type OAuthProviderStub struct {
	provider string
}

func NewOAuthProvider(provider string) *OAuthProviderStub {
	if provider == "" {
		provider = "oauth"
	}
	return &OAuthProviderStub{provider: provider}
}

func (p *OAuthProviderStub) Name() string {
	return p.provider
}

func (p *OAuthProviderStub) BeginAuthURL(string) string {
	return ""
}

func (p *OAuthProviderStub) ExchangeCode(string) (User, error) {
	return User{}, ErrOAuthNotImplemented
}
