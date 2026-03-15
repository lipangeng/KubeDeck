package plugins

import (
	"testing"

	"kubedeck/backend/pkg/sdk"
)

func TestComposeMenusBuildsBlueprintDrivenGroupsAndUnavailableEntries(t *testing.T) {
	descriptors := []sdk.CapabilityDescriptor{
		{
			ID: "core.homepage",
			Menus: []sdk.MenuDescriptor{
				{
					ID:               "menu.homepage",
					WorkflowDomainID: "homepage",
					EntryKey:         "homepage",
					GroupKey:         "core",
					Route:            "/",
					Placement:        sdk.MenuPlacementPrimary,
					Availability:     sdk.MenuAvailabilityEnabled,
					Order:            20,
					Visible:          true,
					Title:            sdk.TextRef{Key: "homepage.title", Fallback: "Homepage"},
				},
			},
		},
		{
			ID: "core.workloads",
			Menus: []sdk.MenuDescriptor{
				{
					ID:               "menu.workloads",
					WorkflowDomainID: "workloads",
					EntryKey:         "workloads",
					GroupKey:         "core",
					Route:            "/workloads",
					Placement:        sdk.MenuPlacementPrimary,
					Availability:     sdk.MenuAvailabilityEnabled,
					Order:            20,
					Visible:          true,
					Title:            sdk.TextRef{Key: "workloads.title", Fallback: "Workloads"},
				},
			},
		},
		{
			ID: "plugin.sample-ops-console",
			Menus: []sdk.MenuDescriptor{
				{
					ID:               "menu.sample-ops-console",
					WorkflowDomainID: "sample-ops-console",
					EntryKey:         "sample-ops-console",
					GroupKey:         "extensions",
					Route:            "/sample-ops-console",
					Placement:        sdk.MenuPlacementPrimary,
					Availability:     sdk.MenuAvailabilityEnabled,
					Order:            30,
					Visible:          true,
					Title:            sdk.TextRef{Key: "sampleOps.title", Fallback: "Sample Ops Console"},
				},
			},
		},
	}

	composed := ComposeMenuComposition(descriptors, nil, nil)
	if len(composed.Blueprint.Groups) != 4 {
		t.Fatalf("expected 4 blueprint groups, got %d", len(composed.Blueprint.Groups))
	}
	if composed.Blueprint.Groups[0].Title.Fallback != "Core" {
		t.Fatalf("expected first blueprint group title to be Core, got %#v", composed.Blueprint.Groups[0])
	}
	if len(composed.Groups) != 4 {
		t.Fatalf("expected 4 resolved groups, got %d", len(composed.Groups))
	}
	coreGroup := composed.Groups[0]
	if coreGroup.Key != "core" {
		t.Fatalf("expected first group key to be core, got %#v", coreGroup)
	}
	if coreGroup.Entries[0].EntryKey != "homepage" || coreGroup.Entries[1].EntryKey != "workloads" {
		t.Fatalf("expected homepage then workloads in core group, got %#v", coreGroup.Entries)
	}
	if coreGroup.Entries[2].EntryKey != "services" {
		t.Fatalf("expected blueprint-only services mount in core group, got %#v", coreGroup.Entries)
	}
	if coreGroup.Entries[2].Availability != sdk.MenuAvailabilityDisabledUnavailable {
		t.Fatalf("expected missing services mount to be disabled-unavailable, got %#v", coreGroup.Entries[2])
	}
	extensionsGroup := composed.Groups[2]
	if extensionsGroup.Key != "extensions" {
		t.Fatalf("expected extensions group in position 3, got %#v", extensionsGroup)
	}
	if len(extensionsGroup.Entries) != 1 || extensionsGroup.Entries[0].EntryKey != "sample-ops-console" {
		t.Fatalf("expected sample plugin mount in extensions, got %#v", extensionsGroup.Entries)
	}
	resourcesGroup := composed.Groups[3]
	if len(resourcesGroup.Entries) != 1 || resourcesGroup.Entries[0].EntryKey != "crds" || !resourcesGroup.Entries[0].IsFallback {
		t.Fatalf("expected fallback crds entry in resources group, got %#v", resourcesGroup.Entries)
	}
}

func TestComposeMenusAppliesGlobalThenClusterOverrides(t *testing.T) {
	descriptors := []sdk.CapabilityDescriptor{
		{
			ID: "core.workloads",
			Menus: []sdk.MenuDescriptor{
				{
					ID:               "menu.workloads",
					WorkflowDomainID: "workloads",
					EntryKey:         "workloads",
					GroupKey:         "core",
					Route:            "/workloads",
					Placement:        sdk.MenuPlacementPrimary,
					Availability:     sdk.MenuAvailabilityEnabled,
					Order:            20,
					Visible:          true,
					Title:            sdk.TextRef{Key: "workloads.title", Fallback: "Workloads"},
				},
			},
		},
		{
			ID: "plugin.sample-ops-console",
			Menus: []sdk.MenuDescriptor{
				{
					ID:               "menu.sample-ops-console",
					WorkflowDomainID: "sample-ops-console",
					EntryKey:         "sample-ops-console",
					GroupKey:         "extensions",
					Route:            "/sample-ops-console",
					Placement:        sdk.MenuPlacementPrimary,
					Availability:     sdk.MenuAvailabilityEnabled,
					Order:            30,
					Visible:          true,
					Title:            sdk.TextRef{Key: "sampleOps.title", Fallback: "Sample Ops Console"},
				},
			},
		},
	}

	globalOverrides := []MenuOverride{
		{
			Scope:           MenuOverrideScopeGlobal,
			MoveEntryKeys:   map[string]string{"sample-ops-console": "platform"},
			HiddenEntryKeys: []string{"secrets"},
		},
	}
	clusterOverrides := []MenuOverride{
		{
			Scope:         MenuOverrideScopeCluster,
			MoveEntryKeys: map[string]string{"sample-ops-console": "core"},
			PinEntryKeys:  []string{"sample-ops-console"},
		},
	}

	composed := ComposeMenuComposition(descriptors, globalOverrides, clusterOverrides)
	if len(composed.Overrides) != 2 {
		t.Fatalf("expected overrides to be preserved in composition, got %#v", composed.Overrides)
	}
	coreGroup := composed.Groups[0]
	if len(coreGroup.Entries) < 2 {
		t.Fatalf("expected moved entry to appear in core group, got %#v", coreGroup.Entries)
	}
	if coreGroup.Entries[0].EntryKey != "sample-ops-console" {
		t.Fatalf("expected cluster pin to move sample-ops-console to front of core group, got %#v", coreGroup.Entries)
	}
	for _, group := range composed.Groups {
		for _, entry := range group.Entries {
			if entry.EntryKey == "secrets" {
				t.Fatalf("expected global hidden entry to be removed from final groups, got %#v", composed.Groups)
			}
		}
	}
}

func TestComposeMenusSupportsScopedOverridesAndOrderingReservations(t *testing.T) {
	descriptors := []sdk.CapabilityDescriptor{
		{
			ID: "core.homepage",
			Menus: []sdk.MenuDescriptor{
				{
					ID:               "menu.homepage",
					WorkflowDomainID: "homepage",
					EntryKey:         "homepage",
					GroupKey:         "core",
					Route:            "/",
					Placement:        sdk.MenuPlacementPrimary,
					Availability:     sdk.MenuAvailabilityEnabled,
					Order:            10,
					Visible:          true,
					Title:            sdk.TextRef{Key: "homepage.title", Fallback: "Homepage"},
				},
			},
		},
		{
			ID: "core.workloads",
			Menus: []sdk.MenuDescriptor{
				{
					ID:               "menu.workloads",
					WorkflowDomainID: "workloads",
					EntryKey:         "workloads",
					GroupKey:         "core",
					Route:            "/workloads",
					Placement:        sdk.MenuPlacementPrimary,
					Availability:     sdk.MenuAvailabilityEnabled,
					Order:            20,
					Visible:          true,
					Title:            sdk.TextRef{Key: "workloads.title", Fallback: "Workloads"},
				},
			},
		},
		{
			ID: "plugin.sample-ops-console",
			Menus: []sdk.MenuDescriptor{
				{
					ID:               "menu.sample-ops-console",
					WorkflowDomainID: "sample-ops-console",
					EntryKey:         "sample-ops-console",
					GroupKey:         "extensions",
					Route:            "/sample-ops-console",
					Placement:        sdk.MenuPlacementPrimary,
					Availability:     sdk.MenuAvailabilityEnabled,
					Order:            30,
					Visible:          true,
					Title:            sdk.TextRef{Key: "sampleOps.title", Fallback: "Sample Ops Console"},
				},
			},
		},
	}

	workGlobalOverrides := []MenuOverride{
		{
			Scope:               MenuOverrideScope("work-global"),
			PinEntryKeys:        []string{"sample-ops-console"},
			GroupOrderOverrides: []string{"extensions", "core", "platform", "resources"},
			ItemOrderOverrides: map[string][]string{
				"core": {"workloads", "homepage", "services"},
			},
		},
	}
	workClusterOverrides := []MenuOverride{
		{
			Scope:           MenuOverrideScope("work-cluster"),
			HiddenEntryKeys: []string{"services"},
		},
	}

	composed := ComposeMenuComposition(descriptors, workGlobalOverrides, workClusterOverrides)

	if len(composed.Overrides) != 2 {
		t.Fatalf("expected scoped overrides to be preserved, got %#v", composed.Overrides)
	}
	if composed.Overrides[0].Scope != MenuOverrideScope("work-global") {
		t.Fatalf("expected work-global scope to survive composition, got %#v", composed.Overrides[0])
	}

	if composed.Groups[0].Key != "extensions" {
		t.Fatalf("expected group order override to move extensions first, got %#v", composed.Groups)
	}
	if composed.Groups[1].Key != "core" {
		t.Fatalf("expected core group after extensions, got %#v", composed.Groups)
	}

	coreGroup := composed.Groups[1]
	if len(coreGroup.Entries) < 2 {
		t.Fatalf("expected reordered core entries, got %#v", coreGroup.Entries)
	}
	if coreGroup.Entries[0].EntryKey != "workloads" || coreGroup.Entries[1].EntryKey != "homepage" {
		t.Fatalf("expected item order override to reorder core entries, got %#v", coreGroup.Entries)
	}
	for _, entry := range coreGroup.Entries {
		if entry.EntryKey == "services" {
			t.Fatalf("expected work-cluster hidden entry to be removed from groups, got %#v", coreGroup.Entries)
		}
	}
}
