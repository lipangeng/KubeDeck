package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"

	"kubedeck/backend/pkg/sdk"
)

type manifestTextRef struct {
	Key         string `json:"key"`
	Fallback    string `json:"fallback"`
	Description string `json:"description"`
}

type manifestPage struct {
	ID               string           `json:"id"`
	WorkflowDomainID string           `json:"workflowDomainId"`
	Route            string           `json:"route"`
	EntryKey         string           `json:"entryKey"`
	Title            manifestTextRef  `json:"title"`
	Description      *manifestTextRef `json:"description"`
}

type manifestMenu struct {
	ID               string           `json:"id"`
	WorkflowDomainID string           `json:"workflowDomainId"`
	EntryKey         string           `json:"entryKey"`
	Route            string           `json:"route"`
	Placement        string           `json:"placement"`
	Order            int              `json:"order"`
	Visible          bool             `json:"visible"`
	Title            manifestTextRef  `json:"title"`
	Description      *manifestTextRef `json:"description"`
}

type manifestAction struct {
	ID               string           `json:"id"`
	WorkflowDomainID string           `json:"workflowDomainId"`
	Surface          string           `json:"surface"`
	Visible          bool             `json:"visible"`
	PermissionHint   string           `json:"permissionHint"`
	Title            manifestTextRef  `json:"title"`
	Description      *manifestTextRef `json:"description"`
}

type manifestSlot struct {
	ID               string           `json:"id"`
	WorkflowDomainID string           `json:"workflowDomainId"`
	SlotID           string           `json:"slotId"`
	Placement        string           `json:"placement"`
	Visible          bool             `json:"visible"`
	Title            *manifestTextRef `json:"title"`
}

type manifestResourcePageExtension struct {
	Kind            string          `json:"kind"`
	CapabilityType  string          `json:"capabilityType"`
	TargetTabID     string          `json:"targetTabId"`
	TabID           string          `json:"tabId"`
	ActionID        string          `json:"actionId"`
	Priority        int             `json:"priority"`
	Title           manifestTextRef `json:"title"`
	ContentFallback string          `json:"contentFallback"`
}

type manifestContributions struct {
	Pages                  []manifestPage                  `json:"pages"`
	Menus                  []manifestMenu                  `json:"menus"`
	Actions                []manifestAction                `json:"actions"`
	Slots                  []manifestSlot                  `json:"slots"`
	ResourcePageExtensions []manifestResourcePageExtension `json:"resourcePageExtensions"`
}

type pluginManifest struct {
	PluginID      string                `json:"pluginId"`
	Version       string                `json:"version"`
	DisplayName   string                `json:"displayName"`
	Contributions manifestContributions `json:"contributions"`
}

type manifestCapabilityProvider struct {
	descriptor sdk.CapabilityDescriptor
}

func (p manifestCapabilityProvider) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return p.descriptor
}

func LoadManifestProvidersFromDir(root string) ([]sdk.CapabilityProvider, error) {
	if root == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	providers := make([]sdk.CapabilityProvider, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(root, entry.Name(), "plugin.manifest.json")
		content, err := os.ReadFile(manifestPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		var manifest pluginManifest
		if err := json.Unmarshal(content, &manifest); err != nil {
			return nil, err
		}

		providers = append(providers, manifestCapabilityProvider{
			descriptor: sdk.CapabilityDescriptor{
				ID:      manifest.PluginID,
				Version: manifest.Version,
				Pages:   toPageDescriptors(manifest.Contributions.Pages),
				Menus:   toMenuDescriptors(manifest.Contributions.Menus),
				Actions: toActionDescriptors(manifest.Contributions.Actions),
				Slots:   toSlotDescriptors(manifest.Contributions.Slots),
				ResourcePageExtensions: toResourcePageExtensionDescriptors(
					manifest.Contributions.ResourcePageExtensions,
				),
			},
		})
	}

	return providers, nil
}

func toTextRef(value manifestTextRef) sdk.TextRef {
	return sdk.TextRef{
		Key:         value.Key,
		Fallback:    value.Fallback,
		Description: value.Description,
	}
}

func toPageDescriptors(items []manifestPage) []sdk.PageDescriptor {
	pages := make([]sdk.PageDescriptor, 0, len(items))
	for _, item := range items {
		page := sdk.PageDescriptor{
			ID:               item.ID,
			WorkflowDomainID: item.WorkflowDomainID,
			Route:            item.Route,
			EntryKey:         item.EntryKey,
			Title:            toTextRef(item.Title),
		}
		if item.Description != nil {
			description := toTextRef(*item.Description)
			page.Description = &description
		}
		pages = append(pages, page)
	}
	return pages
}

func toMenuDescriptors(items []manifestMenu) []sdk.MenuDescriptor {
	menus := make([]sdk.MenuDescriptor, 0, len(items))
	for _, item := range items {
		menu := sdk.MenuDescriptor{
			ID:               item.ID,
			WorkflowDomainID: item.WorkflowDomainID,
			EntryKey:         item.EntryKey,
			Route:            item.Route,
			Placement:        sdk.MenuPlacement(item.Placement),
			Order:            item.Order,
			Visible:          item.Visible,
			Title:            toTextRef(item.Title),
		}
		if item.Description != nil {
			description := toTextRef(*item.Description)
			menu.Description = &description
		}
		menus = append(menus, menu)
	}
	return menus
}

func toActionDescriptors(items []manifestAction) []sdk.ActionDescriptor {
	actions := make([]sdk.ActionDescriptor, 0, len(items))
	for _, item := range items {
		action := sdk.ActionDescriptor{
			ID:               item.ID,
			WorkflowDomainID: item.WorkflowDomainID,
			Surface:          sdk.ActionSurface(item.Surface),
			Visible:          item.Visible,
			PermissionHint:   item.PermissionHint,
			Title:            toTextRef(item.Title),
		}
		if item.Description != nil {
			description := toTextRef(*item.Description)
			action.Description = &description
		}
		actions = append(actions, action)
	}
	return actions
}

func toSlotDescriptors(items []manifestSlot) []sdk.SlotDescriptor {
	slots := make([]sdk.SlotDescriptor, 0, len(items))
	for _, item := range items {
		slot := sdk.SlotDescriptor{
			ID:               item.ID,
			WorkflowDomainID: item.WorkflowDomainID,
			SlotID:           item.SlotID,
			Placement:        sdk.SlotPlacement(item.Placement),
			Visible:          item.Visible,
		}
		if item.Title != nil {
			title := toTextRef(*item.Title)
			slot.Title = &title
		}
		slots = append(slots, slot)
	}
	return slots
}

func toResourcePageExtensionDescriptors(
	items []manifestResourcePageExtension,
) []sdk.ResourcePageExtensionDescriptor {
	extensions := make([]sdk.ResourcePageExtensionDescriptor, 0, len(items))
	for _, item := range items {
		extensions = append(extensions, sdk.ResourcePageExtensionDescriptor{
			Kind:            item.Kind,
			CapabilityType:  sdk.ResourcePageExtensionCapabilityType(item.CapabilityType),
			TargetTabID:     item.TargetTabID,
			TabID:           item.TabID,
			ActionID:        item.ActionID,
			Priority:        item.Priority,
			Title:           toTextRef(item.Title),
			ContentFallback: item.ContentFallback,
		})
	}
	return extensions
}
