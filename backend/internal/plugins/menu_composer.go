package plugins

import (
	"sort"

	"kubedeck/backend/pkg/sdk"
)

func ComposeMenus(descriptors []sdk.CapabilityDescriptor) []sdk.MenuDescriptor {
	menus := make([]sdk.MenuDescriptor, 0)
	for _, descriptor := range descriptors {
		menus = append(menus, descriptor.Menus...)
	}

	sort.SliceStable(menus, func(i, j int) bool {
		if menus[i].Order != menus[j].Order {
			return menus[i].Order < menus[j].Order
		}
		return menus[i].ID < menus[j].ID
	})

	return menus
}
