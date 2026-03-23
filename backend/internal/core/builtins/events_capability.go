package builtins

import (
	"time"

	"kubedeck/backend/pkg/sdk"
)

type EventsCapability struct{}

func (EventsCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.events",
		Version: "v1",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.events",
				WorkflowDomainID: "events",
				Route:            "/events",
				EntryKey:         "events",
				Title:            sdk.TextRef{Key: "events.title", Fallback: "Events"},
			},
		},
		Menus: []sdk.MenuDescriptor{
			{
				ID:               "menu.events",
				WorkflowDomainID: "events",
				EntryKey:         "events",
				GroupKey:         "core",
				Route:            "/events",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            50,
				Visible:          true,
				Title:            sdk.TextRef{Key: "events.title", Fallback: "Events"},
			},
		},
	}
}

func (EventsCapability) WorkflowDomainID() string {
	return "events"
}

func (EventsCapability) ListEvents(namespace string) []sdk.EventItem {
	// Return mock events for now
	return []sdk.EventItem{
		{
			ID:        "event-1",
			Name:      "event-1",
			Namespace: namespace,
			Type:      "Normal",
			Reason:    "Scheduled",
			Message:   "Successfully assigned default/api to node-1",
			Count:     1,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	}
}
