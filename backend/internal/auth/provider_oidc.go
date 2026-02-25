package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	name      string
	config    oauth2.Config
	verifier  *oidc.IDTokenVerifier
	exchangeC context.Context
}

func NewOIDCProviderFromEnv() (*OIDCProvider, error) {
	issuer := strings.TrimSpace(os.Getenv("KUBEDECK_OIDC_ISSUER"))
	clientID := strings.TrimSpace(os.Getenv("KUBEDECK_OIDC_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("KUBEDECK_OIDC_CLIENT_SECRET"))
	redirectURL := strings.TrimSpace(os.Getenv("KUBEDECK_OIDC_REDIRECT_URL"))
	name := strings.TrimSpace(os.Getenv("KUBEDECK_OAUTH_PROVIDER"))
	if name == "" {
		name = "oidc"
	}
	scopes := parseOIDCScopes(os.Getenv("KUBEDECK_OIDC_SCOPES"))
	return NewOIDCProvider(name, issuer, clientID, clientSecret, redirectURL, scopes)
}

func NewOIDCProvider(
	name string,
	issuer string,
	clientID string,
	clientSecret string,
	redirectURL string,
	scopes []string,
) (*OIDCProvider, error) {
	if strings.TrimSpace(issuer) == "" {
		return nil, errors.New("oidc issuer is required")
	}
	if strings.TrimSpace(clientID) == "" {
		return nil, errors.New("oidc client id is required")
	}
	if strings.TrimSpace(redirectURL) == "" {
		return nil, errors.New("oidc redirect url is required")
	}

	ctx := context.Background()
	oidcProvider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("init oidc provider: %w", err)
	}

	if name == "" {
		name = "oidc"
	}
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}
	hasOpenID := false
	for _, scope := range scopes {
		if scope == "openid" {
			hasOpenID = true
			break
		}
	}
	if !hasOpenID {
		scopes = append([]string{"openid"}, scopes...)
	}

	return &OIDCProvider{
		name: name,
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     oidcProvider.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       scopes,
		},
		verifier:  oidcProvider.Verifier(&oidc.Config{ClientID: clientID}),
		exchangeC: ctx,
	}, nil
}

func parseOIDCScopes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"openid", "profile", "email"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		scope := strings.TrimSpace(part)
		if scope == "" {
			continue
		}
		out = append(out, scope)
	}
	return out
}

func (p *OIDCProvider) Name() string {
	return p.name
}

func (p *OIDCProvider) BeginAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *OIDCProvider) ExchangeCode(code string) (User, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return User{}, ErrOAuthInvalidCode
	}

	token, err := p.config.Exchange(p.exchangeC, code)
	if err != nil {
		return User{}, fmt.Errorf("exchange oauth code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || strings.TrimSpace(rawIDToken) == "" {
		return User{}, errors.New("missing id_token in oauth token response")
	}

	idToken, err := p.verifier.Verify(p.exchangeC, rawIDToken)
	if err != nil {
		return User{}, fmt.Errorf("verify id_token: %w", err)
	}

	var claims struct {
		Sub               string   `json:"sub"`
		PreferredUsername string   `json:"preferred_username"`
		Email             string   `json:"email"`
		Name              string   `json:"name"`
		Roles             []string `json:"roles"`
		Groups            []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return User{}, fmt.Errorf("decode id_token claims: %w", err)
	}

	userID := strings.TrimSpace(claims.Sub)
	if userID == "" {
		userID = strings.TrimSpace(claims.Email)
	}
	if userID == "" {
		userID = strings.TrimSpace(claims.PreferredUsername)
	}
	if userID == "" {
		return User{}, errors.New("oidc claim subject is missing")
	}

	username := strings.TrimSpace(claims.PreferredUsername)
	if username == "" {
		username = strings.TrimSpace(claims.Email)
	}
	if username == "" {
		username = strings.TrimSpace(claims.Name)
	}
	if username == "" {
		username = userID
	}

	roles := claims.Roles
	if len(roles) == 0 {
		roles = claims.Groups
	}
	if len(roles) == 0 {
		roles = []string{"viewer"}
	}

	return User{
		ID:                userID,
		Username:          username,
		Roles:             roles,
		AllowedClusters:   []string{"*"},
		AllowedNamespaces: []string{"*"},
	}, nil
}
