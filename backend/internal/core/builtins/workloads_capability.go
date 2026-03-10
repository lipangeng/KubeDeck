package builtins

import (
	"fmt"
	"time"

	"kubedeck/backend/pkg/sdk"
)

type WorkloadsCapability struct{}

func (WorkloadsCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.workloads",
		Version: "v1",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.workloads",
				WorkflowDomainID: "workloads",
				Route:            "/workloads",
				EntryKey:         "workloads",
				Title:            sdk.TextRef{Key: "workloads.title", Fallback: "Workloads"},
			},
		},
		Menus: []sdk.MenuDescriptor{
			{
				ID:               "menu.workloads",
				WorkflowDomainID: "workloads",
				EntryKey:         "workloads",
				GroupKey:         "core",
				Route:            "/workloads",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            20,
				Visible:          true,
				Title:            sdk.TextRef{Key: "workloads.title", Fallback: "Workloads"},
			},
		},
		Actions: []sdk.ActionDescriptor{
			{
				ID:               "create",
				WorkflowDomainID: "workloads",
				Surface:          sdk.ActionSurfaceDrawer,
				Visible:          true,
				Title:            sdk.TextRef{Key: "actions.create", Fallback: "Create"},
			},
			{
				ID:               "apply",
				WorkflowDomainID: "workloads",
				Surface:          sdk.ActionSurfaceDrawer,
				Visible:          true,
				Title:            sdk.TextRef{Key: "actions.apply", Fallback: "Apply"},
			},
		},
	}
}

func (WorkloadsCapability) WorkflowDomainID() string {
	return "workloads"
}

func (WorkloadsCapability) ListWorkloads(cluster string) []sdk.WorkloadItem {
	suffix := cluster
	if suffix == "" {
		suffix = "default"
	}

	return []sdk.WorkloadItem{
		{
			ID:        "workload-api-" + suffix,
			Name:      "api",
			Kind:      "Deployment",
			Namespace: "default",
			Status:    "Running",
			Health:    "Healthy",
			UpdatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		{
			ID:        "workload-web-" + suffix,
			Name:      "web",
			Kind:      "Deployment",
			Namespace: "default",
			Status:    "Pending",
			Health:    "Warning",
			UpdatedAt: time.Date(2026, 3, 10, 10, 5, 0, 0, time.UTC).Format(time.RFC3339),
		},
	}
}

func (WorkloadsCapability) ExecuteAction(
	request sdk.ActionExecutionRequest,
) (sdk.ActionExecutionResult, error) {
	if request.WorkflowDomainID != "workloads" {
		return sdk.ActionExecutionResult{}, nil
	}

	switch request.ActionID {
	case "create":
		return sdk.ActionExecutionResult{
			Accepted:        true,
			Summary:         "create accepted",
			AffectedObjects: []string{"deployment/" + targetName(request)},
		}, nil
	case "apply":
		return sdk.ActionExecutionResult{
			Accepted:        true,
			Summary:         "apply accepted",
			AffectedObjects: []string{"deployment/" + targetName(request)},
		}, nil
	default:
		return sdk.ActionExecutionResult{}, fmt.Errorf("unsupported action id %q", request.ActionID)
	}
}

func targetName(request sdk.ActionExecutionRequest) string {
	if name, ok := request.Input["name"].(string); ok && name != "" {
		return name
	}
	return "sample"
}
