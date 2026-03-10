package builtins

import "kubedeck/backend/pkg/sdk"

type OperationsCapability struct{}

func (OperationsCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.operations",
		Version: "v1",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.operations",
				WorkflowDomainID: "operations",
				Route:            "/operations",
				EntryKey:         "operations",
				Title:            sdk.TextRef{Key: "operations.title", Fallback: "Operations"},
				Description: &sdk.TextRef{
					Key:      "operations.description",
					Fallback: "This page is composed from backend capability metadata and rendered through the generic runtime page.",
				},
			},
		},
		Menus: []sdk.MenuDescriptor{
			{
				ID:               "menu.operations",
				WorkflowDomainID: "operations",
				EntryKey:         "operations",
				Route:            "/operations",
				Placement:        sdk.MenuPlacementPrimary,
				Order:            30,
				Visible:          true,
				Title:            sdk.TextRef{Key: "operations.title", Fallback: "Operations"},
			},
		},
	}
}
