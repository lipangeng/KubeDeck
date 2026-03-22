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
	registry        *plugins.CapabilityRegistry
	menuRepo        storage.UserMenuRepo
	db              *storage.Database
	roleRepo        storage.RoleRepository
	roleBindingRepo storage.RoleBindingRepository
	userRepo        storage.UserRepository
	auditLogRepo    storage.AuditLogRepository
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
	_ = registry.Register(builtins.ClustersCapability{})
	_ = registry.Register(builtins.WorkloadsCapability{})
	_ = registry.Register(builtins.WorkloadsInsightsCapability{})
	_ = registry.Register(builtins.OperationsCapability{})
	if providers, err := plugins.LoadManifestProvidersFromDir(pluginRoot); err == nil {
		for _, provider := range providers {
			_ = registry.Register(provider)
		}
	}

	// Initialize database and repositories
	db, _ := storage.NewDatabase(storage.DatabaseConfig{})

	return &KernelHandler{
		registry:        registry,
		menuRepo:        menuRepo,
		db:              db,
		roleRepo:        db.RoleRepository(),
		roleBindingRepo: db.RoleBindingRepository(),
		userRepo:        db.UserRepository(),
		auditLogRepo:    db.AuditLogRepository(),
	}
}

func (h *KernelHandler) Menus(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	globalOverrides, clusterOverrides := h.menuOverridesForCluster(r.URL.Query().Get("cluster"), scope)
	writeJSON(w, plugins.ComposeScopedKernelSnapshotWithOverrides(
		h.registry.Descriptors(),
		scope,
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
	scope := r.URL.Query().Get("scope")
	globalOverrides, clusterOverrides := h.menuOverridesForCluster(r.URL.Query().Get("cluster"), scope)
	writeJSON(w, plugins.ComposeScopedKernelSnapshotWithOverrides(
		h.registry.Descriptors(),
		scope,
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

func (h *KernelHandler) Clusters(w http.ResponseWriter, _ *http.Request) {
	items := plugins.ResolveClusters(h.registry.Providers())
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

func (h *KernelHandler) menuOverridesForCluster(cluster string, scope string) ([]plugins.MenuOverride, []plugins.MenuOverride) {
	if h.menuRepo == nil {
		return nil, nil
	}
	return filterMenuOverridesForScope(h.menuRepo.GetGlobalOverrides(defaultMenuUserID), scope),
		filterMenuOverridesForScope(h.menuRepo.GetClusterOverrides(defaultMenuUserID, cluster), scope)
}

func filterMenuOverridesForScope(overrides []plugins.MenuOverride, scope string) []plugins.MenuOverride {
	if len(overrides) == 0 {
		return nil
	}
	filtered := make([]plugins.MenuOverride, 0, len(overrides))
	for _, override := range overrides {
		switch scope {
		case "system":
			if override.Scope == plugins.MenuOverrideScopeSystem {
				filtered = append(filtered, override)
			}
		case "cluster":
			if override.Scope == plugins.MenuOverrideScopeCluster || override.Scope == plugins.MenuOverrideScopeClusterMenu {
				filtered = append(filtered, override)
			}
		default:
			if override.Scope == plugins.MenuOverrideScopeGlobal ||
				override.Scope == plugins.MenuOverrideScopeWorkGlobal ||
				override.Scope == plugins.MenuOverrideScopeCluster ||
				override.Scope == plugins.MenuOverrideScopeWorkCluster {
				filtered = append(filtered, override)
			}
		}
	}
	return filtered
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

// ListUsers returns all users
func (h *KernelHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, users)
}

// Login handles OAuth2 login redirect
func (h *KernelHandler) Login(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement OAuth2 login
	http.Redirect(w, r, "/", http.StatusFound)
}

// OAuth2Callback handles OAuth2 callback
func (h *KernelHandler) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement OAuth2 callback
	http.Redirect(w, r, "/", http.StatusFound)
}

// Logout handles user logout
func (h *KernelHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

// GetCurrentUser returns the current authenticated user
func (h *KernelHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract user from auth context
	writeJSON(w, map[string]interface{}{
		"id":       "system",
		"username": "admin",
		"email":    "admin@kubedeck.io",
	})
}
