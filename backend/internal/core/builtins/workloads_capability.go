package builtins

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubedeck/backend/pkg/sdk"
)

type WorkloadsCapability struct{}

func (WorkloadsCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
	return sdk.CapabilityDescriptor{
		ID:      "core.workloads",
		Version: "v1",
		Pages: []sdk.PageDescriptor{
			{
				ID:               "page.workloads",
				WorkflowDomainID: "workloads",
				Route:            "/workloads",
				EntryKey:         "workloads",
				Title:            sdk.TextRef{Key: "workloads.title", Fallback: "Workloads"},
			},
		},
		Menus: []sdk.MenuDescriptor{
			{
				ID:               "menu.workloads",
				WorkflowDomainID: "workloads",
				EntryKey:         "workloads",
				GroupKey:         "core",
				Route:            "/workloads",
				Placement:        sdk.MenuPlacementPrimary,
				Availability:     sdk.MenuAvailabilityEnabled,
				Order:            20,
				Visible:          true,
				Title:            sdk.TextRef{Key: "workloads.title", Fallback: "Workloads"},
			},
		},
		Actions: []sdk.ActionDescriptor{
			{
				ID:               "create",
				WorkflowDomainID: "workloads",
				Surface:          sdk.ActionSurfaceDrawer,
				Visible:          true,
				Title:            sdk.TextRef{Key: "actions.create", Fallback: "Create"},
			},
			{
				ID:               "apply",
				WorkflowDomainID: "workloads",
				Surface:          sdk.ActionSurfaceDrawer,
				Visible:          true,
				Title:            sdk.TextRef{Key: "actions.apply", Fallback: "Apply"},
			},
		},
	}
}

func (WorkloadsCapability) WorkflowDomainID() string {
	return "workloads"
}

func (WorkloadsCapability) ListWorkloads(cluster string) []sdk.WorkloadItem {
	// Try to get real workloads from K8s
	workloads, err := getRealWorkloads(cluster)
	if err == nil && len(workloads) > 0 {
		return workloads
	}

	// Fallback to mock data
	return getMockWorkloads(cluster)
}

func getRealWorkloads(cluster string) ([]sdk.WorkloadItem, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var items []sdk.WorkloadItem

	// Get Deployments
	deployments, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, dep := range deployments.Items {
			items = append(items, deploymentToWorkload(dep))
		}
	}

	// Get Pods
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, pod := range pods.Items {
			items = append(items, podToWorkload(pod))
		}
	}

	return items, nil
}

func deploymentToWorkload(dep appsv1.Deployment) sdk.WorkloadItem {
	status := "Running"
	health := "Healthy"

	if dep.Status.ReadyReplicas != dep.Status.Replicas {
		status = "Progressing"
		health = "Warning"
	}

	if dep.Status.UnavailableReplicas > 0 {
		status = "Unavailable"
		health = "Error"
	}

	return sdk.WorkloadItem{
		ID:        fmt.Sprintf("deployment-%s-%s", dep.Namespace, dep.Name),
		Name:      dep.Name,
		Kind:      "Deployment",
		Namespace: dep.Namespace,
		Status:    status,
		Health:    health,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
}

func podToWorkload(pod corev1.Pod) sdk.WorkloadItem {
	status := string(pod.Status.Phase)
	health := "Healthy"

	if pod.Status.Phase == corev1.PodPending {
		health = "Warning"
	} else if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
		health = "Error"
	}

	return sdk.WorkloadItem{
		ID:        fmt.Sprintf("pod-%s-%s", pod.Namespace, pod.Name),
		Name:      pod.Name,
		Kind:      "Pod",
		Namespace: pod.Namespace,
		Status:    status,
		Health:    health,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
}

func getMockWorkloads(cluster string) []sdk.WorkloadItem {
	suffix := cluster
	if suffix == "" {
		suffix = "default"
	}

	return []sdk.WorkloadItem{
		{
			ID:        "workload-api-" + suffix,
			Name:      "api",
			Kind:      "Deployment",
			Namespace: "default",
			Status:    "Running",
			Health:    "Healthy",
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:        "workload-web-" + suffix,
			Name:      "web",
			Kind:      "Deployment",
			Namespace: "default",
			Status:    "Pending",
			Health:    "Warning",
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	}
}
