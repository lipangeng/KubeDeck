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
