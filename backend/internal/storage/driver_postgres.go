package storage

type postgresStore struct {
	dsn string
}

func newPostgresStore(dsn string) Store {
	return &postgresStore{dsn: dsn}
}

func (s *postgresStore) Driver() string {
	return "postgres"
}

func (s *postgresStore) UserMenus() UserMenuRepo {
	return defaultUserMenuRepo
}

func (s *postgresStore) UserPreferences() UserPreferenceRepo {
	return defaultUserPreferenceRepo
}

func (s *postgresStore) PluginConfigs() PluginConfigRepo {
	return defaultPluginConfigRepo
}

func (s *postgresStore) Users() UserRepo {
	return defaultUserRepo
}

func (s *postgresStore) Tenants() TenantRepo {
	return defaultTenantRepo
}

func (s *postgresStore) TenantMemberships() TenantMembershipRepo {
	return defaultTenantMembershipRepo
}

func (s *postgresStore) Groups() GroupRepo {
	return defaultGroupRepo
}

func (s *postgresStore) Permissions() PermissionRepo {
	return defaultPermissionRepo
}

func (s *postgresStore) Sessions() SessionRepo {
	return defaultSessionRepo
}

func (s *postgresStore) Invites() InviteRepo {
	return defaultInviteRepo
}

func (s *postgresStore) AuditEvents() AuditEventRepo {
	return defaultAuditEventRepo
}
