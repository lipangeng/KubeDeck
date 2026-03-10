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
	menuComposition := ComposeMenuComposition(descriptors, nil, nil)
	return KernelSnapshot{
		Pages:         ComposePages(descriptors),
		Menus:         ComposeMenus(descriptors),
		MenuBlueprint: menuComposition.Blueprint,
		MenuMounts:    menuComposition.Mounts,
		MenuOverrides: menuComposition.Overrides,
		MenuGroups:    menuComposition.Groups,
		Actions:       ComposeActions(descriptors),
		Slots:         ComposeSlots(descriptors),
	}
}
