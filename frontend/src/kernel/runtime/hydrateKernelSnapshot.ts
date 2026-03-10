import type { ActionContribution } from '../contracts/actionContribution';
import type { MenuContribution } from '../contracts/menuContribution';
import type { PageContribution } from '../contracts/pageContribution';
import { createRemoteCapabilityPage } from './createRemoteCapabilityPage';
import type { KernelRegistrySnapshot } from './types';
import type { RemoteActionDescriptor, RemoteKernelMetadata, RemoteMenuDescriptor, RemotePageDescriptor } from './transport';

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
    placement: menu.Placement,
    order: menu.Order,
    visible: menu.Visible,
    route: menu.Route,
    title: toLocalizedText(menu.Title),
    description: menu.Description ? toLocalizedText(menu.Description) : undefined,
  }));
}

function hydrateActions(remoteActions: RemoteActionDescriptor[]): ActionContribution[] {
  return remoteActions.map((action) => ({
    identity: {
      source: 'builtin',
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
}

export function hydrateKernelSnapshot(
  localSnapshot: KernelRegistrySnapshot,
  remoteMetadata: RemoteKernelMetadata,
): KernelRegistrySnapshot {
  const pages = hydratePages(localSnapshot.pages, remoteMetadata.pages);
  return {
    pages: pages.length > 0 ? pages : localSnapshot.pages,
    menus: hydrateMenus(remoteMetadata.menus),
    actions: hydrateActions(remoteMetadata.actions),
    slots: localSnapshot.slots,
  };
}
