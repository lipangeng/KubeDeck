package auth

import (
	"os"
	"strings"
)

type OAuthConfigDiagnostics struct {
	Mode    string
	Ready   bool
	Missing []string
	OIDC    OIDCConfigPresence
}

type OIDCConfigPresence struct {
	IssuerExists       bool
	ClientIDExists     bool
	ClientSecretExists bool
	RedirectURLExists  bool
}

func OAuthConfigDiagnosticsFromEnv() OAuthConfigDiagnostics {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("KUBEDECK_OAUTH_MODE")))
	if mode == "" {
		mode = "stub"
	}

	oidc := OIDCConfigPresence{
		IssuerExists:       strings.TrimSpace(os.Getenv("KUBEDECK_OIDC_ISSUER")) != "",
		ClientIDExists:     strings.TrimSpace(os.Getenv("KUBEDECK_OIDC_CLIENT_ID")) != "",
		ClientSecretExists: strings.TrimSpace(os.Getenv("KUBEDECK_OIDC_CLIENT_SECRET")) != "",
		RedirectURLExists:  strings.TrimSpace(os.Getenv("KUBEDECK_OIDC_REDIRECT_URL")) != "",
	}

	missing := []string{}
	ready := true
	if mode == "oidc" {
		if !oidc.IssuerExists {
			missing = append(missing, "KUBEDECK_OIDC_ISSUER")
		}
		if !oidc.ClientIDExists {
			missing = append(missing, "KUBEDECK_OIDC_CLIENT_ID")
		}
		if !oidc.ClientSecretExists {
			missing = append(missing, "KUBEDECK_OIDC_CLIENT_SECRET")
		}
		if !oidc.RedirectURLExists {
			missing = append(missing, "KUBEDECK_OIDC_REDIRECT_URL")
		}
		ready = len(missing) == 0
	}

	return OAuthConfigDiagnostics{
		Mode:    mode,
		Ready:   ready,
		Missing: missing,
		OIDC:    oidc,
	}
}
