package storage

import (
	"sync"

	"kubedeck/backend/internal/plugins"
)

// UserMenuRepo manages user-customized menu state.
type UserMenuRepo interface {
	GetGlobalOverrides(userID string) []plugins.MenuOverride
	GetClusterOverrides(userID string, cluster string) []plugins.MenuOverride
	SaveGlobalOverrides(userID string, overrides []plugins.MenuOverride) error
	SaveClusterOverrides(userID string, cluster string, overrides []plugins.MenuOverride) error
}

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

type inMemoryUserMenuRepo struct {
	mu       sync.RWMutex
	global   map[string][]plugins.MenuOverride
	clusters map[string]map[string][]plugins.MenuOverride
}
type stubUserPreferenceRepo struct{}
type stubPluginConfigRepo struct{}

func newInMemoryUserMenuRepo() UserMenuRepo {
	return &inMemoryUserMenuRepo{
		global:   make(map[string][]plugins.MenuOverride),
		clusters: make(map[string]map[string][]plugins.MenuOverride),
	}
}

func (r *inMemoryUserMenuRepo) GetGlobalOverrides(userID string) []plugins.MenuOverride {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneMenuOverrides(r.global[userID])
}

func (r *inMemoryUserMenuRepo) GetClusterOverrides(userID string, cluster string) []plugins.MenuOverride {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneMenuOverrides(r.clusters[userID][cluster])
}

func (r *inMemoryUserMenuRepo) SaveGlobalOverrides(userID string, overrides []plugins.MenuOverride) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.global[userID] = cloneMenuOverrides(overrides)
	return nil
}

func (r *inMemoryUserMenuRepo) SaveClusterOverrides(
	userID string,
	cluster string,
	overrides []plugins.MenuOverride,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clusterOverrides, ok := r.clusters[userID]
	if !ok {
		clusterOverrides = make(map[string][]plugins.MenuOverride)
		r.clusters[userID] = clusterOverrides
	}
	clusterOverrides[cluster] = cloneMenuOverrides(overrides)
	return nil
}

func cloneMenuOverrides(overrides []plugins.MenuOverride) []plugins.MenuOverride {
	if len(overrides) == 0 {
		return nil
	}
	cloned := make([]plugins.MenuOverride, 0, len(overrides))
	for _, override := range overrides {
		copyOverride := plugins.MenuOverride{
			Scope:           override.Scope,
			HiddenEntryKeys: append([]string(nil), override.HiddenEntryKeys...),
			PinEntryKeys:    append([]string(nil), override.PinEntryKeys...),
		}
		if len(override.MoveEntryKeys) > 0 {
			copyOverride.MoveEntryKeys = make(map[string]string, len(override.MoveEntryKeys))
			for key, value := range override.MoveEntryKeys {
				copyOverride.MoveEntryKeys[key] = value
			}
		}
		cloned = append(cloned, copyOverride)
	}
	return cloned
}

var (
	defaultUserPreferenceRepo UserPreferenceRepo = stubUserPreferenceRepo{}
	defaultPluginConfigRepo   PluginConfigRepo   = stubPluginConfigRepo{}
)
