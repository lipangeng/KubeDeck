package api

import (
	"encoding/json"
	"net/http"

	"kubedeck/backend/internal/plugins/helm"
)

var helmPlugin *helm.HelmPlugin

func init() {
	helmPlugin = &helm.HelmPlugin{}
}

// HelmHandler handles Helm-related requests via plugin
func (h *KernelHandler) HelmHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	switch r.URL.Query().Get("action") {
	case "list":
		namespace := r.URL.Query().Get("namespace")
		if namespace == "" {
			namespace = "default"
		}
		
		result, err := helmPlugin.Execute(ctx, "releases.list", map[string]interface{}{
			"namespace": namespace,
		})
		writeHelmResult(w, result, err)
		
	case "install":
		var req helm.InstallParams
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		
		result, err := helmPlugin.Execute(ctx, "releases.install", map[string]interface{}{
			"name":      req.Name,
			"chart":     req.Chart,
			"namespace": req.Namespace,
			"values":    req.Values,
		})
		writeHelmResult(w, result, err)
		
	case "upgrade":
		var req helm.InstallParams
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		
		result, err := helmPlugin.Execute(ctx, "releases.upgrade", map[string]interface{}{
			"name":      req.Name,
			"chart":     req.Chart,
			"namespace": req.Namespace,
			"values":    req.Values,
		})
		writeHelmResult(w, result, err)
		
	case "uninstall":
		name := r.URL.Query().Get("name")
		namespace := r.URL.Query().Get("namespace")
		if namespace == "" {
			namespace = "default"
		}
		
		result, err := helmPlugin.Execute(ctx, "releases.uninstall", map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		})
		writeHelmResult(w, result, err)
		
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
	}
}

// HelmRepoHandler handles Helm repository operations via plugin
func (h *KernelHandler) HelmRepoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	switch r.URL.Query().Get("action") {
	case "list":
		result, err := helmPlugin.Execute(ctx, "repos.list", nil)
		writeHelmResult(w, result, err)
		
	case "add":
		var req helm.AddRepoParams
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		
		result, err := helmPlugin.Execute(ctx, "repos.add", map[string]interface{}{
			"name": req.Name,
			"url":  req.URL,
		})
		writeHelmResult(w, result, err)
		
	case "remove":
		name := r.URL.Query().Get("name")
		
		result, err := helmPlugin.Execute(ctx, "repos.remove", map[string]interface{}{
			"name": name,
		})
		writeHelmResult(w, result, err)
		
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
	}
}

// HelmChartsHandler lists available charts from repositories via plugin
func (h *KernelHandler) HelmChartsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		http.Error(w, "repo required", http.StatusBadRequest)
		return
	}
	
	result, err := helmPlugin.Execute(ctx, "charts.search", map[string]interface{}{
		"repo": repo,
	})
	writeHelmResult(w, result, err)
}

func writeHelmResult(w http.ResponseWriter, result interface{}, err error) {
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	writeJSON(w, result)
}
