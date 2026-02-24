package auth

// EvaluateAccess applies minimal cluster/namespace allowlist checks.
func EvaluateAccess(user User, req AccessRequest) AccessDecision {
	if req.Cluster == "" {
		return AccessDecision{Allowed: false, Reason: "cluster_required"}
	}

	if !matchesAllowlist(req.Cluster, user.AllowedClusters) {
		return AccessDecision{Allowed: false, Reason: "cluster_denied"}
	}

	if req.Namespace != "" && !matchesAllowlist(req.Namespace, user.AllowedNamespaces) {
		return AccessDecision{Allowed: false, Reason: "namespace_denied"}
	}

	return AccessDecision{Allowed: true}
}

func matchesAllowlist(value string, allowlist []string) bool {
	for _, item := range allowlist {
		if item == "*" || item == value {
			return true
		}
	}
	return false
}
