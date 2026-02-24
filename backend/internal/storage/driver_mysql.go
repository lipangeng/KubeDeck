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
