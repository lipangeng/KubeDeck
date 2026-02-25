package storage

type sqliteStore struct {
	dsn string
}

func newSQLiteStore(dsn string) Store {
	return &sqliteStore{dsn: dsn}
}

func (s *sqliteStore) Driver() string {
	return "sqlite"
}

func (s *sqliteStore) UserMenus() UserMenuRepo {
	return defaultUserMenuRepo
}

func (s *sqliteStore) UserPreferences() UserPreferenceRepo {
	return defaultUserPreferenceRepo
}

func (s *sqliteStore) PluginConfigs() PluginConfigRepo {
	return defaultPluginConfigRepo
}

func (s *sqliteStore) Users() UserRepo {
	return defaultUserRepo
}

func (s *sqliteStore) Tenants() TenantRepo {
	return defaultTenantRepo
}

func (s *sqliteStore) TenantMemberships() TenantMembershipRepo {
	return defaultTenantMembershipRepo
}

func (s *sqliteStore) Groups() GroupRepo {
	return defaultGroupRepo
}

func (s *sqliteStore) Permissions() PermissionRepo {
	return defaultPermissionRepo
}

func (s *sqliteStore) Sessions() SessionRepo {
	return defaultSessionRepo
}

func (s *sqliteStore) Invites() InviteRepo {
	return defaultInviteRepo
}

func (s *sqliteStore) AuditEvents() AuditEventRepo {
	return defaultAuditEventRepo
}
