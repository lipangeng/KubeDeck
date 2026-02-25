package auth

import (
	"errors"
	"testing"
)

func TestNewOAuthProviderFromEnv_OIDCInitFailureProductionDisablesFallback(t *testing.T) {
	t.Setenv("KUBEDECK_OAUTH_MODE", "oidc")
	t.Setenv("KUBEDECK_ENV", "production")
	t.Setenv("KUBEDECK_OIDC_ISSUER", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_ID", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_SECRET", "")
	t.Setenv("KUBEDECK_OIDC_REDIRECT_URL", "")

	provider := NewOAuthProviderFromEnv()
	if provider == nil {
		t.Fatal("expected oauth provider")
	}
	if provider.Name() != "oidc" {
		t.Fatalf("expected provider name oidc, got %q", provider.Name())
	}
	initErr := OAuthProviderInitError(provider)
	if initErr == nil {
		t.Fatal("expected oauth provider init error in production")
	}
	if initErr.Error() != "oidc issuer is required" {
		t.Fatalf("unexpected init error: %v", initErr)
	}

	_, err := provider.ExchangeCode("any")
	if !errors.Is(err, ErrOAuthProviderUnavailable) {
		t.Fatalf("expected ErrOAuthProviderUnavailable, got %v", err)
	}
}

func TestNewOAuthProviderFromEnv_OIDCInitFailureTestKeepsStubFallback(t *testing.T) {
	t.Setenv("KUBEDECK_OAUTH_MODE", "oidc")
	t.Setenv("KUBEDECK_ENV", "test")
	t.Setenv("KUBEDECK_OIDC_ISSUER", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_ID", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_SECRET", "")
	t.Setenv("KUBEDECK_OIDC_REDIRECT_URL", "")

	provider := NewOAuthProviderFromEnv()
	if provider == nil {
		t.Fatal("expected oauth provider")
	}
	if provider.Name() != "oauth" {
		t.Fatalf("expected fallback stub provider name oauth, got %q", provider.Name())
	}
	if OAuthProviderInitError(provider) != nil {
		t.Fatal("expected no init error for fallback stub provider")
	}
}
