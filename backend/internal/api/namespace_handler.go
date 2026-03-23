package api

import (
	"net/http"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubedeck/backend/internal/k8s"
)

// NamespaceHandler handles namespace-related requests
func (h *KernelHandler) NamespaceHandler(w http.ResponseWriter, r *http.Request) {
	clientManager := k8s.GlobalClientManager()
	clientset, err := clientManager.GetCurrentClient()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"namespaces": []string{"default"},
			"error": "No cluster configured",
		})
		return
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type NamespaceInfo struct {
		Name      string    `json:"name"`
		Status    string    `json:"status"`
		CreatedAt string    `json:"created_at"`
	}

	nsList := make([]NamespaceInfo, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		status := "Active"
		if ns.Status.Phase == "Terminating" {
			status = "Terminating"
		}

		nsList = append(nsList, NamespaceInfo{
			Name:      ns.Name,
			Status:    status,
			CreatedAt: ns.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}

	// Sort by name
	sort.Slice(nsList, func(i, j int) bool {
		return nsList[i].Name < nsList[j].Name
	})

	writeJSON(w, map[string]interface{}{
		"namespaces": nsList,
	})
}

// GetNamespaceHandler gets details for a specific namespace
func (h *KernelHandler) GetNamespaceHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	clientManager := k8s.GlobalClientManager()
	clientset, err := clientManager.GetCurrentClient()
	if err != nil {
		http.Error(w, "No cluster configured", http.StatusInternalServerError)
		return
	}

	ns, err := clientset.CoreV1().Namespaces().Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	type NamespaceDetail struct {
		Name        string            `json:"name"`
		Status      string            `json:"status"`
		Phase       string            `json:"phase"`
		CreatedAt   string            `json:"created_at"`
		Labels      map[string]string `json:"labels,omitempty"`
		Annotations map[string]string `json:"annotations,omitempty"`
	}

	detail := NamespaceDetail{
		Name:        ns.Name,
		Status:      string(ns.Status.Phase),
		Phase:       string(ns.Status.Phase),
		CreatedAt:   ns.CreationTimestamp.Time.Format(time.RFC3339),
		Labels:      ns.Labels,
		Annotations: ns.Annotations,
	}

	writeJSON(w, detail)
}
