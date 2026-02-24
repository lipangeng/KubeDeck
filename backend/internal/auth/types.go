package auth

// User represents the authenticated identity plus coarse RBAC allowlists.
type User struct {
	ID                string
	Username          string
	Roles             []string
	AllowedClusters   []string
	AllowedNamespaces []string
}

// AccessRequest is the target scope for an authorization decision.
type AccessRequest struct {
	Cluster   string
	Namespace string
}

// AccessDecision is the result of RBAC/allowlist evaluation.
type AccessDecision struct {
	Allowed bool
	Reason  string
}

// Provider defines a local/auth backend contract.
type Provider interface {
	Name() string
	Authenticate(username, password string) (User, error)
}

// OAuthProvider defines an oauth-capable provider contract.
type OAuthProvider interface {
	Name() string
	BeginAuthURL(state string) string
	ExchangeCode(code string) (User, error)
}
