package plugins

import "kubedeck/backend/pkg/sdk"

func buildMenuMounts(descriptors []sdk.CapabilityDescriptor) []sdk.MenuDescriptor {
	mounts := make([]sdk.MenuDescriptor, 0)
	for _, descriptor := range descriptors {
		mounts = append(mounts, descriptor.Menus...)
	}
	return mounts
}
