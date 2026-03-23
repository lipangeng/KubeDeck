package k8s

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WorkloadItem represents a Kubernetes workload
type WorkloadItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	Replicas  string `json:"replicas,omitempty"`
	Image     string `json:"image,omitempty"`
	UpdatedAt string `json:"updated_at"`
}

// WorkloadsProvider provides workloads from Kubernetes cluster
type WorkloadsProvider struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewWorkloadsProvider creates a new workloads provider
func NewWorkloadsProvider(clientset *kubernetes.Clientset, namespace string) *WorkloadsProvider {
	return &WorkloadsProvider{
		clientset: clientset,
		namespace: namespace,
	}
}

// ListWorkloads lists all workloads in the cluster
func (p *WorkloadsProvider) ListWorkloads(ctx context.Context, namespace string) ([]WorkloadItem, error) {
	var workloads []WorkloadItem

	if namespace == "" {
		namespace = p.namespace
	}

	// List Deployments
	deployments, err := p.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, dep := range deployments.Items {
		workloads = append(workloads, deploymentToWorkloadItem(dep))
	}

	// List Pods
	pods, err := p.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
		workloads = append(workloads, podToWorkloadItem(pod))
	}

	// List StatefulSets
	statefulSets, err := p.clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}

	for _, ss := range statefulSets.Items {
		workloads = append(workloads, statefulSetToWorkloadItem(ss))
	}

	// List DaemonSets
	daemonSets, err := p.clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets: %w", err)
	}

	for _, ds := range daemonSets.Items {
		workloads = append(workloads, daemonSetToWorkloadItem(ds))
	}

	return workloads, nil
}

// GetDeployment gets a deployment by name
func (p *WorkloadsProvider) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return p.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetPod gets a pod by name
func (p *WorkloadsProvider) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return p.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ScaleDeployment scales a deployment
func (p *WorkloadsProvider) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	scale, err := p.clientset.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = p.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	return err
}

// RestartDeployment restarts a deployment by updating annotations
func (p *WorkloadsProvider) RestartDeployment(ctx context.Context, namespace, name string) error {
	deployment, err := p.GetDeployment(ctx, namespace, name)
	if err != nil {
		return err
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = p.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	return err
}

// GetPodLogs gets logs from a pod
func (p *WorkloadsProvider) GetPodLogs(ctx context.Context, namespace, name, container string, tailLines int64) (string, error) {
	req := p.clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	buf := make([]byte, 2000)
	n, err := stream.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func deploymentToWorkloadItem(dep appsv1.Deployment) WorkloadItem {
	status := "Running"
	health := "Healthy"

	if dep.Status.Replicas != dep.Status.ReadyReplicas {
		status = "Progressing"
		health = "Warning"
	}

	if dep.Status.UnavailableReplicas > 0 {
		status = "Unavailable"
		health = "Error"
	}

	image := ""
	if len(dep.Spec.Template.Spec.Containers) > 0 {
		image = dep.Spec.Template.Spec.Containers[0].Image
	}

	return WorkloadItem{
		ID:        fmt.Sprintf("deployment-%s-%s", dep.Namespace, dep.Name),
		Name:      dep.Name,
		Kind:      "Deployment",
		Namespace: dep.Namespace,
		Status:    status,
		Health:    health,
		Replicas:  fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, dep.Status.Replicas),
		Image:     image,
		UpdatedAt: dep.Status.Conditions[0].LastUpdateTime.Time.Format(time.RFC3339),
	}
}

func podToWorkloadItem(pod corev1.Pod) WorkloadItem {
	status := string(pod.Status.Phase)
	health := "Healthy"

	if pod.Status.Phase == corev1.PodPending {
		health = "Warning"
	} else if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
		health = "Error"
	}

	image := ""
	if len(pod.Spec.Containers) > 0 {
		image = pod.Spec.Containers[0].Image
	}

	return WorkloadItem{
		ID:        fmt.Sprintf("pod-%s-%s", pod.Namespace, pod.Name),
		Name:      pod.Name,
		Kind:      "Pod",
		Namespace: pod.Namespace,
		Status:    status,
		Health:    health,
		Image:     image,
		UpdatedAt: pod.CreationTimestamp.Time.Format(time.RFC3339),
	}
}

func statefulSetToWorkloadItem(ss appsv1.StatefulSet) WorkloadItem {
	status := "Running"
	health := "Healthy"

	if ss.Status.Replicas != ss.Status.ReadyReplicas {
		status = "Progressing"
		health = "Warning"
	}

	image := ""
	if len(ss.Spec.Template.Spec.Containers) > 0 {
		image = ss.Spec.Template.Spec.Containers[0].Image
	}

	return WorkloadItem{
		ID:        fmt.Sprintf("statefulset-%s-%s", ss.Namespace, ss.Name),
		Name:      ss.Name,
		Kind:      "StatefulSet",
		Namespace: ss.Namespace,
		Status:    status,
		Health:    health,
		Replicas:  fmt.Sprintf("%d/%d", ss.Status.ReadyReplicas, ss.Status.Replicas),
		Image:     image,
		UpdatedAt: ss.CreationTimestamp.Time.Format(time.RFC3339),
	}
}

func daemonSetToWorkloadItem(ds appsv1.DaemonSet) WorkloadItem {
	status := "Running"
	health := "Healthy"

	if ds.Status.NumberReady != ds.Status.DesiredNumberScheduled {
		status = "Progressing"
		health = "Warning"
	}

	image := ""
	if len(ds.Spec.Template.Spec.Containers) > 0 {
		image = ds.Spec.Template.Spec.Containers[0].Image
	}

	return WorkloadItem{
		ID:        fmt.Sprintf("daemonset-%s-%s", ds.Namespace, ds.Name),
		Name:      ds.Name,
		Kind:      "DaemonSet",
		Namespace: ds.Namespace,
		Status:    status,
		Health:    health,
		Replicas:  fmt.Sprintf("%d/%d", ds.Status.NumberReady, ds.Status.DesiredNumberScheduled),
		Image:     image,
		UpdatedAt: ds.CreationTimestamp.Time.Format(time.RFC3339),
	}
}
