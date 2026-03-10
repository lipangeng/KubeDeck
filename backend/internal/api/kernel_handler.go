package api

import (
	"encoding/json"
	"net/http"

	"kubedeck/backend/internal/core/builtins"
	"kubedeck/backend/internal/plugins"
)

type KernelHandler struct {
	registry *plugins.CapabilityRegistry
}

func NewKernelHandler() *KernelHandler {
	registry := plugins.NewCapabilityRegistry()
	_ = registry.Register(builtins.HomepageCapability{})
	_ = registry.Register(builtins.WorkloadsCapability{})
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

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
