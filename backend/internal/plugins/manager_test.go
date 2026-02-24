package plugins

import (
	"testing"

	"kubedeck/backend/pkg/sdk"
)

type testPlugin struct {
	id string
}

type nilablePlugin struct {
	id string
}

func (p testPlugin) ID() string {
	return p.id
}

func (p *nilablePlugin) ID() string {
	return p.id
}

func TestManagerRegistersPlugin(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("expected manager instance")
	}

	first := testPlugin{id: "metrics"}
	if err := manager.Register(first); err != nil {
		t.Fatalf("expected first register to succeed: %v", err)
	}

	duplicate := testPlugin{id: "metrics"}
	err := manager.Register(duplicate)
	if err == nil {
		t.Fatal("expected duplicate plugin ID error")
	}

	err = manager.Register(nil)
	if err == nil {
		t.Fatal("expected nil plugin error")
	}

	t.Run("typed nil plugin returns error without panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("expected no panic for typed nil plugin, got: %v", r)
			}
		}()

		var typedNil *nilablePlugin
		err := manager.Register(typedNil)
		if err == nil {
			t.Fatal("expected typed nil plugin error")
		}
	})
}

var _ sdk.Plugin = testPlugin{}
var _ sdk.Plugin = (*nilablePlugin)(nil)
