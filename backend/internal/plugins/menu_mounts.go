package plugins

import "kubedeck/backend/pkg/sdk"

func buildMenuMounts(descriptors []sdk.CapabilityDescriptor) []MenuMount {
	mounts := make([]MenuMount, 0)
	for _, descriptor := range descriptors {
		sourceType := inferMenuMountSourceType(descriptor.ID)
		for _, menu := range descriptor.Menus {
			groupKey := menu.GroupKey
			if groupKey == "" {
				groupKey = "extensions"
			}
			mounts = append(mounts, MenuMount{
				ID:               menu.ID,
				CapabilityID:     descriptor.ID,
				SourceType:       sourceType,
				WorkflowDomainID: menu.WorkflowDomainID,
				EntryKey:         menu.EntryKey,
				GroupKey:         groupKey,
				Route:            menu.Route,
				Placement:        menu.Placement,
				Availability:     menu.Availability,
				IsFallback:       menu.IsFallback,
				Order:            menu.Order,
				Visible:          menu.Visible,
				Title:            menu.Title,
				Description:      menu.Description,
			})
		}
	}
	return mounts
}
