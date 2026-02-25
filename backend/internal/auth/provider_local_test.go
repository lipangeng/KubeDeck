package auth

import "testing"

func TestLocalProviderRejectsInvalidCredentials(t *testing.T) {
	t.Setenv("KUBEDECK_LOCAL_AUTH_ENABLED", "1")
	t.Setenv("KUBEDECK_LOCAL_AUTH_USERNAME", "ops-admin")
	t.Setenv("KUBEDECK_LOCAL_AUTH_PASSWORD", "S3cur3Pass!")

	provider := NewLocalProvider()
	if _, err := provider.Authenticate("ops-admin", "wrong"); err == nil {
		t.Fatalf("expected invalid credentials to be rejected")
	}
}

func TestLocalProviderUsesConfiguredFixedCredential(t *testing.T) {
	t.Setenv("KUBEDECK_LOCAL_AUTH_ENABLED", "1")
	t.Setenv("KUBEDECK_LOCAL_AUTH_USERNAME", "ops-admin")
	t.Setenv("KUBEDECK_LOCAL_AUTH_PASSWORD", "S3cur3Pass!")

	provider := NewLocalProvider()
	user, err := provider.Authenticate("ops-admin", "S3cur3Pass!")
	if err != nil {
		t.Fatalf("expected configured credential login success, got err=%v", err)
	}
	if user.Username != "ops-admin" {
		t.Fatalf("expected username ops-admin, got %q", user.Username)
	}
}

func TestLocalProviderAllowsMultipleUsernamesWhenUsernameNotConfigured(t *testing.T) {
	t.Setenv("KUBEDECK_LOCAL_AUTH_ENABLED", "1")
	t.Setenv("KUBEDECK_LOCAL_AUTH_USERNAME", "")
	t.Setenv("KUBEDECK_LOCAL_AUTH_PASSWORD", "SharedPass#1")

	provider := NewLocalProvider()
	if _, err := provider.Authenticate("viewer", "SharedPass#1"); err != nil {
		t.Fatalf("expected viewer login success with configured fixed password, got err=%v", err)
	}
	if _, err := provider.Authenticate("admin", "SharedPass#1"); err != nil {
		t.Fatalf("expected admin login success with configured fixed password, got err=%v", err)
	}
}

func TestLocalProviderDisabledByDefaultInProduction(t *testing.T) {
	t.Setenv("KUBEDECK_ENV", "production")
	t.Setenv("KUBEDECK_LOCAL_AUTH_ENABLED", "")
	t.Setenv("KUBEDECK_LOCAL_AUTH_USERNAME", "")
	t.Setenv("KUBEDECK_LOCAL_AUTH_PASSWORD", "")

	provider := NewLocalProvider()
	if _, err := provider.Authenticate("admin", "pw"); err == nil {
		t.Fatalf("expected local provider disabled in production by default")
	}
}

func TestLocalProviderCanBeExplicitlyEnabledInProductionWithEnvCredential(t *testing.T) {
	t.Setenv("KUBEDECK_ENV", "production")
	t.Setenv("KUBEDECK_LOCAL_AUTH_ENABLED", "true")
	t.Setenv("KUBEDECK_LOCAL_AUTH_USERNAME", "prod-admin")
	t.Setenv("KUBEDECK_LOCAL_AUTH_PASSWORD", "ProdStrongPass#1")

	provider := NewLocalProvider()
	if _, err := provider.Authenticate("prod-admin", "ProdStrongPass#1"); err != nil {
		t.Fatalf("expected explicit production local provider login success, got err=%v", err)
	}
}

func TestLocalProviderGeneratesStableUserIDFromUsername(t *testing.T) {
	t.Setenv("KUBEDECK_LOCAL_AUTH_ENABLED", "1")
	t.Setenv("KUBEDECK_LOCAL_AUTH_USERNAME", "")
	t.Setenv("KUBEDECK_LOCAL_AUTH_PASSWORD", "SharedPass#1")

	provider := NewLocalProvider()
	first, err := provider.Authenticate("viewer", "SharedPass#1")
	if err != nil {
		t.Fatalf("expected first authenticate success, got err=%v", err)
	}
	second, err := provider.Authenticate("VIEWER", "SharedPass#1")
	if err != nil {
		t.Fatalf("expected second authenticate success, got err=%v", err)
	}
	third, err := provider.Authenticate("admin", "SharedPass#1")
	if err != nil {
		t.Fatalf("expected third authenticate success, got err=%v", err)
	}
	if first.ID == "" {
		t.Fatalf("expected non-empty user id")
	}
	if first.ID != second.ID {
		t.Fatalf("expected stable id for same normalized username, got %q != %q", first.ID, second.ID)
	}
	if first.ID == third.ID {
		t.Fatalf("expected different ids for different usernames, got %q", first.ID)
	}
}
