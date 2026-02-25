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
	claims    oidcClaimConfig
}

type oidcClaimConfig struct {
	subjectClaim       string
	usernameClaim      string
	roleClaims         []string
	roleMap            map[string]string
	defaultRole        string
	allowedRoles       map[string]struct{}
	requireAllowedRole bool
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
	claims := oidcClaimConfig{
		subjectClaim:       normalizeOrDefault(os.Getenv("KUBEDECK_OIDC_SUBJECT_CLAIM"), "sub"),
		usernameClaim:      normalizeOrDefault(os.Getenv("KUBEDECK_OIDC_USERNAME_CLAIM"), "preferred_username"),
		roleClaims:         parseOIDCRoleClaims(os.Getenv("KUBEDECK_OIDC_ROLE_CLAIMS")),
		roleMap:            parseOIDCRoleMap(os.Getenv("KUBEDECK_OIDC_ROLE_MAP")),
		defaultRole:        normalizeOrDefault(os.Getenv("KUBEDECK_OIDC_DEFAULT_ROLE"), "viewer"),
		allowedRoles:       parseOIDCAllowedRoles(os.Getenv("KUBEDECK_OIDC_ALLOWED_ROLES")),
		requireAllowedRole: parseEnvBool(os.Getenv("KUBEDECK_OIDC_REQUIRE_ALLOWED_ROLE")),
	}
	return NewOIDCProvider(name, issuer, clientID, clientSecret, redirectURL, scopes, claims)
}

func NewOIDCProvider(
	name string,
	issuer string,
	clientID string,
	clientSecret string,
	redirectURL string,
	scopes []string,
	claims oidcClaimConfig,
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
	if strings.TrimSpace(claims.subjectClaim) == "" {
		claims.subjectClaim = "sub"
	}
	if strings.TrimSpace(claims.usernameClaim) == "" {
		claims.usernameClaim = "preferred_username"
	}
	if len(claims.roleClaims) == 0 {
		claims.roleClaims = []string{"roles", "groups"}
	}
	if strings.TrimSpace(claims.defaultRole) == "" {
		claims.defaultRole = "viewer"
	}
	if claims.roleMap == nil {
		claims.roleMap = map[string]string{}
	}
	if claims.allowedRoles == nil {
		claims.allowedRoles = map[string]struct{}{}
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
		claims:    claims,
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

func parseOIDCRoleClaims(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"roles", "groups"}
	}
	items := strings.Split(raw, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return []string{"roles", "groups"}
	}
	return out
}

func parseOIDCRoleMap(raw string) map[string]string {
	out := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	return out
}

func parseOIDCAllowedRoles(raw string) map[string]struct{} {
	allowed := map[string]struct{}{}
	for _, item := range strings.Split(raw, ",") {
		role := strings.TrimSpace(item)
		if role == "" {
			continue
		}
		allowed[role] = struct{}{}
	}
	return allowed
}

func parseEnvBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func normalizeOrDefault(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
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

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return User{}, fmt.Errorf("decode id_token claims: %w", err)
	}
	userID, username, roles, err := extractOIDCIdentity(claims, p.claims)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:                userID,
		Username:          username,
		Roles:             roles,
		AllowedClusters:   []string{"*"},
		AllowedNamespaces: []string{"*"},
	}, nil
}

func extractOIDCIdentity(
	claims map[string]any,
	config oidcClaimConfig,
) (userID string, username string, roles []string, err error) {
	userID = claimString(claims, config.subjectClaim)
	if userID == "" {
		userID = claimString(claims, "sub")
	}
	if userID == "" {
		userID = claimString(claims, "email")
	}
	if userID == "" {
		userID = claimString(claims, "preferred_username")
	}
	if userID == "" {
		return "", "", nil, errors.New("oidc claim subject is missing")
	}

	username = claimString(claims, config.usernameClaim)
	if username == "" {
		username = claimString(claims, "preferred_username")
	}
	if username == "" {
		username = claimString(claims, "email")
	}
	if username == "" {
		username = claimString(claims, "name")
	}
	if username == "" {
		username = userID
	}

	roleValues := make([]string, 0, 4)
	for _, claim := range config.roleClaims {
		roleValues = append(roleValues, claimStrings(claims, claim)...)
	}
	normalizedRoles := normalizeRoleValues(roleValues, config.roleMap)
	if len(normalizedRoles) == 0 {
		normalizedRoles = []string{config.defaultRole}
	}
	normalizedRoles = filterAllowedRoles(normalizedRoles, config.allowedRoles)
	if len(normalizedRoles) == 0 && config.requireAllowedRole {
		return "", "", nil, errors.New("oidc no allowed role after whitelist filter")
	}
	if len(normalizedRoles) == 0 {
		normalizedRoles = []string{config.defaultRole}
	}
	return userID, username, normalizedRoles, nil
}

func claimString(claims map[string]any, key string) string {
	value, ok := claims[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func claimStrings(claims map[string]any, key string) []string {
	value, ok := claims[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil
		}
		return []string{trimmed}
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			out = append(out, trimmed)
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				continue
			}
			trimmed := strings.TrimSpace(text)
			if trimmed == "" {
				continue
			}
			out = append(out, trimmed)
		}
		return out
	default:
		return nil
	}
}

func normalizeRoleValues(raw []string, roleMap map[string]string) []string {
	if roleMap == nil {
		roleMap = map[string]string{}
	}
	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, role := range raw {
		trimmed := strings.TrimSpace(role)
		if trimmed == "" {
			continue
		}
		mapped, ok := roleMap[trimmed]
		if ok && strings.TrimSpace(mapped) != "" {
			trimmed = strings.TrimSpace(mapped)
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func filterAllowedRoles(roles []string, allowed map[string]struct{}) []string {
	if len(allowed) == 0 {
		return roles
	}
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		if _, ok := allowed[role]; ok {
			out = append(out, role)
		}
	}
	return out
}
