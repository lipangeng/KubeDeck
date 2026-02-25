package auth

import (
	"errors"
	"os"
	"strings"
)

var ErrLocalProviderDisabled = errors.New("local provider disabled")
var ErrLocalInvalidCredentials = errors.New("local provider invalid credentials")

// LocalProvider is the MVP local-auth stub provider.
type LocalProvider struct {
	enabled  bool
	username string
	password string
}

func NewLocalProvider() *LocalProvider {
	isProd := isProductionEnv()
	enabled := !isProd
	if rawEnabled := strings.TrimSpace(os.Getenv("KUBEDECK_LOCAL_AUTH_ENABLED")); rawEnabled != "" {
		enabled = strings.EqualFold(rawEnabled, "1") || strings.EqualFold(rawEnabled, "true") || strings.EqualFold(rawEnabled, "yes")
	}

	username := strings.TrimSpace(os.Getenv("KUBEDECK_LOCAL_AUTH_USERNAME"))
	password := os.Getenv("KUBEDECK_LOCAL_AUTH_PASSWORD")
	if !isProd {
		if password == "" {
			password = "pw"
		}
	}

	if enabled && password == "" {
		enabled = false
	}

	return &LocalProvider{
		enabled:  enabled,
		username: username,
		password: password,
	}
}

func (p *LocalProvider) Name() string {
	return "local"
}

func (p *LocalProvider) Authenticate(username, password string) (User, error) {
	if !p.enabled {
		return User{}, ErrLocalProviderDisabled
	}
	if strings.TrimSpace(username) == "" {
		return User{}, ErrLocalInvalidCredentials
	}
	if p.username != "" && username != p.username {
		return User{}, ErrLocalInvalidCredentials
	}
	if password != p.password {
		return User{}, ErrLocalInvalidCredentials
	}

	roles := []string{"viewer"}
	if username == "admin" || username == "owner" {
		roles = []string{"admin"}
	}

	return User{
		ID:                "local-test-user",
		Username:          username,
		Roles:             roles,
		AllowedClusters:   []string{"*"},
		AllowedNamespaces: []string{"*"},
	}, nil
}

func isProductionEnv() bool {
	for _, key := range []string{"KUBEDECK_ENV", "APP_ENV", "GO_ENV"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		return strings.EqualFold(value, "production") || strings.EqualFold(value, "prod")
	}
	return false
}
