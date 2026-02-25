package api

import (
	"strings"
)

var rolePermissions = map[string][]string{
	"owner": {
		"*",
	},
	"admin": {
		"*",
	},
	"viewer": {
		"iam:read",
		"tenant:read",
		"menu:read",
		"resource:read",
		"cluster:switch",
	},
}

func sessionHasAnyPermission(session authSession, required []string) bool {
	for _, permission := range required {
		if hasPermission(session.User.Roles, permission) {
			return true
		}
	}
	return false
}

func hasPermission(roles []string, permission string) bool {
	target := strings.ToLower(strings.TrimSpace(permission))
	if target == "" {
		return true
	}
	for _, role := range roles {
		roleKey := strings.ToLower(strings.TrimSpace(role))
		if roleKey == "" {
			continue
		}
		if roleKey == target || roleKey == "*" {
			return true
		}
		for _, granted := range rolePermissions[roleKey] {
			if granted == "*" || strings.EqualFold(granted, target) {
				return true
			}
		}
	}
	return false
}
