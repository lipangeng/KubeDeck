package backendplugintemplate

import "kubedeck/backend/pkg/sdk"

type Capability struct{}

func New() sdk.CapabilityProvider {
	return Capability{}
}

func (Capability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "example-backend-plugin",
		Version: "0.1.0",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.example-backend-plugin",
				WorkflowDomainID: "example-backend-plugin",
				Route:            "/example-backend-plugin",
				EntryKey:         "example-backend-plugin",
				Title: sdk.TextRef{
					Key:      "exampleBackendPlugin.title",
					Fallback: "Example Backend Plugin",
				},
			},
		},
		Menus: []sdk.MenuDescriptor{
			{
				ID:               "menu.example-backend-plugin",
				WorkflowDomainID: "example-backend-plugin",
				EntryKey:         "example-backend-plugin",
				Route:            "/example-backend-plugin",
				Placement:        sdk.MenuPlacementPrimary,
				Order:            100,
				Visible:          true,
				Title: sdk.TextRef{
					Key:      "exampleBackendPlugin.title",
					Fallback: "Example Backend Plugin",
				},
			},
		},
		Actions: []sdk.ActionDescriptor{
			{
				ID:               "refresh-example-backend-plugin",
				WorkflowDomainID: "example-backend-plugin",
				Surface:          sdk.ActionSurfaceInline,
				Visible:          true,
				Title: sdk.TextRef{
					Key:      "exampleBackendPlugin.actions.refresh",
					Fallback: "Refresh Example Backend Plugin",
				},
			},
		},
		Slots: []sdk.SlotDescriptor{
			{
				ID:               "slot.example-backend-plugin.summary",
				WorkflowDomainID: "example-backend-plugin",
				SlotID:           "example-backend-plugin.summary",
				Placement:        sdk.SlotPlacementSummary,
				Visible:          true,
				Title: &sdk.TextRef{
					Key:      "exampleBackendPlugin.slots.summary",
					Fallback: "Example Backend Plugin Summary",
				},
			},
		},
		ResourcePageExtensions: []sdk.ResourcePageExtensionDescriptor{
			{
				Kind:           "StatefulSet",
				CapabilityType: sdk.ResourcePageExtensionPageTakeover,
				TabID:          "example.statefulset.takeover",
				Priority:       50,
				Title: sdk.TextRef{
					Key:      "exampleBackendPlugin.resource.statefulset",
					Fallback: "StatefulSet takeover",
				},
				ContentFallback: "Example backend plugin StatefulSet takeover",
			},
		},
	}
}
