package storage

type sqliteStore struct {
	dsn      string
	userMenu UserMenuRepo
}

func newSQLiteStore(dsn string) Store {
	return &sqliteStore{
		dsn:      dsn,
		userMenu: newInMemoryUserMenuRepo(),
	}
}

func (s *sqliteStore) Driver() string {
	return "sqlite"
}

func (s *sqliteStore) UserMenus() UserMenuRepo {
	return s.userMenu
}

func (s *sqliteStore) UserPreferences() UserPreferenceRepo {
	return defaultUserPreferenceRepo
}

func (s *sqliteStore) PluginConfigs() PluginConfigRepo {
	return defaultPluginConfigRepo
}
