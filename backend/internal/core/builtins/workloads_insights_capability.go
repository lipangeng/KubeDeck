package builtins

import "kubedeck/backend/pkg/sdk"

type WorkloadsInsightsCapability struct{}

func (WorkloadsInsightsCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.workloads-insights",
		Version: "v1",
		Slots: []sdk.SlotDescriptor{
			{
				ID:               "slot.workloads.summary.insights",
				WorkflowDomainID: "workloads",
				SlotID:           "workloads.summary.insights",
				Placement:        sdk.SlotPlacementSummary,
				Visible:          true,
				Title: &sdk.TextRef{
					Key:      "workloads.insights.title",
					Fallback: "Kernel Insights",
				},
			},
		},
	}
}
