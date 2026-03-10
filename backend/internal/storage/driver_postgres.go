package storage

type postgresStore struct {
	dsn      string
	userMenu UserMenuRepo
}

func newPostgresStore(dsn string) Store {
	return &postgresStore{
		dsn:      dsn,
		userMenu: newInMemoryUserMenuRepo(),
	}
}

func (s *postgresStore) Driver() string {
	return "postgres"
}

func (s *postgresStore) UserMenus() UserMenuRepo {
	return s.userMenu
}

func (s *postgresStore) UserPreferences() UserPreferenceRepo {
	return defaultUserPreferenceRepo
}

func (s *postgresStore) PluginConfigs() PluginConfigRepo {
	return defaultPluginConfigRepo
}
