import type { PageContribution } from '../contracts/pageContribution';
import { HomepagePage } from './pages/HomepagePage';
import { WorkloadsPage } from './pages/WorkloadsPage';
import { ClustersPage } from './pages/ClustersPage';
import { ClustersConfigPage } from './pages/ClustersConfigPage';
import { PodPage } from './pages/PodPage';
import { DeploymentPage } from './pages/DeploymentPage';
import { ServicePage } from './pages/ServicePage';
import { ConfigMapPage } from './pages/ConfigMapPage';
import { IngressPage } from './pages/IngressPage';
import { RolesPage } from './pages/RolesPage';
import { EventsPage } from './pages/EventsPage';

export function registerBuiltInPages(): PageContribution[] {
  return [
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.homepage',
        contributionId: 'page.homepage',
      },
      workflowDomainId: 'homepage',
      route: '/',
      entryKey: 'homepage',
      title: { key: 'homepage.title', fallback: 'Homepage' },
      description: {
        key: 'homepage.description',
        fallback: 'Enter the core built-in workflow domains from the kernel shell.',
      },
      component: HomepagePage,
      order: 10,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.clusters',
        contributionId: 'page.clusters',
      },
      workflowDomainId: 'clusters',
      route: '/clusters',
      entryKey: 'clusters',
      title: { key: 'clusters.title', fallback: 'Clusters' },
      description: {
        key: 'clusters.description',
        fallback: 'Manage your Kubernetes clusters from a unified control plane.',
      },
      component: ClustersPage,
      order: 15,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.clusters-config',
        contributionId: 'page.clusters-config',
      },
      workflowDomainId: 'clusters-config',
      route: '/clusters/config',
      entryKey: 'clusters-config',
      title: { key: 'clusters.config', fallback: 'Cluster Configuration' },
      description: {
        key: 'clusters.config.description',
        fallback: 'Manage Kubernetes cluster connections.',
      },
      component: ClustersConfigPage,
      order: 16,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.workloads',
        contributionId: 'page.workloads',
      },
      workflowDomainId: 'workloads',
      route: '/workloads',
      entryKey: 'workloads',
      title: { key: 'workloads.title', fallback: 'Workloads' },
      description: {
        key: 'workloads.description',
        fallback: 'View the first built-in workload domain through the kernel registry.',
      },
      component: WorkloadsPage,
      order: 20,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.pods',
        contributionId: 'page.pod-detail',
      },
      workflowDomainId: 'pods',
      route: '/pods/:namespace/:name',
      entryKey: 'pod-detail',
      title: { key: 'pod.detail', fallback: 'Pod Details' },
      description: {
        key: 'pod.description',
        fallback: 'View pod details, logs, and terminal access.',
      },
      component: PodPage,
      order: 30,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.deployments',
        contributionId: 'page.deployment-detail',
      },
      workflowDomainId: 'deployments',
      route: '/deployments/:namespace/:name',
      entryKey: 'deployment-detail',
      title: { key: 'deployment.detail', fallback: 'Deployment Details' },
      description: {
        key: 'deployment.description',
        fallback: 'Manage deployment replicas, rollout, and rollback.',
      },
      component: DeploymentPage,
      order: 35,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.services',
        contributionId: 'page.service-detail',
      },
      workflowDomainId: 'services',
      route: '/services/:namespace/:name',
      entryKey: 'service-detail',
      title: { key: 'service.detail', fallback: 'Service Details' },
      description: {
        key: 'service.description',
        fallback: 'View service endpoints, ports, and configuration.',
      },
      component: ServicePage,
      order: 38,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.configmaps',
        contributionId: 'page.configmap-detail',
      },
      workflowDomainId: 'configmaps',
      route: '/configmaps/:namespace/:name',
      entryKey: 'configmap-detail',
      title: { key: 'configmap.detail', fallback: 'ConfigMap Details' },
      description: {
        key: 'configmap.description',
        fallback: 'Manage ConfigMap and Secret data.',
      },
      component: ConfigMapPage,
      order: 39,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.ingress',
        contributionId: 'page.ingress-detail',
      },
      workflowDomainId: 'ingress',
      route: '/ingress/:namespace/:name',
      entryKey: 'ingress-detail',
      title: { key: 'ingress.detail', fallback: 'Ingress Details' },
      description: {
        key: 'ingress.description',
        fallback: 'Manage Ingress routing rules and TLS configuration.',
      },
      component: IngressPage,
      order: 40,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.rbac',
        contributionId: 'page.roles',
      },
      workflowDomainId: 'rbac',
      route: '/roles',
      entryKey: 'roles',
      title: { key: 'rbac.roles', fallback: 'Roles' },
      description: {
        key: 'rbac.description',
        fallback: 'Manage RBAC roles and cross-cluster permissions.',
      },
      component: RolesPage,
      order: 50,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.events',
        contributionId: 'page.events',
      },
      workflowDomainId: 'events',
      route: '/events',
      entryKey: 'events',
      title: { key: 'events.title', fallback: 'Events' },
      description: {
        key: 'events.description',
        fallback: 'View Kubernetes cluster events.',
      },
      component: EventsPage,
      order: 55,
    },
  ];
}
