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

	return User{
		ID:                "local-test-user",
		Username:          username,
		Roles:             []string{"viewer"},
		AllowedClusters:   []string{"*"},
		AllowedNamespaces: []string{"*"},
	}, nil
}
