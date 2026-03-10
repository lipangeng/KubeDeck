package plugins

import "kubedeck/backend/pkg/sdk"

func ComposeActions(descriptors []sdk.CapabilityDescriptor) []sdk.ActionDescriptor {
	actions := make([]sdk.ActionDescriptor, 0)
	for _, descriptor := range descriptors {
		actions = append(actions, descriptor.Actions...)
	}
	return actions
}
