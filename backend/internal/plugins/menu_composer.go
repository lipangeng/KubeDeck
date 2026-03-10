package plugins

import (
	"sort"

	"kubedeck/backend/pkg/sdk"
)

func ComposeMenus(descriptors []sdk.CapabilityDescriptor) []sdk.MenuDescriptor {
	menus := appendFallbackMenus(buildMenuMounts(descriptors))
	groupOrder := make(map[string]int)
	for _, group := range defaultMenuBlueprint() {
		groupOrder[group.Key] = group.Order
	}

	sort.SliceStable(menus, func(i, j int) bool {
		leftGroupOrder, ok := groupOrder[menus[i].GroupKey]
		if !ok {
			leftGroupOrder = 1000
		}
		rightGroupOrder, ok := groupOrder[menus[j].GroupKey]
		if !ok {
			rightGroupOrder = 1000
		}
		if leftGroupOrder != rightGroupOrder {
			return leftGroupOrder < rightGroupOrder
		}
		if menus[i].Order != menus[j].Order {
			return menus[i].Order < menus[j].Order
		}
		return menus[i].ID < menus[j].ID
	})

	return menus
}
