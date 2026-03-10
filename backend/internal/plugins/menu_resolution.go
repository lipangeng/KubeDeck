package plugins

import "kubedeck/backend/pkg/sdk"

func appendFallbackMenus(menus []sdk.MenuDescriptor) []sdk.MenuDescriptor {
	for _, menu := range menus {
		if menu.EntryKey == "crds" {
			return menus
		}
	}

	return append(menus, sdk.MenuDescriptor{
		ID:               "menu.crds",
		WorkflowDomainID: "crds",
		EntryKey:         "crds",
		GroupKey:         "resources",
		Route:            "/resources/crds",
		Placement:        sdk.MenuPlacementSecondary,
		Availability:     sdk.MenuAvailabilityEnabled,
		IsFallback:       true,
		Order:            999,
		Visible:          true,
		Title:            sdk.TextRef{Key: "resources.crds.title", Fallback: "CRDs"},
	})
}
