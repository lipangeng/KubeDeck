package builtins

import "kubedeck/backend/pkg/sdk"

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
				Route:            "/workloads",
				Placement:        sdk.MenuPlacementPrimary,
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
