package sdk

// MenuPlacement describes where a menu item should appear in kernel-composed navigation.
type MenuPlacement string

const (
	MenuPlacementPrimary   MenuPlacement = "primary"
	MenuPlacementSecondary MenuPlacement = "secondary"
	MenuPlacementContext   MenuPlacement = "context"
)

// MenuAvailability describes whether a configured menu entry can be used in the current context.
type MenuAvailability string

const (
	MenuAvailabilityEnabled             MenuAvailability = "enabled"
	MenuAvailabilityDisabledUnavailable MenuAvailability = "disabled-unavailable"
	MenuAvailabilityHidden              MenuAvailability = "hidden"
)

// MenuDescriptor declares one navigation contribution.
type MenuDescriptor struct {
	ID               string
	WorkflowDomainID string
	EntryKey         string
	GroupKey         string
	Route            string
	Placement        MenuPlacement
	Availability     MenuAvailability
	IsFallback       bool
	Order            int
	Visible          bool
	Title            TextRef
	Description      *TextRef
}
