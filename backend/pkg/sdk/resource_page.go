package sdk

type ResourcePageExtensionCapabilityType string

const (
	ResourcePageExtensionTab       ResourcePageExtensionCapabilityType = "tab"
	ResourcePageExtensionTabReplace ResourcePageExtensionCapabilityType = "tab-replace"
)

// ResourcePageExtensionDescriptor declares one resource-page tab capability exposed by a backend capability.
type ResourcePageExtensionDescriptor struct {
	Kind            string
	CapabilityType  ResourcePageExtensionCapabilityType
	TargetTabID     string
	TabID           string
	Title           TextRef
	ContentFallback string
}
