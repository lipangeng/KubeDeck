package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DeploymentActionRequest represents a deployment action request
type DeploymentActionRequest struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Action    string `json:"action"`
	Replicas  *int32 `json:"replicas,omitempty"`
}

// DeploymentHandler handles deployment-related requests
func (h *KernelHandler) DeploymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeploymentActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	// Get K8s client
	config, err := rest.InClusterConfig()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"success": false,
			"error":   "K8s not configured: " + err.Error(),
		})
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var result map[string]interface{}
	switch req.Action {
	case "scale":
		result, err = scaleDeployment(r.Context(), clientset, req)
	case "restart":
		result, err = restartDeployment(r.Context(), clientset, req)
	case "pause":
		result, err = pauseDeployment(r.Context(), clientset, req)
	case "resume":
		result, err = resumeDeployment(r.Context(), clientset, req)
	default:
		http.Error(w, "unknown action: "+req.Action, http.StatusBadRequest)
		return
	}

	if err != nil {
		writeJSON(w, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, result)
}

func scaleDeployment(ctx context.Context, clientset *kubernetes.Clientset, req DeploymentActionRequest) (map[string]interface{}, error) {
	if req.Replicas == nil {
		return nil, fmt.Errorf("replicas required for scale action")
	}

	deployment, err := clientset.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	deployment.Spec.Replicas = req.Replicas

	updated, err := clientset.AppsV1().Deployments(req.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Scaled deployment %s to %d replicas", req.Name, *req.Replicas),
		"replicas": map[string]int32{
			"desired": *req.Replicas,
			"current": updated.Status.Replicas,
			"ready":   updated.Status.ReadyReplicas,
		},
	}, nil
}

func restartDeployment(ctx context.Context, clientset *kubernetes.Clientset, req DeploymentActionRequest) (map[string]interface{}, error) {
	deployment, err := clientset.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = clientset.AppsV1().Deployments(req.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Restarted deployment %s", req.Name),
	}, nil
}

func pauseDeployment(ctx context.Context, clientset *kubernetes.Clientset, req DeploymentActionRequest) (map[string]interface{}, error) {
	deployment, err := clientset.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	deployment.Spec.Paused = true

	_, err = clientset.AppsV1().Deployments(req.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Paused deployment %s", req.Name),
	}, nil
}

func resumeDeployment(ctx context.Context, clientset *kubernetes.Clientset, req DeploymentActionRequest) (map[string]interface{}, error) {
	deployment, err := clientset.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	deployment.Spec.Paused = false

	_, err = clientset.AppsV1().Deployments(req.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Resumed deployment %s", req.Name),
	}, nil
}

// GetDeploymentHandler gets deployment details
func (h *KernelHandler) GetDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}
	name := r.URL.Query().Get("name")

	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	// Get K8s client
	config, err := rest.InClusterConfig()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"error": "K8s not configured",
		})
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	type DeploymentDetail struct {
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		Replicas          int32             `json:"replicas"`
		ReadyReplicas     int32             `json:"ready_replicas"`
		AvailableReplicas int32             `json:"available_replicas"`
		UpdatedReplicas   int32             `json:"updated_replicas"`
		Strategy          string            `json:"strategy"`
		Images            []string          `json:"images"`
		Labels            map[string]string `json:"labels,omitempty"`
		CreatedAt         string            `json:"created_at"`
	}

	images := make([]string, 0, len(deployment.Spec.Template.Spec.Containers))
	for _, c := range deployment.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}

	strategy := "RollingUpdate"
	if deployment.Spec.Strategy.Type == appsv1.RecreateDeploymentStrategyType {
		strategy = "Recreate"
	}

	detail := DeploymentDetail{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Replicas:          deployment.Status.Replicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		Strategy:          strategy,
		Images:            images,
		Labels:            deployment.Labels,
		CreatedAt:         deployment.CreationTimestamp.Time.Format(time.RFC3339),
	}

	writeJSON(w, detail)
}

// ListDeploymentsHandler lists deployments in a namespace
func (h *KernelHandler) ListDeploymentsHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	// Get K8s client
	config, err := rest.InClusterConfig()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"deployments": []interface{}{},
		})
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"deployments": []interface{}{},
			"error":       err.Error(),
		})
		return
	}

	type DeploymentSummary struct {
		Name          string `json:"name"`
		ReadyReplicas int32  `json:"ready_replicas"`
		TotalReplicas int32  `json:"total_replicas"`
		Available     bool   `json:"available"`
	}

	list := make([]DeploymentSummary, 0, len(deployments.Items))
	for _, dep := range deployments.Items {
		available := dep.Status.AvailableReplicas > 0
		list = append(list, DeploymentSummary{
			Name:          dep.Name,
			ReadyReplicas: dep.Status.ReadyReplicas,
			TotalReplicas: dep.Status.Replicas,
			Available:     available,
		})
	}

	writeJSON(w, map[string]interface{}{
		"deployments": list,
		"total":       len(deployments.Items),
	})
}

// BatchDeploymentActionRequest represents a batch deployment action request
type BatchDeploymentActionRequest struct {
	Namespace string   `json:"namespace"`
	Names     []string `json:"names"`
	Action    string   `json:"action"`
	Replicas  *int32   `json:"replicas,omitempty"`
}

// BatchDeploymentHandler handles batch deployment operations
func (h *KernelHandler) BatchDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BatchDeploymentActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"success": false,
			"error":   "K8s not configured",
		})
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	results := make([]map[string]interface{}, 0, len(req.Names))
	
	for _, name := range req.Names {
		actionReq := DeploymentActionRequest{
			Namespace: req.Namespace,
			Name:      name,
			Action:    req.Action,
			Replicas:  req.Replicas,
		}

		var result map[string]interface{}
		var err error

		switch req.Action {
		case "scale":
			result, err = scaleDeployment(r.Context(), clientset, actionReq)
		case "restart":
			result, err = restartDeployment(r.Context(), clientset, actionReq)
		case "pause":
			result, err = pauseDeployment(r.Context(), clientset, actionReq)
		case "resume":
			result, err = resumeDeployment(r.Context(), clientset, actionReq)
		}

		if err != nil {
			results = append(results, map[string]interface{}{
				"name":    name,
				"success": false,
				"error":   err.Error(),
			})
		} else {
			results = append(results, result)
		}
	}

	writeJSON(w, map[string]interface{}{
		"success": true,
		"results": results,
		"total":   len(req.Names),
	})
}
