package auth

import (
	"testing"
	"time"
)

func TestEvaluateAccess(t *testing.T) {
	tests := []struct {
		name          string
		clusters      []string
		namespaces    []string
		cluster       string
		namespace     string
		wantAllowed   bool
		wantDenyCause string
	}{
		{
			name:          "allows exact cluster and namespace",
			clusters:      []string{"dev"},
			namespaces:    []string{"team-a"},
			cluster:       "dev",
			namespace:     "team-a",
			wantAllowed:   true,
			wantDenyCause: "",
		},
		{
			name:          "denies when cluster is not in allowlist",
			clusters:      []string{"prod"},
			namespaces:    []string{"team-a"},
			cluster:       "dev",
			namespace:     "team-a",
			wantAllowed:   false,
			wantDenyCause: "cluster_denied",
		},
		{
			name:          "denies when namespace is not in allowlist",
			clusters:      []string{"dev"},
			namespaces:    []string{"team-a"},
			cluster:       "dev",
			namespace:     "team-b",
			wantAllowed:   false,
			wantDenyCause: "namespace_denied",
		},
		{
			name:          "allows wildcard cluster and namespace",
			clusters:      []string{"*"},
			namespaces:    []string{"*"},
			cluster:       "staging",
			namespace:     "ops",
			wantAllowed:   true,
			wantDenyCause: "",
		},
		{
			name:          "allows cluster-scoped request with namespace allowlist present",
			clusters:      []string{"dev"},
			namespaces:    []string{"team-a"},
			cluster:       "dev",
			namespace:     "",
			wantAllowed:   true,
			wantDenyCause: "",
		},
		{
			name:          "denies when cluster is required",
			clusters:      []string{"*"},
			namespaces:    []string{"*"},
			cluster:       "",
			namespace:     "team-a",
			wantAllowed:   false,
			wantDenyCause: "cluster_required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				ID:                "u-1",
				Username:          "alice",
				AllowedClusters:   tt.clusters,
				AllowedNamespaces: tt.namespaces,
			}
			req := AccessRequest{Cluster: tt.cluster, Namespace: tt.namespace}

			decision := EvaluateAccess(user, req)
			if decision.Allowed != tt.wantAllowed {
				t.Fatalf("allowed mismatch: got %v, want %v", decision.Allowed, tt.wantAllowed)
			}
			if decision.Reason != tt.wantDenyCause {
				t.Fatalf("deny cause mismatch: got %q, want %q", decision.Reason, tt.wantDenyCause)
			}
		})
	}
}

func TestTenantMembershipIsActiveAt(t *testing.T) {
	now := time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC)
	past := now.Add(-2 * time.Hour)
	future := now.Add(2 * time.Hour)

	tests := []struct {
		name       string
		membership TenantMembership
		want       bool
	}{
		{
			name: "active with open-ended expiry",
			membership: TenantMembership{
				TenantID:      "t-1",
				UserID:        "u-1",
				EffectiveFrom: past,
			},
			want: true,
		},
		{
			name: "inactive before effective time",
			membership: TenantMembership{
				TenantID:      "t-1",
				UserID:        "u-1",
				EffectiveFrom: future,
			},
			want: false,
		},
		{
			name: "inactive after expiry",
			membership: TenantMembership{
				TenantID:       "t-1",
				UserID:         "u-1",
				EffectiveFrom:  past.Add(-2 * time.Hour),
				EffectiveUntil: &past,
			},
			want: false,
		},
		{
			name: "active inside closed interval",
			membership: TenantMembership{
				TenantID:       "t-1",
				UserID:         "u-1",
				EffectiveFrom:  past,
				EffectiveUntil: &future,
			},
			want: true,
		},
		{
			name: "boundary at effective until is inactive",
			membership: TenantMembership{
				TenantID:       "t-1",
				UserID:         "u-1",
				EffectiveFrom:  past,
				EffectiveUntil: &now,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.membership.IsActiveAt(now)
			if got != tt.want {
				t.Fatalf("active mismatch: got %v, want %v", got, tt.want)
			}
		})
	}
}
