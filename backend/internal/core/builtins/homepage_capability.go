package builtins

import "kubedeck/backend/pkg/sdk"

type HomepageCapability struct{}

func (HomepageCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.homepage",
		Version: "v1",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.homepage",
				WorkflowDomainID: "homepage",
				Route:            "/",
				EntryKey:         "homepage",
				Title:            sdk.TextRef{Key: "homepage.title", Fallback: "Homepage"},
			},
		},
		Menus: []sdk.MenuDescriptor{
			{
				ID:               "menu.homepage",
				WorkflowDomainID: "homepage",
				EntryKey:         "homepage",
				Route:            "/",
				Placement:        sdk.MenuPlacementPrimary,
				Order:            10,
				Visible:          true,
				Title:            sdk.TextRef{Key: "homepage.title", Fallback: "Homepage"},
			},
		},
	}
}
