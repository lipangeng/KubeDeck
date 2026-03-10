package sdk

// ActionSurface declares how an action should be presented by the frontend.
type ActionSurface string

const (
	ActionSurfaceDrawer ActionSurface = "drawer"
	ActionSurfaceDialog ActionSurface = "dialog"
	ActionSurfaceInline ActionSurface = "inline"
	ActionSurfacePage   ActionSurface = "page"
)

// ActionDescriptor declares one executable workflow action.
type ActionDescriptor struct {
	ID               string
	WorkflowDomainID string
	Surface          ActionSurface
	Visible          bool
	PermissionHint   string
	Title            TextRef
	Description      *TextRef
}

// ActionTarget declares the backend-authoritative execution target.
type ActionTarget struct {
	Cluster   string
	Namespace string
	Scope     string
}

// ActionExecutionRequest is the kernel-facing shape for one action execution.
type ActionExecutionRequest struct {
	ActionID         string
	WorkflowDomainID string
	Target           ActionTarget
	Input            map[string]any
}

// ActionExecutionResult captures the minimum result summary shape.
type ActionExecutionResult struct {
	Accepted        bool
	Summary         string
	AffectedObjects []string
	FailedObjects   []string
}
