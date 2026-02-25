package storage

// UserMenuRepo manages user-customized menu state.
type UserMenuRepo interface{}

// UserPreferenceRepo manages user-level preferences.
type UserPreferenceRepo interface{}

// PluginConfigRepo manages plugin configuration state.
type PluginConfigRepo interface{}

// UserRepo manages IAM users.
type UserRepo interface{}

// TenantRepo manages IAM tenants.
type TenantRepo interface{}

// TenantMembershipRepo manages user-tenant memberships.
type TenantMembershipRepo interface{}

// GroupRepo manages IAM groups.
type GroupRepo interface{}

// PermissionRepo manages IAM permissions.
type PermissionRepo interface{}

// SessionRepo manages IAM sessions.
type SessionRepo interface{}

// InviteRepo manages IAM invites.
type InviteRepo interface{}

// AuditEventRepo manages IAM audit events.
type AuditEventRepo interface{}

// Store exposes selected storage backend metadata and repositories.
type Store interface {
	Driver() string
	UserMenus() UserMenuRepo
	UserPreferences() UserPreferenceRepo
	PluginConfigs() PluginConfigRepo
	Users() UserRepo
	Tenants() TenantRepo
	TenantMemberships() TenantMembershipRepo
	Groups() GroupRepo
	Permissions() PermissionRepo
	Sessions() SessionRepo
	Invites() InviteRepo
	AuditEvents() AuditEventRepo
}

type stubUserMenuRepo struct{}
type stubUserPreferenceRepo struct{}
type stubPluginConfigRepo struct{}
type stubUserRepo struct{}
type stubTenantRepo struct{}
type stubTenantMembershipRepo struct{}
type stubGroupRepo struct{}
type stubPermissionRepo struct{}
type stubSessionRepo struct{}
type stubInviteRepo struct{}
type stubAuditEventRepo struct{}

var (
	defaultUserMenuRepo         UserMenuRepo         = stubUserMenuRepo{}
	defaultUserPreferenceRepo   UserPreferenceRepo   = stubUserPreferenceRepo{}
	defaultPluginConfigRepo     PluginConfigRepo     = stubPluginConfigRepo{}
	defaultUserRepo             UserRepo             = stubUserRepo{}
	defaultTenantRepo           TenantRepo           = stubTenantRepo{}
	defaultTenantMembershipRepo TenantMembershipRepo = stubTenantMembershipRepo{}
	defaultGroupRepo            GroupRepo            = stubGroupRepo{}
	defaultPermissionRepo       PermissionRepo       = stubPermissionRepo{}
	defaultSessionRepo          SessionRepo          = stubSessionRepo{}
	defaultInviteRepo           InviteRepo           = stubInviteRepo{}
	defaultAuditEventRepo       AuditEventRepo       = stubAuditEventRepo{}
)
