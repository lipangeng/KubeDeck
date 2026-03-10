package plugins

import "kubedeck/backend/pkg/sdk"

func defaultMenuBlueprint() MenuBlueprint {
	return MenuBlueprint{
		Groups: []MenuBlueprintGroup{
			{
				Key:   "core",
				Order: 10,
				Title: sdk.TextRef{Key: "menu.group.core", Fallback: "Core"},
			},
			{
				Key:   "platform",
				Order: 20,
				Title: sdk.TextRef{Key: "menu.group.platform", Fallback: "Platform"},
			},
			{
				Key:   "extensions",
				Order: 30,
				Title: sdk.TextRef{Key: "menu.group.extensions", Fallback: "Extensions"},
			},
			{
				Key:   "resources",
				Order: 40,
				Title: sdk.TextRef{Key: "menu.group.resources", Fallback: "Resources"},
			},
		},
		Entries: []MenuBlueprintEntry{
			{
				EntryKey:         "homepage",
				WorkflowDomainID: "homepage",
				DefaultGroupKey:  "core",
				Route:            "/",
				Order:            10,
				Placement:        sdk.MenuPlacementPrimary,
				Title:            sdk.TextRef{Key: "homepage.title", Fallback: "Homepage"},
				SourceType:       MenuMountSourceBuiltin,
			},
			{
				EntryKey:         "workloads",
				WorkflowDomainID: "workloads",
				DefaultGroupKey:  "core",
				Route:            "/workloads",
				Order:            20,
				Placement:        sdk.MenuPlacementPrimary,
				Title:            sdk.TextRef{Key: "workloads.title", Fallback: "Workloads"},
				SourceType:       MenuMountSourceBuiltin,
			},
			{
				EntryKey:         "services",
				WorkflowDomainID: "services",
				DefaultGroupKey:  "core",
				Route:            "/services",
				Order:            30,
				Placement:        sdk.MenuPlacementPrimary,
				Title:            sdk.TextRef{Key: "services.title", Fallback: "Services"},
				SourceType:       MenuMountSourceBuiltin,
			},
			{
				EntryKey:         "config",
				WorkflowDomainID: "config",
				DefaultGroupKey:  "core",
				Route:            "/config",
				Order:            40,
				Placement:        sdk.MenuPlacementPrimary,
				Title:            sdk.TextRef{Key: "config.title", Fallback: "Config"},
				SourceType:       MenuMountSourceBuiltin,
			},
			{
				EntryKey:         "secrets",
				WorkflowDomainID: "secrets",
				DefaultGroupKey:  "core",
				Route:            "/secrets",
				Order:            50,
				Placement:        sdk.MenuPlacementPrimary,
				Title:            sdk.TextRef{Key: "secrets.title", Fallback: "Secrets"},
				SourceType:       MenuMountSourceBuiltin,
			},
			{
				EntryKey:         "operations",
				WorkflowDomainID: "operations",
				DefaultGroupKey:  "platform",
				Route:            "/operations",
				Order:            10,
				Placement:        sdk.MenuPlacementPrimary,
				Title:            sdk.TextRef{Key: "operations.title", Fallback: "Operations"},
				SourceType:       MenuMountSourceBuiltin,
			},
			{
				EntryKey:         "crds",
				WorkflowDomainID: "crds",
				DefaultGroupKey:  "resources",
				Route:            "/resources/crds",
				Order:            999,
				Placement:        sdk.MenuPlacementSecondary,
				Title:            sdk.TextRef{Key: "resources.crds.title", Fallback: "CRDs"},
				SourceType:       MenuMountSourceFallback,
				IsFallback:       true,
			},
		},
	}
}
