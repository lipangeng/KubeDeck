package sdk

// TextRef carries user-facing metadata in a future-localizable shape.
type TextRef struct {
	Key         string
	Fallback    string
	Description string
}

// CapabilityDescriptor declares what one built-in capability or plugin provides.
type CapabilityDescriptor struct {
	ID                     string
	Version                string
	Pages                  []PageDescriptor
	Menus                  []MenuDescriptor
	Actions                []ActionDescriptor
	Slots                  []SlotDescriptor
	ResourcePageExtensions []ResourcePageExtensionDescriptor
}

// CapabilityProvider exposes one capability descriptor to the kernel.
type CapabilityProvider interface {
	CapabilityDescriptor() CapabilityDescriptor
}

// WorkloadItem describes one workload row exposed by a workflow data provider.
type WorkloadItem struct {
	ID        string
	Name      string
	Kind      string
	Namespace string
	Status    string
	Health    string
	UpdatedAt string
}

// WorkloadProvider exposes workflow-owned workload data.
type WorkloadProvider interface {
	CapabilityProvider
	WorkflowDomainID() string
	ListWorkloads(cluster string) []WorkloadItem
}

// ActionExecutor executes kernel actions behind a capability boundary.
type ActionExecutor interface {
	CapabilityProvider
	ExecuteAction(ActionExecutionRequest) (ActionExecutionResult, error)
}
