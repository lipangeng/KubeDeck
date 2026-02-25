package auth

import (
	"os"
	"strings"
)

func IsProductionRuntime() bool {
	return resolveRuntimeMode() == "production"
}

func resolveRuntimeMode() string {
	for _, key := range []string{"KUBEDECK_ENV", "KUBEDECK_RUNTIME_MODE", "APP_ENV", "GO_ENV"} {
		if mode := normalizeRuntimeMode(os.Getenv(key)); mode != "" {
			return mode
		}
	}

	if strings.HasSuffix(strings.ToLower(strings.TrimSpace(os.Args[0])), ".test") {
		return "test"
	}
	return "development"
}

func normalizeRuntimeMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "prod", "production":
		return "production"
	case "test", "testing":
		return "test"
	case "dev", "development", "local":
		return "development"
	default:
		return ""
	}
}
