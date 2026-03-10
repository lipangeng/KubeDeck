import { describe, expect, it } from 'vitest';
import {
  completeClusterSwitch,
  createClusterContext,
  failClusterSwitch,
  requestClusterSwitch,
} from './clusterContext';

describe('clusterContext', () => {
  it('creates a ready cluster context', () => {
    expect(createClusterContext('prod')).toEqual({
      id: 'prod',
      status: 'ready',
      lastStableId: 'prod',
    });
  });

  it('marks cluster as switching without changing active id', () => {
    const cluster = createClusterContext('prod');
    expect(requestClusterSwitch(cluster)).toEqual({
      id: 'prod',
      status: 'switching',
      lastStableId: 'prod',
    });
  });

  it('completes a cluster switch with a new stable id', () => {
    const cluster = requestClusterSwitch(createClusterContext('prod'));
    expect(completeClusterSwitch(cluster, 'staging')).toEqual({
      id: 'staging',
      status: 'ready',
      lastStableId: 'staging',
    });
  });

  it('falls back to the last stable cluster on switch failure', () => {
    const cluster = requestClusterSwitch(createClusterContext('prod'));
    expect(failClusterSwitch(cluster)).toEqual({
      id: 'prod',
      status: 'ready',
      lastStableId: 'prod',
    });
  });
});
