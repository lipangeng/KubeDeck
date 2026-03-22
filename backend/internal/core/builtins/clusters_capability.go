package builtins

import (
	"time"

	"kubedeck/backend/pkg/sdk"
)

type ClustersCapability struct{}

func (ClustersCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.clusters",
		Version: "v1",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.clusters",
				WorkflowDomainID: "clusters",
				Route:            "/clusters",
				EntryKey:         "clusters",
				Title:            sdk.TextRef{Key: "clusters.title", Fallback: "Clusters"},
			},
		},
		Menus: []sdk.MenuDescriptor{
			{
				ID:               "menu.clusters",
				WorkflowDomainID: "clusters",
				EntryKey:         "clusters",
				GroupKey:         "core",
				Route:            "/clusters",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            15,
				Visible:          true,
				Title:            sdk.TextRef{Key: "clusters.title", Fallback: "Clusters"},
			},
		},
		Actions: []sdk.ActionDescriptor{
			{
				ID:               "add",
				WorkflowDomainID: "clusters",
				Surface:          sdk.ActionSurfaceDrawer,
				Visible:          true,
				Title:            sdk.TextRef{Key: "actions.addCluster", Fallback: "Add Cluster"},
			},
			{
				ID:               "remove",
				WorkflowDomainID: "clusters",
				Surface:          sdk.ActionSurfaceDrawer,
				Visible:          true,
				Title:            sdk.TextRef{Key: "actions.removeCluster", Fallback: "Remove"},
			},
		},
	}
}

func (ClustersCapability) WorkflowDomainID() string {
	return "clusters"
}

func (ClustersCapability) ListClusters() []sdk.ClusterItem {
	return []sdk.ClusterItem{
		{
			ID:        "default",
			Name:      "default",
			Status:    "Ready",
			Health:    "Healthy",
			Version:   "v1.28.0",
			Nodes:     3,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:        "prod-eu1",
			Name:      "prod-eu1",
			Status:    "Ready",
			Health:    "Healthy",
			Version:   "v1.28.0",
			Nodes:     5,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:        "staging-us1",
			Name:      "staging-us1",
			Status:    "Ready",
			Health:    "Warning",
			Version:   "v1.27.5",
			Nodes:     2,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	}
}

func (ClustersCapability) ExecuteAction(
	request sdk.ActionExecutionRequest,
) (sdk.ActionExecutionResult, error) {
	if request.WorkflowDomainID != "clusters" {
		return sdk.ActionExecutionResult{}, nil
	}

	switch request.ActionID {
	case "add":
		return sdk.ActionExecutionResult{
			Accepted:        true,
			Summary:         "Cluster add dialog opened",
			AffectedObjects: []string{"cluster/new"},
		}, nil
	case "remove":
		name := getTargetName(request)
		return sdk.ActionExecutionResult{
			Accepted:        true,
			Summary:         "Cluster remove initiated for " + name,
			AffectedObjects: []string{"cluster/" + name},
		}, nil
	default:
		return sdk.ActionExecutionResult{}, nil
	}
}

func getTargetName(request sdk.ActionExecutionRequest) string {
	if name, ok := request.Input["name"].(string); ok && name != "" {
		return name
	}
	return "unknown"
}
