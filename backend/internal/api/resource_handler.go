package api

import "net/http"

type ResourceHandler struct{}

type workloadItemDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	UpdatedAt string `json:"updatedAt"`
}

func NewResourceHandler() *ResourceHandler {
	return &ResourceHandler{}
}

func (h *ResourceHandler) Workloads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	namespace := r.URL.Query().Get("namespace")
	items := workloadsForCluster(cluster)
	if namespace != "" && namespace != "all" {
		filtered := make([]workloadItemDTO, 0, len(items))
		for _, item := range items {
			if item.Namespace == namespace {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	if err := writeJSON(w, http.StatusOK, map[string]any{
		"cluster":   cluster,
		"namespace": namespace,
		"items":     items,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *ResourceHandler) Apply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	if err := writeJSON(w, http.StatusOK, map[string]any{
		"status":  "accepted",
			"message": "resource apply stub",
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
	}
}

func workloadsForCluster(cluster string) []workloadItemDTO {
	switch cluster {
	case "prod":
		return []workloadItemDTO{
			{
				ID:        "prod-default-api",
				Name:      "api",
				Kind:      "Deployment",
				Namespace: "default",
				Status:    "Running",
				Health:    "Healthy",
				UpdatedAt: "2026-03-10T08:15:00Z",
			},
			{
				ID:        "prod-default-web",
				Name:      "web",
				Kind:      "Deployment",
				Namespace: "default",
				Status:    "Pending",
				Health:    "Warning",
				UpdatedAt: "2026-03-10T08:18:00Z",
			},
			{
				ID:        "prod-default-web-service",
				Name:      "web",
				Kind:      "Service",
				Namespace: "default",
				Status:    "Active",
				Health:    "Healthy",
				UpdatedAt: "2026-03-10T08:20:00Z",
			},
		}
	case "staging":
		return []workloadItemDTO{
			{
				ID:        "staging-default-api",
				Name:      "api",
				Kind:      "Deployment",
				Namespace: "default",
				Status:    "Running",
				Health:    "Healthy",
				UpdatedAt: "2026-03-10T07:45:00Z",
			},
		}
	default:
		return []workloadItemDTO{
			{
				ID:        "dev-default-api",
				Name:      "api",
				Kind:      "Deployment",
				Namespace: "default",
				Status:    "Running",
				Health:    "Healthy",
				UpdatedAt: "2026-03-10T06:30:00Z",
			},
			{
				ID:        "dev-default-api-service",
				Name:      "api",
				Kind:      "Service",
				Namespace: "default",
				Status:    "Active",
				Health:    "Healthy",
				UpdatedAt: "2026-03-10T06:32:00Z",
			},
			{
				ID:        "dev-team-a-worker",
				Name:      "worker",
				Kind:      "Deployment",
				Namespace: "team-a",
				Status:    "CrashLoopBackOff",
				Health:    "Critical",
				UpdatedAt: "2026-03-10T06:40:00Z",
			},
		}
	}
}
