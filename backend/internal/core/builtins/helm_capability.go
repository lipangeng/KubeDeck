package builtins

import "kubedeck/backend/pkg/sdk"

type HelmCapability struct{}

func (HelmCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.helm",
		Version: "v1",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.helm",
				WorkflowDomainID: "helm",
				Route:            "/helm",
				EntryKey:         "helm",
				Title:            sdk.TextRef{Key: "helm.title", Fallback: "Helm Charts"},
			},
		},
		Menus: []sdk.MenuDescriptor{
			{
				ID:               "menu.helm",
				WorkflowDomainID: "helm",
				EntryKey:         "helm",
				GroupKey:         "core",
				Route:            "/helm",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            55,
				Visible:          true,
				Title:            sdk.TextRef{Key: "helm.title", Fallback: "Helm Charts"},
			},
		},
	}
}

func (HelmCapability) WorkflowDomainID() string {
	return "helm"
}
