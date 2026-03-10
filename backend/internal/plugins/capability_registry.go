package plugins

import (
	"fmt"

	"kubedeck/backend/pkg/sdk"
)

// CapabilityRegistry aggregates backend capability providers.
type CapabilityRegistry struct {
	providers map[string]sdk.CapabilityProvider
}

func NewCapabilityRegistry() *CapabilityRegistry {
	return &CapabilityRegistry{providers: map[string]sdk.CapabilityProvider{}}
}

func (r *CapabilityRegistry) Register(provider sdk.CapabilityProvider) error {
	if provider == nil {
		return fmt.Errorf("capability provider is nil")
	}

	descriptor := provider.CapabilityDescriptor()
	if descriptor.ID == "" {
		return fmt.Errorf("capability descriptor id is required")
	}
	if _, exists := r.providers[descriptor.ID]; exists {
		return fmt.Errorf("capability %q already registered", descriptor.ID)
	}

	r.providers[descriptor.ID] = provider
	return nil
}

func (r *CapabilityRegistry) Descriptors() []sdk.CapabilityDescriptor {
	descriptors := make([]sdk.CapabilityDescriptor, 0, len(r.providers))
	for _, provider := range r.providers {
		descriptors = append(descriptors, provider.CapabilityDescriptor())
	}
	return descriptors
}

func (r *CapabilityRegistry) Providers() []sdk.CapabilityProvider {
	providers := make([]sdk.CapabilityProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}
