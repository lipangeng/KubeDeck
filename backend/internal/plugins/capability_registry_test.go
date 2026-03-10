package plugins

import (
	"testing"

	"kubedeck/backend/pkg/sdk"
)

type testProvider struct {
	descriptor sdk.CapabilityDescriptor
}

func (p testProvider) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return p.descriptor
}

func TestCapabilityRegistryRegistersAndListsDescriptors(t *testing.T) {
	registry := NewCapabilityRegistry()
	err := registry.Register(testProvider{descriptor: sdk.CapabilityDescriptor{ID: "core.homepage"}})
	if err != nil {
		t.Fatalf("register capability: %v", err)
	}

	descriptors := registry.Descriptors()
	if len(descriptors) != 1 {
		t.Fatalf("expected 1 descriptor, got %d", len(descriptors))
	}
	if descriptors[0].ID != "core.homepage" {
		t.Fatalf("expected descriptor id core.homepage, got %q", descriptors[0].ID)
	}
}

func TestCapabilityRegistryRejectsDuplicateDescriptors(t *testing.T) {
	registry := NewCapabilityRegistry()
	provider := testProvider{descriptor: sdk.CapabilityDescriptor{ID: "core.homepage"}}
	if err := registry.Register(provider); err != nil {
		t.Fatalf("register capability: %v", err)
	}
	if err := registry.Register(provider); err == nil {
		t.Fatalf("expected duplicate capability registration to fail")
	}
}
