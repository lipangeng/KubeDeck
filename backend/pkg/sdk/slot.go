package sdk

type SlotPlacement string

const (
	SlotPlacementSummary SlotPlacement = "summary"
	SlotPlacementPanel   SlotPlacement = "panel"
	SlotPlacementToolbar SlotPlacement = "toolbar"
	SlotPlacementContext SlotPlacement = "context"
)

// SlotDescriptor declares one frontend-visible extension slot.
type SlotDescriptor struct {
	ID               string
	WorkflowDomainID string
	SlotID           string
	Placement        SlotPlacement
	Visible          bool
	Title            *TextRef
}
