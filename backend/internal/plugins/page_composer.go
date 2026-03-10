package plugins

import "kubedeck/backend/pkg/sdk"

func ComposePages(descriptors []sdk.CapabilityDescriptor) []sdk.PageDescriptor {
	pages := make([]sdk.PageDescriptor, 0)
	for _, descriptor := range descriptors {
		pages = append(pages, descriptor.Pages...)
	}
	return pages
}
