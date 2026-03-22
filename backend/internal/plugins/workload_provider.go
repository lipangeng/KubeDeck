package plugins

import "kubedeck/backend/pkg/sdk"

func ResolveWorkloads(
	providers []sdk.CapabilityProvider,
	workflowDomainID string,
	cluster string,
) []sdk.WorkloadItem {
	for _, provider := range providers {
		workloadProvider, ok := provider.(sdk.WorkloadProvider)
		if !ok {
			continue
		}
		if workloadProvider.WorkflowDomainID() != workflowDomainID {
			continue
		}
		return workloadProvider.ListWorkloads(cluster)
	}
	return nil
}

func ResolveClusters(providers []sdk.CapabilityProvider) []sdk.ClusterItem {
	for _, provider := range providers {
		clusterProvider, ok := provider.(sdk.ClusterProvider)
		if !ok {
			continue
		}
		return clusterProvider.ListClusters()
	}
	return nil
}
