package storage

type mysqlStore struct {
	dsn      string
	userMenu UserMenuRepo
}

func newMySQLStore(dsn string) Store {
	return &mysqlStore{
		dsn:      dsn,
		userMenu: newInMemoryUserMenuRepo(),
	}
}

func (s *mysqlStore) Driver() string {
	return "mysql"
}

func (s *mysqlStore) UserMenus() UserMenuRepo {
	return s.userMenu
}

func (s *mysqlStore) UserPreferences() UserPreferenceRepo {
	return defaultUserPreferenceRepo
}

func (s *mysqlStore) PluginConfigs() PluginConfigRepo {
	return defaultPluginConfigRepo
}
