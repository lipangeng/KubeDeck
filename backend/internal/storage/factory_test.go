package storage

import "testing"

func TestNewStore(t *testing.T) {
	tests := []struct {
		name      string
		driver     string
		wantDriver string
		wantErr   bool
	}{
		{name: "sqlite", driver: "sqlite", wantDriver: "sqlite"},
		{name: "empty defaults to sqlite", driver: "", wantDriver: "sqlite"},
		{name: "sqlite trimmed and case insensitive", driver: "  SqlIte  ", wantDriver: "sqlite"},
		{name: "mysql", driver: "mysql", wantDriver: "mysql"},
		{name: "mysql case insensitive", driver: "MySQL", wantDriver: "mysql"},
		{name: "postgres", driver: "postgres", wantDriver: "postgres"},
		{name: "unknown", driver: "oracle", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewStore(tt.driver, "dsn")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if store == nil {
				t.Fatal("expected store")
			}
			if got := store.Driver(); got != tt.wantDriver {
				t.Fatalf("driver mismatch: got %q want %q", got, tt.wantDriver)
			}
			if store.UserMenus() == nil {
				t.Fatal("expected non-nil user menu repo")
			}
			if store.UserPreferences() == nil {
				t.Fatal("expected non-nil user preference repo")
			}
			if store.PluginConfigs() == nil {
				t.Fatal("expected non-nil plugin config repo")
			}
		})
	}
}
