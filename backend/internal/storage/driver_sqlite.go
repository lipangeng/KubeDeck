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
