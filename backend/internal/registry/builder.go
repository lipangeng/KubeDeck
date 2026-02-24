package registry

// BuildSnapshot merges base system metadata with dynamic metadata.
func BuildSnapshot(system Snapshot, dynamic Snapshot) Snapshot {
	pages := append(append([]PageMeta{}, system.Pages...), dynamic.Pages...)
	for i := range pages {
		pages[i].Slots = append([]string{}, pages[i].Slots...)
	}

	menus := append(append([]MenuItem{}, system.Menus...), dynamic.Menus...)
	for i := range menus {
		menus[i].PermissionHints = append([]string{}, menus[i].PermissionHints...)
	}

	return Snapshot{
		ResourceTypes: append(append([]ResourceType{}, system.ResourceTypes...), dynamic.ResourceTypes...),
		Pages:         pages,
		Slots:         append(append([]SlotMeta{}, system.Slots...), dynamic.Slots...),
		Menus:         menus,
	}
}
