package plugins

import "kubedeck/backend/pkg/sdk"

type KernelSnapshot struct {
	Pages   []sdk.PageDescriptor   `json:"pages"`
	Menus   []sdk.MenuDescriptor   `json:"menus"`
	Actions []sdk.ActionDescriptor `json:"actions"`
}

func ComposeKernelSnapshot(descriptors []sdk.CapabilityDescriptor) KernelSnapshot {
	return KernelSnapshot{
		Pages:   ComposePages(descriptors),
		Menus:   ComposeMenus(descriptors),
		Actions: ComposeActions(descriptors),
	}
}
