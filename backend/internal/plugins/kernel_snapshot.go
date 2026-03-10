package plugins

import "kubedeck/backend/pkg/sdk"

type KernelSnapshot struct {
	Pages         []sdk.PageDescriptor   `json:"pages"`
	Menus         []sdk.MenuDescriptor   `json:"menus"`
	MenuBlueprint MenuBlueprint          `json:"menuBlueprint"`
	MenuMounts    []MenuMount            `json:"menuMounts"`
	MenuOverrides []MenuOverride         `json:"menuOverrides"`
	MenuGroups    []MenuResolvedGroup    `json:"menuGroups"`
	Actions       []sdk.ActionDescriptor `json:"actions"`
	Slots         []sdk.SlotDescriptor   `json:"slots"`
}

func ComposeKernelSnapshot(descriptors []sdk.CapabilityDescriptor) KernelSnapshot {
	return ComposeKernelSnapshotWithOverrides(descriptors, nil, nil)
}

func ComposeKernelSnapshotWithOverrides(
	descriptors []sdk.CapabilityDescriptor,
	globalOverrides []MenuOverride,
	clusterOverrides []MenuOverride,
) KernelSnapshot {
	menuComposition := ComposeMenuComposition(descriptors, globalOverrides, clusterOverrides)
	return KernelSnapshot{
		Pages:         ComposePages(descriptors),
		Menus:         flattenMenuGroups(menuComposition.Groups),
		MenuBlueprint: menuComposition.Blueprint,
		MenuMounts:    menuComposition.Mounts,
		MenuOverrides: menuComposition.Overrides,
		MenuGroups:    menuComposition.Groups,
		Actions:       ComposeActions(descriptors),
		Slots:         ComposeSlots(descriptors),
	}
}

func flattenMenuGroups(groups []MenuResolvedGroup) []sdk.MenuDescriptor {
	menus := make([]sdk.MenuDescriptor, 0)
	for _, group := range groups {
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
