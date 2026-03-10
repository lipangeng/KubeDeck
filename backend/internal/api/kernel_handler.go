package api

import (
	"encoding/json"
	"net/http"
	"os"

	"kubedeck/backend/internal/core/builtins"
	"kubedeck/backend/internal/plugins"
	"kubedeck/backend/pkg/sdk"
)

type KernelHandler struct {
	registry *plugins.CapabilityRegistry
}

func NewKernelHandler() *KernelHandler {
	return NewKernelHandlerWithPluginRoot(os.Getenv("KUBEDECK_PLUGIN_DIR"))
}

func NewKernelHandlerWithPluginRoot(pluginRoot string) *KernelHandler {
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
	return &KernelHandler{registry: registry}
}

func (h *KernelHandler) Menus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, plugins.ComposeMenus(h.registry.Descriptors()))
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

func (h *KernelHandler) Snapshot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, plugins.ComposeKernelSnapshot(h.registry.Descriptors()))
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
