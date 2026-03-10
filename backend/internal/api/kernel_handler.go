package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"kubedeck/backend/internal/core/builtins"
	"kubedeck/backend/internal/plugins"
	"kubedeck/backend/internal/storage"
	"kubedeck/backend/pkg/sdk"
)

type KernelHandler struct {
	registry *plugins.CapabilityRegistry
	menuRepo storage.UserMenuRepo
}

const defaultMenuUserID = "default-user"

func NewKernelHandler() *KernelHandler {
	store, _ := storage.NewStore("", "")
	return NewKernelHandlerWithDependencies(
		resolvePluginRoot(os.Getenv("KUBEDECK_PLUGIN_DIR")),
		store.UserMenus(),
	)
}

func NewKernelHandlerWithPluginRoot(pluginRoot string) *KernelHandler {
	store, _ := storage.NewStore("", "")
	return NewKernelHandlerWithDependencies(pluginRoot, store.UserMenus())
}

func NewKernelHandlerWithDependencies(
	pluginRoot string,
	menuRepo storage.UserMenuRepo,
) *KernelHandler {
	registry := plugins.NewCapabilityRegistry()
	_ = registry.Register(builtins.HomepageCapability{})
	_ = registry.Register(builtins.WorkloadsCapability{})
	_ = registry.Register(builtins.WorkloadsInsightsCapability{})
	_ = registry.Register(builtins.OperationsCapability{})
	if providers, err := plugins.LoadManifestProvidersFromDir(pluginRoot); err == nil {
		for _, provider := range providers {
			_ = registry.Register(provider)
		}
	}
	return &KernelHandler{
		registry: registry,
		menuRepo: menuRepo,
	}
}

func (h *KernelHandler) Menus(w http.ResponseWriter, r *http.Request) {
	globalOverrides, clusterOverrides := h.menuOverridesForCluster(r.URL.Query().Get("cluster"))
	writeJSON(w, plugins.ComposeKernelSnapshotWithOverrides(
		h.registry.Descriptors(),
		globalOverrides,
		clusterOverrides,
	).Menus)
}

func (h *KernelHandler) Pages(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, plugins.ComposePages(h.registry.Descriptors()))
}

func (h *KernelHandler) Actions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, plugins.ComposeActions(h.registry.Descriptors()))
}

func (h *KernelHandler) Slots(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, plugins.ComposeSlots(h.registry.Descriptors()))
}

func (h *KernelHandler) Snapshot(w http.ResponseWriter, r *http.Request) {
	globalOverrides, clusterOverrides := h.menuOverridesForCluster(r.URL.Query().Get("cluster"))
	writeJSON(w, plugins.ComposeKernelSnapshotWithOverrides(
		h.registry.Descriptors(),
		globalOverrides,
		clusterOverrides,
	))
}

type menuPreferencesPayload struct {
	GlobalOverrides  []plugins.MenuOverride `json:"globalOverrides"`
	ClusterOverrides []plugins.MenuOverride `json:"clusterOverrides"`
}

func (h *KernelHandler) MenuPreferences(w http.ResponseWriter, r *http.Request) {
	cluster := r.URL.Query().Get("cluster")
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, menuPreferencesPayload{
			GlobalOverrides:  h.menuRepo.GetGlobalOverrides(defaultMenuUserID),
			ClusterOverrides: h.menuRepo.GetClusterOverrides(defaultMenuUserID, cluster),
		})
	case http.MethodPut:
		var payload menuPreferencesPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid menu preference payload", http.StatusBadRequest)
			return
		}
		if payload.GlobalOverrides != nil {
			if err := h.menuRepo.SaveGlobalOverrides(defaultMenuUserID, payload.GlobalOverrides); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if cluster != "" && payload.ClusterOverrides != nil {
			if err := h.menuRepo.SaveClusterOverrides(defaultMenuUserID, cluster, payload.ClusterOverrides); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *KernelHandler) Workloads(w http.ResponseWriter, r *http.Request) {
	workflowDomainID := r.URL.Query().Get("workflowDomainId")
	if workflowDomainID == "" {
		workflowDomainID = "workloads"
	}
	cluster := r.URL.Query().Get("cluster")
	items := plugins.ResolveWorkloads(h.registry.Providers(), workflowDomainID, cluster)
	writeJSON(w, items)
}

func (h *KernelHandler) ExecuteAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request sdk.ActionExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid action request", http.StatusBadRequest)
		return
	}

	result, err := plugins.ExecuteAction(h.registry.Providers(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, result)
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *KernelHandler) menuOverridesForCluster(cluster string) ([]plugins.MenuOverride, []plugins.MenuOverride) {
	if h.menuRepo == nil {
		return nil, nil
	}
	return h.menuRepo.GetGlobalOverrides(defaultMenuUserID), h.menuRepo.GetClusterOverrides(defaultMenuUserID, cluster)
}

func resolvePluginRoot(configured string) string {
	if configured != "" {
		return configured
	}

	candidates := make([]string, 0, 5)
	if _, currentFile, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "plugins"))
	}
	candidates = append(candidates,
		"plugins",
		filepath.Join("..", "plugins"),
		filepath.Join("..", "..", "plugins"),
		filepath.Join("..", "..", "..", "plugins"),
	)
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}
