package auth

// LocalProvider is the MVP local-auth stub provider.
type LocalProvider struct{}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{}
}

func (p *LocalProvider) Name() string {
	return "local"
}

func (p *LocalProvider) Authenticate(username, _ string) (User, error) {
	if username == "" {
		username = "local-test"
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
