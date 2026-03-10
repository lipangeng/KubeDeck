package plugins

import "kubedeck/backend/pkg/sdk"

func ComposeMenuComposition(
	descriptors []sdk.CapabilityDescriptor,
	globalOverrides []MenuOverride,
	clusterOverrides []MenuOverride,
) MenuComposition {
	blueprint := defaultMenuBlueprint()
	mounts := buildMenuMounts(descriptors)
	resolved, usedMounts := composeBlueprintEntries(blueprint, mounts)
	resolved = appendUnconfiguredMounts(resolved, mounts, usedMounts)

	overrides := make([]MenuOverride, 0, len(globalOverrides)+len(clusterOverrides))
	overrides = append(overrides, globalOverrides...)
	overrides = append(overrides, clusterOverrides...)
	resolved = applyMenuOverrides(resolved, overrides)

	return MenuComposition{
		Blueprint: blueprint,
		Mounts:    mounts,
		Overrides: overrides,
		Groups:    buildMenuGroups(blueprint, resolved),
	}
}

func ComposeMenus(descriptors []sdk.CapabilityDescriptor) []sdk.MenuDescriptor {
	composition := ComposeMenuComposition(descriptors, nil, nil)
	menus := make([]sdk.MenuDescriptor, 0)
	for _, group := range composition.Groups {
		for _, entry := range group.Entries {
			menus = append(menus, sdk.MenuDescriptor{
				ID:               entry.ID,
				WorkflowDomainID: entry.WorkflowDomainID,
				EntryKey:         entry.EntryKey,
				GroupKey:         entry.GroupKey,
				Route:            entry.Route,
				Placement:        entry.Placement,
				Availability:     entry.Availability,
				IsFallback:       entry.IsFallback,
				Order:            entry.Order,
				Visible:          entry.Visible,
				Title:            entry.Title,
				Description:      entry.Description,
			})
		}
	}
	return menus
}
