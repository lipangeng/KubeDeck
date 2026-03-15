package plugins

import "kubedeck/backend/pkg/sdk"

func buildMenuMountsForScope(descriptors []sdk.CapabilityDescriptor, scope string) []MenuMount {
	switch scope {
	case "system":
		return []MenuMount{
			{
				ID:               "menu.system.menu-settings",
				CapabilityID:     "builtin.system.menu-settings",
				SourceType:       MenuMountSourceBuiltin,
				WorkflowDomainID: "system-menu-settings",
				EntryKey:         "menu-settings",
				GroupKey:         "config",
				Route:            "/settings/menu",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            10,
				Visible:          true,
				Title:            sdk.TextRef{Key: "settings.menu.title", Fallback: "System Menu Settings"},
			},
			{
				ID:               "menu.system.plugin-settings",
				CapabilityID:     "builtin.system.plugin-settings",
				SourceType:       MenuMountSourceBuiltin,
				WorkflowDomainID: "system-plugin-settings",
				EntryKey:         "plugin-settings",
				GroupKey:         "config",
				Route:            "/settings/plugins",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            20,
				Visible:          true,
				Title:            sdk.TextRef{Key: "settings.plugins.title", Fallback: "Plugin Settings"},
			},
		}
	case "cluster":
		return []MenuMount{
			{
				ID:               "menu.cluster.menu-settings",
				CapabilityID:     "builtin.cluster.menu-settings",
				SourceType:       MenuMountSourceBuiltin,
				WorkflowDomainID: "cluster-menu-settings",
				EntryKey:         "menu-settings",
				GroupKey:         "config",
				Route:            "/settings/menu",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            10,
				Visible:          true,
				Title:            sdk.TextRef{Key: "settings.menu.cluster.title", Fallback: "Cluster Menu Settings"},
			},
			{
				ID:               "menu.cluster.extensions",
				CapabilityID:     "builtin.cluster.extensions",
				SourceType:       MenuMountSourceBuiltin,
				WorkflowDomainID: "cluster-extensions",
				EntryKey:         "extensions",
				GroupKey:         "config",
				Route:            "/settings/extensions",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            20,
				Visible:          true,
				Title:            sdk.TextRef{Key: "settings.extensions.title", Fallback: "Cluster Extensions"},
			},
		}
	default:
		return buildMenuMounts(descriptors)
	}
}
