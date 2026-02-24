package plugins

import (
	"fmt"
	"reflect"

	"kubedeck/backend/pkg/sdk"
)

// Manager stores plugins by unique ID.
type Manager struct {
	plugins map[string]sdk.Plugin
}

func NewManager() *Manager {
	return &Manager{plugins: map[string]sdk.Plugin{}}
}

func (m *Manager) Register(plugin sdk.Plugin) error {
	if plugin == nil || isTypedNil(plugin) {
		return fmt.Errorf("plugin is nil")
	}

	id := plugin.ID()
	if _, exists := m.plugins[id]; exists {
		return fmt.Errorf("plugin %q already registered", id)
	}

	m.plugins[id] = plugin
	return nil
}

func isTypedNil(plugin sdk.Plugin) bool {
	value := reflect.ValueOf(plugin)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
