package auth

import "time"

// TenantMembership describes user's membership in a tenant with validity window.
type TenantMembership struct {
	TenantID       string
	UserID         string
	EffectiveFrom  time.Time
	EffectiveUntil *time.Time
}

// IsActiveAt returns true when membership is active at the provided timestamp.
func (m TenantMembership) IsActiveAt(now time.Time) bool {
	if !m.EffectiveFrom.IsZero() && now.Before(m.EffectiveFrom) {
		return false
	}
	if m.EffectiveUntil != nil && !now.Before(*m.EffectiveUntil) {
		return false
	}
	return true
}

