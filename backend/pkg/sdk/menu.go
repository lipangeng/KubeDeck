package sdk

// MenuPlacement describes where a menu item should appear in kernel-composed navigation.
type MenuPlacement string

const (
	MenuPlacementPrimary   MenuPlacement = "primary"
	MenuPlacementSecondary MenuPlacement = "secondary"
	MenuPlacementContext   MenuPlacement = "context"
)

// MenuDescriptor declares one navigation contribution.
type MenuDescriptor struct {
	ID               string
	WorkflowDomainID string
	EntryKey         string
	Route            string
	Placement        MenuPlacement
	Order            int
	Visible          bool
	Title            TextRef
	Description      *TextRef
}
