import type { ActionContribution } from '../contracts/actionContribution';
import type { MenuContribution } from '../contracts/menuContribution';
import type { PageContribution } from '../contracts/pageContribution';
import type { SlotContribution } from '../contracts/slotContribution';
import { createRemoteCapabilityPage } from './createRemoteCapabilityPage';
import { createRemoteCapabilitySlot } from './createRemoteCapabilitySlot';
import type { KernelNavigationGroup } from './menu/types';
import type { KernelRegistrySnapshot } from './types';
import type {
  RemoteActionDescriptor,
  RemoteKernelMetadata,
  RemoteMenuDescriptor,
  RemoteMenuGroup,
  RemotePageDescriptor,
  RemoteResourcePageExtensionDescriptor,
  RemoteSlotDescriptor,
} from './transport';
import type { ResourcePageExtension } from '../resource-pages/types';

function toLocalizedText(text: { Key: string; Fallback: string; Description?: string }) {
  return {
    key: text.Key,
    fallback: text.Fallback,
    description: text.Description,
  };
}

function hydratePages(
  localPages: PageContribution[],
  remotePages: RemotePageDescriptor[],
): PageContribution[] {
  const localByEntryKey = new Map(localPages.map((page) => [page.entryKey, page]));
  const hydratedPages = remotePages
    .map((page) => {
      const local = localByEntryKey.get(page.EntryKey);
      if (!local) {
        const title = toLocalizedText(page.Title);
        const description = page.Description ? toLocalizedText(page.Description) : undefined;
        const remotePage: PageContribution = {
          identity: {
            source: 'plugin',
            capabilityId: `server.${page.WorkflowDomainID}`,
            contributionId: page.ID,
          },
          workflowDomainId: page.WorkflowDomainID,
          route: page.Route,
          entryKey: page.EntryKey,
          title,
          description,
          component: createRemoteCapabilityPage(title, description),
          visible: true,
        };
        return remotePage;
      }
      return {
        ...local,
        workflowDomainId: page.WorkflowDomainID,
        route: page.Route,
        entryKey: page.EntryKey,
        title: toLocalizedText(page.Title),
        description: page.Description ? toLocalizedText(page.Description) : local.description,
      };
    })
    .filter((page): page is NonNullable<typeof page> => page !== null);

  return hydratedPages;
}

function hydrateMenus(remoteMenus: RemoteMenuDescriptor[]): MenuContribution[] {
  return remoteMenus.map((menu) => ({
    identity: {
      source: 'builtin',
      capabilityId: `server.${menu.WorkflowDomainID}`,
      contributionId: menu.ID,
    },
    workflowDomainId: menu.WorkflowDomainID,
    entryKey: menu.EntryKey,
    groupKey: menu.GroupKey,
    placement: menu.Placement,
    availability: menu.Availability,
    isFallback: menu.IsFallback,
    order: menu.Order,
    visible: menu.Visible,
    route: menu.Route,
    title: toLocalizedText(menu.Title),
    description: menu.Description ? toLocalizedText(menu.Description) : undefined,
  }));
}

export function hydrateRemoteMenuGroups(remoteGroups: RemoteMenuGroup[]): KernelNavigationGroup[] {
  return remoteGroups.map((group) => ({
    key: group.key,
    order: group.order,
    title: toLocalizedText(group.title),
    entries: group.entries.map((entry) => ({
      identity: {
        source: entry.SourceType === 'plugin' ? 'plugin' : 'builtin',
        capabilityId: entry.CapabilityID,
        contributionId: entry.ID,
      },
      workflowDomainId: entry.WorkflowDomainID,
      entryKey: entry.EntryKey,
      groupKey: entry.GroupKey,
      placement: entry.Placement,
      availability: entry.Availability,
      isFallback: entry.IsFallback,
      order: entry.Order,
      visible: entry.Visible,
      route: entry.Route,
      title: toLocalizedText(entry.Title),
      description: entry.Description ? toLocalizedText(entry.Description) : undefined,
    })),
  }));
}

function hydrateActions(
  localActions: ActionContribution[],
  remoteActions: RemoteActionDescriptor[],
): ActionContribution[] {
  const hydratedRemote = remoteActions.map((action) => ({
    identity: {
      source: 'builtin' as const,
      capabilityId: `server.${action.WorkflowDomainID}`,
      contributionId: action.ID,
    },
    workflowDomainId: action.WorkflowDomainID,
    actionId: action.ID,
    surface: action.Surface,
    visible: action.Visible,
    permissionHint: action.PermissionHint,
    title: toLocalizedText(action.Title),
    description: action.Description ? toLocalizedText(action.Description) : undefined,
  }));

  const merged = [...localActions];
  const existingIds = new Set(localActions.map((action) => action.actionId));
  for (const action of hydratedRemote) {
    if (!existingIds.has(action.actionId)) {
      merged.push(action);
    }
  }
  return merged;
}

function hydrateSlots(
  localSlots: SlotContribution[],
  remoteSlots: RemoteSlotDescriptor[],
): SlotContribution[] {
  const localBySlotId = new Map(localSlots.map((slot) => [slot.slotId, slot]));
  const hydratedRemote = remoteSlots.map((slot) => {
    const title = slot.Title ? toLocalizedText(slot.Title) : undefined;
    const local = localBySlotId.get(slot.SlotID);
    if (local) {
      return {
        ...local,
        workflowDomainId: slot.WorkflowDomainID,
        slotId: slot.SlotID,
        placement: slot.Placement,
        visible: slot.Visible,
        title: title ?? local.title,
      };
    }

    const remoteSlot: SlotContribution = {
      identity: {
        source: 'plugin',
        capabilityId: `server.${slot.WorkflowDomainID}`,
        contributionId: slot.ID,
      },
      workflowDomainId: slot.WorkflowDomainID,
      slotId: slot.SlotID,
      placement: slot.Placement,
      visible: slot.Visible,
      title,
      component: createRemoteCapabilitySlot(title),
    };
    return remoteSlot;
  });

  const merged = [...localSlots];
  const existingSlotIds = new Set(localSlots.map((slot) => slot.slotId));
  for (const slot of hydratedRemote) {
    if (!existingSlotIds.has(slot.slotId)) {
      merged.push(slot);
    }
  }
  return merged;
}

function resourcePageExtensionKey(extension: {
  kind: string;
  capabilityType?: string;
  targetTabId?: string;
  tabId?: string;
  placement?: string;
  actionId?: string;
}) {
  return `${extension.kind}:${extension.capabilityType ?? 'tab'}:${extension.targetTabId ?? ''}:${extension.tabId ?? ''}:${extension.placement ?? ''}:${extension.actionId ?? ''}`;
}

function getResourcePageExtensionIdentity(extension: ResourcePageExtension) {
  if (
    extension.capabilityType === undefined ||
    extension.capabilityType === 'tab' ||
    extension.capabilityType === 'tab-replace'
  ) {
    return {
      kind: extension.kind,
      capabilityType: extension.capabilityType ?? 'tab',
      targetTabId: extension.targetTabId,
      tabId: extension.tabId,
      placement: undefined,
    };
  }

  if (extension.capabilityType === 'slot') {
    return {
      kind: extension.kind,
      capabilityType: extension.capabilityType,
      targetTabId: undefined,
      tabId: undefined,
      placement: extension.placement,
      actionId: undefined,
    };
  }

  if (extension.capabilityType === 'action') {
    return {
      kind: extension.kind,
      capabilityType: extension.capabilityType,
      targetTabId: undefined,
      tabId: undefined,
      placement: undefined,
      actionId: extension.actionId,
    };
  }

  return {
    kind: extension.kind,
    capabilityType: extension.capabilityType,
    targetTabId: undefined,
    tabId: undefined,
    placement: undefined,
    actionId: undefined,
  };
}

function hydrateResourcePageExtensions(
  localExtensions: ResourcePageExtension[],
  remoteExtensions: RemoteResourcePageExtensionDescriptor[],
): ResourcePageExtension[] {
  const hydratedRemote = remoteExtensions.map((extension) => {
    if (extension.CapabilityType === 'page-takeover') {
      return {
        kind: extension.Kind,
        capabilityType: 'page-takeover' as const,
        priority: extension.Priority,
        origin: 'remote' as const,
        renderPage: (options: { resource?: { name: string } }) =>
          options.resource?.name
            ? `${extension.ContentFallback} for ${options.resource.name}`
            : extension.ContentFallback,
      };
    }

    if (extension.CapabilityType === 'action') {
      return {
        kind: extension.Kind,
        capabilityType: 'action' as const,
        actionId: extension.ActionID ?? '',
        priority: extension.Priority,
        origin: 'remote' as const,
        createAction: (options: { resource?: { name: string } }) => ({
          id: `${extension.ActionID ?? 'resource-action'}:${options.resource?.name ?? 'resource'}`,
          title: extension.Title.Fallback,
          actionId: extension.ActionID ?? '',
        }),
      };
    }

    if (extension.CapabilityType === 'slot') {
      return {
        kind: extension.Kind,
        capabilityType: 'slot' as const,
        placement: 'summary' as const,
        priority: extension.Priority,
        origin: 'remote' as const,
        renderSlot: (options: { resource?: { name: string } }) =>
          options.resource?.name
            ? `${extension.ContentFallback ?? extension.Title.Fallback} for ${options.resource.name}`
            : (extension.ContentFallback ?? extension.Title.Fallback),
      };
    }

    return {
      kind: extension.Kind,
      capabilityType: extension.CapabilityType,
      targetTabId: extension.TargetTabID,
      tabId: extension.TabID ?? `${extension.Kind.toLowerCase()}.remote-tab`,
      priority: extension.Priority,
      origin: 'remote' as const,
      createTab: (options: { resource?: { name: string } }) => ({
        id: extension.TabID ?? `${extension.Kind.toLowerCase()}.remote-tab`,
        title: extension.Title.Fallback,
        capabilityType: extension.CapabilityType,
        content: options.resource?.name
          ? `${extension.ContentFallback} for ${options.resource.name}`
          : extension.ContentFallback,
      }),
    };
  });

  const merged = [...localExtensions];
  const existingKeys = new Set(
    localExtensions.map((extension) => resourcePageExtensionKey(getResourcePageExtensionIdentity(extension))),
  );

  for (const extension of hydratedRemote) {
    if (extension.capabilityType === 'page-takeover') {
      merged.push(extension);
      continue;
    }
    const key = resourcePageExtensionKey(extension);
    if (!existingKeys.has(key)) {
      merged.push(extension);
      existingKeys.add(key);
    }
  }

  return merged;
}

export function hydrateKernelSnapshot(
  localSnapshot: KernelRegistrySnapshot,
  remoteMetadata: RemoteKernelMetadata,
): KernelRegistrySnapshot {
  const pages = hydratePages(localSnapshot.pages, remoteMetadata.pages);
  return {
    pages: pages.length > 0 ? pages : localSnapshot.pages,
    menus: hydrateMenus(remoteMetadata.menus),
    menuGroups:
      remoteMetadata.menuGroups && remoteMetadata.menuGroups.length > 0
        ? hydrateRemoteMenuGroups(remoteMetadata.menuGroups)
        : localSnapshot.menuGroups,
    actions: hydrateActions(localSnapshot.actions, remoteMetadata.actions),
    slots: hydrateSlots(localSnapshot.slots, remoteMetadata.slots),
    resourcePageExtensions: hydrateResourcePageExtensions(
      localSnapshot.resourcePageExtensions,
      remoteMetadata.resourcePageExtensions ?? [],
    ),
  };
}
