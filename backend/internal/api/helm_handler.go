package api

import (
	"encoding/json"
	"net/http"
	"os/exec"
)

// HelmHandler handles Helm-related requests
func (h *KernelHandler) HelmHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Query().Get("action") {
	case "list":
		h.listHelmReleases(w, r)
	case "install":
		h.installHelmChart(w, r)
	case "upgrade":
		h.upgradeHelmRelease(w, r)
	case "uninstall":
		h.uninstallHelmRelease(w, r)
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
	}
}

// HelmRelease represents a Helm release
type HelmRelease struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   int    `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

func (h *KernelHandler) listHelmReleases(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	// Execute helm list command
	cmd := exec.Command("helm", "list", "-n", namespace, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		// If helm is not installed, return empty list
		writeJSON(w, map[string]interface{}{
			"releases": []HelmRelease{},
			"error": "Helm not installed or no releases found",
		})
		return
	}

	var releases []HelmRelease
	if err := json.Unmarshal(output, &releases); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"releases": releases,
	})
}

// InstallRequest represents a Helm install request
type InstallRequest struct {
	Name       string            `json:"name"`
	Chart      string            `json:"chart"`
	Namespace  string            `json:"namespace"`
	Values     map[string]string `json:"values,omitempty"`
}

func (h *KernelHandler) installHelmChart(w http.ResponseWriter, r *http.Request) {
	var req InstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Chart == "" {
		http.Error(w, "name and chart required", http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	// Build helm install command
	args := []string{"install", req.Name, req.Chart, "-n", req.Namespace}
	
	// Add values
	for key, value := range req.Values {
		args = append(args, "--set", key+"="+value)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, map[string]string{
			"success": "false",
			"error":   string(output),
		})
		return
	}

	writeJSON(w, map[string]string{
		"success": "true",
		"message": "Helm chart installed successfully",
	})
}

func (h *KernelHandler) upgradeHelmRelease(w http.ResponseWriter, r *http.Request) {
	var req InstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Chart == "" {
		http.Error(w, "name and chart required", http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	// Build helm upgrade command
	args := []string{"upgrade", req.Name, req.Chart, "-n", req.Namespace}
	
	for key, value := range req.Values {
		args = append(args, "--set", key+"="+value)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, map[string]string{
			"success": "false",
			"error":   string(output),
		})
		return
	}

	writeJSON(w, map[string]string{
		"success": "true",
		"message": "Helm release upgraded successfully",
	})
}

func (h *KernelHandler) uninstallHelmRelease(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	if namespace == "" {
		namespace = "default"
	}

	cmd := exec.Command("helm", "uninstall", name, "-n", namespace)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, map[string]string{
			"success": "false",
			"error":   string(output),
		})
		return
	}

	writeJSON(w, map[string]string{
		"success": "true",
		"message": "Helm release uninstalled successfully",
	})
}

// HelmRepoHandler handles Helm repository operations
func (h *KernelHandler) HelmRepoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Query().Get("action") {
	case "list":
		h.listHelmRepos(w, r)
	case "add":
		h.addHelmRepo(w, r)
	case "remove":
		h.removeHelmRepo(w, r)
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
	}
}

// HelmRepo represents a Helm repository
type HelmRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (h *KernelHandler) listHelmRepos(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("helm", "repo", "list", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"repos": []HelmRepo{},
		})
		return
	}

	var repos []HelmRepo
	if err := json.Unmarshal(output, &repos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"repos": repos,
	})
}

// AddRepoRequest represents a request to add a Helm repository
type AddRepoRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (h *KernelHandler) addHelmRepo(w http.ResponseWriter, r *http.Request) {
	var req AddRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.URL == "" {
		http.Error(w, "name and url required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("helm", "repo", "add", req.Name, req.URL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, map[string]string{
			"success": "false",
			"error":   string(output),
		})
		return
	}

	writeJSON(w, map[string]string{
		"success": "true",
		"message": "Helm repository added successfully",
	})
}

func (h *KernelHandler) removeHelmRepo(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("helm", "repo", "remove", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, map[string]string{
			"success": "false",
			"error":   string(output),
		})
		return
	}

	writeJSON(w, map[string]string{
		"success": "true",
		"message": "Helm repository removed successfully",
	})
}

// HelmChartsHandler lists available charts from repositories
func (h *KernelHandler) HelmChartsHandler(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		http.Error(w, "repo required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("helm", "search", "repo", repo, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"charts": []interface{}{},
		})
		return
	}

	var charts []interface{}
	if err := json.Unmarshal(output, &charts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"charts": charts,
	})
}
