package sdk

type ResourcePageExtensionCapabilityType string

const (
	ResourcePageExtensionTab       ResourcePageExtensionCapabilityType = "tab"
	ResourcePageExtensionTabReplace ResourcePageExtensionCapabilityType = "tab-replace"
	ResourcePageExtensionPageTakeover ResourcePageExtensionCapabilityType = "page-takeover"
)

// ResourcePageExtensionDescriptor declares one resource-page tab capability exposed by a backend capability.
type ResourcePageExtensionDescriptor struct {
	Kind            string
	CapabilityType  ResourcePageExtensionCapabilityType
	TargetTabID     string
	TabID           string
	Priority        int
	Title           TextRef
	ContentFallback string
}
