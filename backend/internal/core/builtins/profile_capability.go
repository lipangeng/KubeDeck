package builtins

import "kubedeck/backend/pkg/sdk"

type ProfileCapability struct{}

func (ProfileCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.profile",
		Version: "v1",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.profile",
				WorkflowDomainID: "profile",
				Route:            "/profile",
				EntryKey:         "profile",
				Title:            sdk.TextRef{Key: "profile.title", Fallback: "Profile"},
			},
		},
		Menus: []sdk.MenuDescriptor{},
	}
}

func (ProfileCapability) WorkflowDomainID() string {
	return "profile"
}
