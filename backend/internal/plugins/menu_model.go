package plugins

import (
	"sort"
	"strings"

	"kubedeck/backend/pkg/sdk"
)

type MenuMountSourceType string

const (
	MenuMountSourceBuiltin  MenuMountSourceType = "builtin"
	MenuMountSourcePlugin   MenuMountSourceType = "plugin"
	MenuMountSourceCRD      MenuMountSourceType = "crd"
	MenuMountSourceFallback MenuMountSourceType = "fallback"
)

type MenuBlueprintGroup struct {
	Key   string      `json:"key"`
	Order int         `json:"order"`
	Title sdk.TextRef `json:"title"`
}

type MenuBlueprintEntry struct {
	EntryKey         string              `json:"entryKey"`
	WorkflowDomainID string              `json:"workflowDomainId"`
	DefaultGroupKey  string              `json:"defaultGroupKey"`
	Route            string              `json:"route"`
	Order            int                 `json:"order"`
	Placement        sdk.MenuPlacement   `json:"placement"`
	Title            sdk.TextRef         `json:"title"`
	SourceType       MenuMountSourceType `json:"sourceType"`
	IsFallback       bool                `json:"isFallback,omitempty"`
}

type MenuBlueprint struct {
	Groups  []MenuBlueprintGroup `json:"groups"`
	Entries []MenuBlueprintEntry `json:"entries"`
}

type MenuMount struct {
	ID               string               `json:"id"`
	CapabilityID     string               `json:"capabilityId"`
	SourceType       MenuMountSourceType  `json:"sourceType"`
	WorkflowDomainID string               `json:"workflowDomainId"`
	EntryKey         string               `json:"entryKey"`
	GroupKey         string               `json:"groupKey"`
	Route            string               `json:"route"`
	Placement        sdk.MenuPlacement    `json:"placement"`
	Availability     sdk.MenuAvailability `json:"availability"`
	IsFallback       bool                 `json:"isFallback,omitempty"`
	Order            int                  `json:"order"`
	Visible          bool                 `json:"visible"`
	Title            sdk.TextRef          `json:"title"`
	Description      *sdk.TextRef         `json:"description,omitempty"`
}

type MenuOverrideScope string

const (
	MenuOverrideScopeGlobal      MenuOverrideScope = "global"
	MenuOverrideScopeCluster     MenuOverrideScope = "cluster"
	MenuOverrideScopeWorkGlobal  MenuOverrideScope = "work-global"
	MenuOverrideScopeWorkCluster MenuOverrideScope = "work-cluster"
	MenuOverrideScopeSystem      MenuOverrideScope = "system"
	MenuOverrideScopeClusterMenu MenuOverrideScope = "cluster"
)

type MenuOverride struct {
	Scope               MenuOverrideScope    `json:"scope"`
	HiddenEntryKeys     []string             `json:"hiddenEntryKeys,omitempty"`
	MoveEntryKeys       map[string]string    `json:"moveEntryKeys,omitempty"`
	PinEntryKeys        []string             `json:"pinEntryKeys,omitempty"`
	GroupOrderOverrides []string             `json:"groupOrderOverrides,omitempty"`
	ItemOrderOverrides  map[string][]string  `json:"itemOrderOverrides,omitempty"`
}

type MenuResolvedEntry struct {
	ID               string               `json:"id"`
	CapabilityID     string               `json:"capabilityId"`
	SourceType       MenuMountSourceType  `json:"sourceType"`
	WorkflowDomainID string               `json:"workflowDomainId"`
	EntryKey         string               `json:"entryKey"`
	GroupKey         string               `json:"groupKey"`
	Route            string               `json:"route"`
	Placement        sdk.MenuPlacement    `json:"placement"`
	Availability     sdk.MenuAvailability `json:"availability"`
	IsFallback       bool                 `json:"isFallback,omitempty"`
	Order            int                  `json:"order"`
	Visible          bool                 `json:"visible"`
	Title            sdk.TextRef          `json:"title"`
	Description      *sdk.TextRef         `json:"description,omitempty"`
	Mounted          bool                 `json:"mounted"`
	Configured       bool                 `json:"configured"`
	Pinned           bool                 `json:"pinned,omitempty"`
	SortOrder        int                  `json:"-"`
}

type MenuResolvedGroup struct {
	Key     string              `json:"key"`
	Order   int                 `json:"order"`
	Title   sdk.TextRef         `json:"title"`
	Entries []MenuResolvedEntry `json:"entries"`
}

type MenuComposition struct {
	Blueprint MenuBlueprint       `json:"menuBlueprint"`
	Mounts    []MenuMount         `json:"menuMounts"`
	Overrides []MenuOverride      `json:"menuOverrides"`
	Groups    []MenuResolvedGroup `json:"menuGroups"`
}

func inferMenuMountSourceType(capabilityID string) MenuMountSourceType {
	switch {
	case strings.HasPrefix(capabilityID, "plugin."):
		return MenuMountSourcePlugin
	case strings.HasPrefix(capabilityID, "crd."):
		return MenuMountSourceCRD
	default:
		return MenuMountSourceBuiltin
	}
}

func groupOrderMap(groups []MenuBlueprintGroup) map[string]int {
	orders := make(map[string]int, len(groups))
	for _, group := range groups {
		orders[group.Key] = group.Order
	}
	return orders
}

func sortResolvedEntries(entries []MenuResolvedEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Pinned != entries[j].Pinned {
			return entries[i].Pinned
		}
		if entries[i].SortOrder != entries[j].SortOrder {
			return entries[i].SortOrder < entries[j].SortOrder
		}
		if entries[i].Order != entries[j].Order {
			return entries[i].Order < entries[j].Order
		}
		return entries[i].EntryKey < entries[j].EntryKey
	})
}
