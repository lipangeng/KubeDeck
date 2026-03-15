package sdk

type ResourcePageExtensionCapabilityType string

const (
	ResourcePageExtensionTab       ResourcePageExtensionCapabilityType = "tab"
	ResourcePageExtensionTabReplace ResourcePageExtensionCapabilityType = "tab-replace"
	ResourcePageExtensionPageTakeover ResourcePageExtensionCapabilityType = "page-takeover"
	ResourcePageExtensionAction     ResourcePageExtensionCapabilityType = "action"
	ResourcePageExtensionSlot      ResourcePageExtensionCapabilityType = "slot"
)

// ResourcePageExtensionDescriptor declares one resource-page capability exposed by a backend capability.
type ResourcePageExtensionDescriptor struct {
	Kind            string
	CapabilityType  ResourcePageExtensionCapabilityType
	TargetTabID     string
	TabID           string
	ActionID        string
	Placement       string
	Priority        int
	Title           TextRef
	ContentFallback string
}
