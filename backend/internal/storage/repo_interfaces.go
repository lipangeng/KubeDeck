package storage

// UserMenuRepo manages user-customized menu state.
type UserMenuRepo interface{}

// UserPreferenceRepo manages user-level preferences.
type UserPreferenceRepo interface{}

// PluginConfigRepo manages plugin configuration state.
type PluginConfigRepo interface{}

// Store exposes selected storage backend metadata and repositories.
type Store interface {
	Driver() string
	UserMenus() UserMenuRepo
	UserPreferences() UserPreferenceRepo
	PluginConfigs() PluginConfigRepo
}

type stubUserMenuRepo struct{}
type stubUserPreferenceRepo struct{}
type stubPluginConfigRepo struct{}

var (
	defaultUserMenuRepo       UserMenuRepo       = stubUserMenuRepo{}
	defaultUserPreferenceRepo UserPreferenceRepo = stubUserPreferenceRepo{}
	defaultPluginConfigRepo   PluginConfigRepo   = stubPluginConfigRepo{}
)
