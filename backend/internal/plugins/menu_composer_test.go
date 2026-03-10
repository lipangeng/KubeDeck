package plugins

import (
	"testing"

	"kubedeck/backend/pkg/sdk"
)

func TestComposeMenusOrdersByOrderThenID(t *testing.T) {
	descriptors := []sdk.CapabilityDescriptor{
		{
			ID: "b",
			Menus: []sdk.MenuDescriptor{
				{ID: "workloads", Order: 20},
			},
		},
		{
			ID: "a",
			Menus: []sdk.MenuDescriptor{
				{ID: "homepage", Order: 10},
				{ID: "secondary", Order: 20},
			},
		},
	}

	menus := ComposeMenus(descriptors)
	if len(menus) != 3 {
		t.Fatalf("expected 3 menus, got %d", len(menus))
	}
	if menus[0].ID != "homepage" || menus[1].ID != "secondary" || menus[2].ID != "workloads" {
		t.Fatalf("unexpected menu order: %#v", menus)
	}
}
