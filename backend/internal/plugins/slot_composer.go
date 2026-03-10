package plugins

import "kubedeck/backend/pkg/sdk"

func ComposeSlots(descriptors []sdk.CapabilityDescriptor) []sdk.SlotDescriptor {
	slots := make([]sdk.SlotDescriptor, 0)
	for _, descriptor := range descriptors {
		slots = append(slots, descriptor.Slots...)
	}
	return slots
}
