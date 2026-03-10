package plugins

import (
	"sort"

	"kubedeck/backend/pkg/sdk"
)

type resolvedEntryState struct {
	entry  MenuResolvedEntry
	hidden bool
}

func composeBlueprintEntries(
	blueprint MenuBlueprint,
	mounts []MenuMount,
) ([]resolvedEntryState, map[string]struct{}) {
	mountByKey := make(map[string]MenuMount, len(mounts))
	for _, mount := range mounts {
		mountByKey[mount.EntryKey] = mount
	}

	usedMounts := make(map[string]struct{}, len(blueprint.Entries))
	resolved := make([]resolvedEntryState, 0, len(blueprint.Entries))
	for _, item := range blueprint.Entries {
		mount, ok := mountByKey[item.EntryKey]
		if ok {
			usedMounts[item.EntryKey] = struct{}{}
			resolved = append(resolved, resolvedEntryState{
				entry: MenuResolvedEntry{
					ID:               mount.ID,
					CapabilityID:     mount.CapabilityID,
					SourceType:       mount.SourceType,
					WorkflowDomainID: mount.WorkflowDomainID,
					EntryKey:         item.EntryKey,
					GroupKey:         item.DefaultGroupKey,
					Route:            mount.Route,
					Placement:        mount.Placement,
					Availability:     mount.Availability,
					IsFallback:       mount.IsFallback || item.IsFallback,
					Order:            item.Order,
					Visible:          mount.Visible,
					Title:            mount.Title,
					Description:      mount.Description,
					Mounted:          true,
					Configured:       true,
				},
			})
			continue
		}

		resolved = append(resolved, resolvedEntryState{
			entry: MenuResolvedEntry{
				ID:               "configured." + item.EntryKey,
				CapabilityID:     "configured." + item.EntryKey,
				SourceType:       item.SourceType,
				WorkflowDomainID: item.WorkflowDomainID,
				EntryKey:         item.EntryKey,
				GroupKey:         item.DefaultGroupKey,
				Route:            item.Route,
				Placement:        item.Placement,
				Availability:     sdk.MenuAvailabilityDisabledUnavailable,
				IsFallback:       item.IsFallback,
				Order:            item.Order,
				Visible:          true,
				Title:            item.Title,
				Mounted:          false,
				Configured:       true,
			},
		})
	}

	return resolved, usedMounts
}

func appendUnconfiguredMounts(
	resolved []resolvedEntryState,
	mounts []MenuMount,
	usedMounts map[string]struct{},
) []resolvedEntryState {
	for _, mount := range mounts {
		if _, used := usedMounts[mount.EntryKey]; used {
			continue
		}
		resolved = append(resolved, resolvedEntryState{
			entry: MenuResolvedEntry{
				ID:               mount.ID,
				CapabilityID:     mount.CapabilityID,
				SourceType:       mount.SourceType,
				WorkflowDomainID: mount.WorkflowDomainID,
				EntryKey:         mount.EntryKey,
				GroupKey:         mount.GroupKey,
				Route:            mount.Route,
				Placement:        mount.Placement,
				Availability:     mount.Availability,
				IsFallback:       mount.IsFallback,
				Order:            mount.Order,
				Visible:          mount.Visible,
				Title:            mount.Title,
				Description:      mount.Description,
				Mounted:          true,
				Configured:       false,
			},
		})
	}
	return resolved
}

func applyMenuOverrides(entries []resolvedEntryState, overrides []MenuOverride) []resolvedEntryState {
	for index := range entries {
		for _, override := range overrides {
			for _, hiddenEntryKey := range override.HiddenEntryKeys {
				if entries[index].entry.EntryKey == hiddenEntryKey {
					entries[index].hidden = true
					entries[index].entry.Availability = sdk.MenuAvailabilityHidden
				}
			}
			if groupKey, ok := override.MoveEntryKeys[entries[index].entry.EntryKey]; ok && groupKey != "" {
				entries[index].entry.GroupKey = groupKey
			}
			for _, pinnedEntryKey := range override.PinEntryKeys {
				if entries[index].entry.EntryKey == pinnedEntryKey {
					entries[index].entry.Pinned = true
				}
			}
		}
	}
	return entries
}

func buildMenuGroups(
	blueprint MenuBlueprint,
	entries []resolvedEntryState,
) []MenuResolvedGroup {
	groupDefinitions := make(map[string]MenuResolvedGroup, len(blueprint.Groups))
	groupOrders := groupOrderMap(blueprint.Groups)
	for _, group := range blueprint.Groups {
		groupDefinitions[group.Key] = MenuResolvedGroup{
			Key:   group.Key,
			Order: group.Order,
			Title: group.Title,
		}
	}

	for _, state := range entries {
		if state.hidden || !state.entry.Visible || state.entry.Availability == sdk.MenuAvailabilityHidden {
			continue
		}
		group := groupDefinitions[state.entry.GroupKey]
		if group.Key == "" {
			group = MenuResolvedGroup{
				Key:   state.entry.GroupKey,
				Order: 1000,
				Title: sdk.TextRef{Key: "menu.group." + state.entry.GroupKey, Fallback: state.entry.GroupKey},
			}
		}
		group.Entries = append(group.Entries, state.entry)
		groupDefinitions[state.entry.GroupKey] = group
		if _, ok := groupOrders[state.entry.GroupKey]; !ok {
			groupOrders[state.entry.GroupKey] = group.Order
		}
	}

	groups := make([]MenuResolvedGroup, 0, len(groupDefinitions))
	for _, group := range groupDefinitions {
		sortResolvedEntries(group.Entries)
		groups = append(groups, group)
	}

	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Order != groups[j].Order {
			return groups[i].Order < groups[j].Order
		}
		return groups[i].Key < groups[j].Key
	})

	return groups
}
