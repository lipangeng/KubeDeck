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
				{
					ID:           "workloads",
					GroupKey:     "core",
					Availability: sdk.MenuAvailabilityEnabled,
					Order:        20,
				},
			},
		},
		{
			ID: "a",
			Menus: []sdk.MenuDescriptor{
				{
					ID:           "homepage",
					GroupKey:     "core",
					Availability: sdk.MenuAvailabilityEnabled,
					Order:        10,
				},
				{
					ID:           "secondary",
					GroupKey:     "resources",
					Availability: sdk.MenuAvailabilityDisabledUnavailable,
					IsFallback:   true,
					Order:        20,
				},
			},
		},
	}

	menus := ComposeMenus(descriptors)
	if len(menus) != 4 {
		t.Fatalf("expected 4 menus including fallback, got %d", len(menus))
	}
	if menus[0].ID != "homepage" || menus[1].ID != "workloads" || menus[2].ID != "secondary" || menus[3].ID != "menu.crds" {
		t.Fatalf("unexpected menu order: %#v", menus)
	}
	if menus[0].GroupKey != "core" {
		t.Fatalf("expected first menu group to be preserved, got %q", menus[0].GroupKey)
	}
	if menus[2].Availability != sdk.MenuAvailabilityDisabledUnavailable {
		t.Fatalf("expected disabled-unavailable availability, got %q", menus[2].Availability)
	}
	if !menus[2].IsFallback {
		t.Fatalf("expected fallback flag to be preserved")
	}
	if !menus[3].IsFallback || menus[3].EntryKey != "crds" {
		t.Fatalf("expected trailing fallback crds menu, got %#v", menus[3])
	}
}

func TestComposeMenusAppendsCrdsFallbackAndRespectsBlueprintGroupOrder(t *testing.T) {
	descriptors := []sdk.CapabilityDescriptor{
		{
			ID: "core.workloads",
			Menus: []sdk.MenuDescriptor{
				{
					ID:           "menu.workloads",
					EntryKey:     "workloads",
					GroupKey:     "core",
					Availability: sdk.MenuAvailabilityEnabled,
					Order:        20,
				},
			},
		},
		{
			ID: "plugin.ops-console",
			Menus: []sdk.MenuDescriptor{
				{
					ID:           "menu.ops-console",
					EntryKey:     "ops-console",
					GroupKey:     "extensions",
					Availability: sdk.MenuAvailabilityEnabled,
					Order:        5,
				},
			},
		},
	}

	menus := ComposeMenus(descriptors)
	if len(menus) != 3 {
		t.Fatalf("expected 3 menus including fallback, got %d", len(menus))
	}
	if menus[0].GroupKey != "core" || menus[0].EntryKey != "workloads" {
		t.Fatalf("expected core workloads menu first, got %#v", menus[0])
	}
	if menus[1].GroupKey != "extensions" || menus[1].EntryKey != "ops-console" {
		t.Fatalf("expected extensions menu second, got %#v", menus[1])
	}
	if menus[2].EntryKey != "crds" {
		t.Fatalf("expected fallback crds menu last, got %#v", menus[2])
	}
	if menus[2].GroupKey != "resources" {
		t.Fatalf("expected fallback crds menu to be in resources group, got %#v", menus[2])
	}
	if !menus[2].IsFallback {
		t.Fatalf("expected crds menu to be marked as fallback")
	}
}
