package builtins

import (
	"context"
	"time"

	"kubedeck/backend/internal/k8s"
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
	clientManager := k8s.GlobalClientManager()
	
	var clientset interface{}
	var err error
	
	if cluster != "" {
		clientset, err = clientManager.GetClient(cluster)
	} else {
		clientset, err = clientManager.GetCurrentClient()
	}
	
	if err != nil {
		return getMockWorkloads(cluster)
	}
	
	provider := k8s.NewWorkloadsProvider(clientset.(*k8s.Clientset), "default")
	ctx := context.Background()
	
	workloads, err := provider.ListWorkloads(ctx, "default")
	if err != nil {
		return getMockWorkloads(cluster)
	}
	
	result := make([]sdk.WorkloadItem, 0, len(workloads))
	for _, w := range workloads {
		result = append(result, sdk.WorkloadItem{
			ID:        w.ID,
			Name:      w.Name,
			Kind:      w.Kind,
			Namespace: w.Namespace,
			Status:    w.Status,
			Health:    w.Health,
			UpdatedAt: w.UpdatedAt,
		})
	}
	
	return result
}

func getMockWorkloads(cluster string) []sdk.WorkloadItem {
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
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:        "workload-web-" + suffix,
			Name:      "web",
			Kind:      "Deployment",
			Namespace: "default",
			Status:    "Pending",
			Health:    "Warning",
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	}
}
