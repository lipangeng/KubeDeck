import type { ActiveCluster } from './types';

export function createClusterContext(clusterId = 'default'): ActiveCluster {
  return {
    id: clusterId,
    status: 'ready',
    lastStableId: clusterId,
  };
}

export function requestClusterSwitch(cluster: ActiveCluster): ActiveCluster {
  return {
    ...cluster,
    status: 'switching',
  };
}

export function completeClusterSwitch(
  cluster: ActiveCluster,
  nextClusterId: string,
): ActiveCluster {
  return {
    id: nextClusterId,
    status: 'ready',
    lastStableId: nextClusterId,
  };
}

export function failClusterSwitch(cluster: ActiveCluster): ActiveCluster {
  if (cluster.lastStableId) {
    return {
      ...cluster,
      id: cluster.lastStableId,
      status: 'ready',
    };
  }

  return {
    ...cluster,
    status: 'failed',
  };
}
