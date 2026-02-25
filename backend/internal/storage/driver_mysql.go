package storage

type mysqlStore struct {
	dsn string
}

func newMySQLStore(dsn string) Store {
	return &mysqlStore{dsn: dsn}
}

func (s *mysqlStore) Driver() string {
	return "mysql"
}

func (s *mysqlStore) UserMenus() UserMenuRepo {
	return defaultUserMenuRepo
}

func (s *mysqlStore) UserPreferences() UserPreferenceRepo {
	return defaultUserPreferenceRepo
}

func (s *mysqlStore) PluginConfigs() PluginConfigRepo {
	return defaultPluginConfigRepo
}

func (s *mysqlStore) Users() UserRepo {
	return defaultUserRepo
}

func (s *mysqlStore) Tenants() TenantRepo {
	return defaultTenantRepo
}

func (s *mysqlStore) TenantMemberships() TenantMembershipRepo {
	return defaultTenantMembershipRepo
}

func (s *mysqlStore) Groups() GroupRepo {
	return defaultGroupRepo
}

func (s *mysqlStore) Permissions() PermissionRepo {
	return defaultPermissionRepo
}

func (s *mysqlStore) Sessions() SessionRepo {
	return defaultSessionRepo
}

func (s *mysqlStore) Invites() InviteRepo {
	return defaultInviteRepo
}

func (s *mysqlStore) AuditEvents() AuditEventRepo {
	return defaultAuditEventRepo
}
