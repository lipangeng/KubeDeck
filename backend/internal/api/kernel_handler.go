package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"kubedeck/backend/internal/auth"
	"kubedeck/backend/internal/cache"
	"kubedeck/backend/internal/core/builtins"
	"kubedeck/backend/internal/encryption"
	"kubedeck/backend/internal/k8s"
	"kubedeck/backend/internal/plugins"
	"kubedeck/backend/internal/storage"
	"kubedeck/backend/pkg/sdk"
)

type KernelHandler struct {
	registry          *plugins.CapabilityRegistry
	menuRepo          storage.UserMenuRepo
	db                *storage.Database
	roleRepo          storage.RoleRepository
	roleBindingRepo   storage.RoleBindingRepository
	userRepo          storage.UserRepository
	auditLogRepo      storage.AuditLogRepository
	oauthManager      *auth.Manager
	k8sSyncer         *k8s.RBACSyncer
	cache             *cache.Cache
	clusterConfigRepo storage.ClusterConfigRepository
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
	_ = registry.Register(builtins.EventsCapability{})
	_ = registry.Register(builtins.HelmCapability{})
	_ = registry.Register(builtins.ProfileCapability{})
	if providers, err := plugins.LoadManifestProvidersFromDir(pluginRoot); err == nil {
		for _, provider := range providers {
			_ = registry.Register(provider)
		}
	}

	// Initialize database and repositories
	db, _ := storage.NewDatabase(storage.DatabaseConfig{})

	// Initialize OAuth2 manager
	oauthManager := auth.NewManager(generateJWTSecret())

	// Register default OAuth2 providers (can be configured via API)
	_ = oauthManager.RegisterProvider(&auth.OAuth2Config{
		Provider:     "github",
		ClientID:     os.Getenv("OAUTH_GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"user:email"},
	})

	_ = oauthManager.RegisterProvider(&auth.OAuth2Config{
		Provider:     "google",
		ClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
	})

	// Initialize K8s RBAC syncer
	k8sSyncer, _ := k8s.NewRBACSyncer(os.Getenv("K8S_NAMESPACE"))

	// Initialize cache
	cacheInstance := cache.Global()

	// Initialize cluster config repository
	clusterConfigRepo := storage.NewClusterConfigRepository(db.DB())

	return &KernelHandler{
		registry:          registry,
		menuRepo:          menuRepo,
		db:                db,
		roleRepo:          db.RoleRepository(),
		roleBindingRepo:   db.RoleBindingRepository(),
		userRepo:          db.UserRepository(),
		auditLogRepo:      db.AuditLogRepository(),
		oauthManager:      oauthManager,
		k8sSyncer:         k8sSyncer,
		cache:             cacheInstance,
		clusterConfigRepo: clusterConfigRepo,
	}
}

func generateJWTSecret() string {
	// Generate a random secret for JWT signing
	// In production, this should be loaded from environment/config
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateUUID() string {
	// Simple UUID v4 generation
	b := make([]byte, 16)
	rand.Read(b)
	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
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
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "github"
	}

	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}

	authURL, _, err := h.oauthManager.GenerateAuthURL(provider, redirect)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

// OAuth2Callback handles OAuth2 callback
func (h *KernelHandler) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	// Validate state
	stateData, err := h.oauthManager.ValidateState(state)
	if err != nil {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token and get user info
	token, userInfo, err := h.oauthManager.ExchangeCode(r.Context(), provider, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get or create user in database
	user, err := h.getOrCreateUser(userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	jwtToken, err := h.oauthManager.GenerateJWT(userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create session
	var refreshToken string
	if token.RefreshToken != "" {
		refreshToken = token.RefreshToken
	}
	session := h.oauthManager.CreateSession(user.ID.String(), refreshToken)

	// Set auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		MaxAge:   604800, // 7 days
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to original destination
	redirectURL := stateData.Redirect
	if redirectURL == "" {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// LoginLocal handles local username/password login
func (h *KernelHandler) LoginLocal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, map[string]string{
			"message": "用户名和密码不能为空",
		})
		return
	}

	// TODO: Implement actual user authentication against database
	// For now, accept 'admin' with any password for testing
	if req.Username == "admin" {
		// Create mock user info
		userInfo := &auth.UserInfo{
			ID:       "local:admin",
			Username: req.Username,
			Email:    "admin@kubedeck.io",
			Provider: "local",
			Verified: true,
		}

		// Generate JWT
		jwtToken, err := h.oauthManager.GenerateJWT(userInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create session
		session := h.oauthManager.CreateSession(userInfo.ID, "")

		// Set cookies
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    jwtToken,
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    session.ID,
			Path:     "/",
			MaxAge:   604800,
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})

		writeJSON(w, map[string]interface{}{
			"success": true,
			"user": map[string]string{
				"id":       userInfo.ID,
				"username": userInfo.Username,
				"email":    userInfo.Email,
			},
		})
		return
	}

	writeJSON(w, map[string]string{
		"message": "用户名或密码错误",
	})
}

// Logout handles user logout
func (h *KernelHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session from cookie
	sessionCookie, err := r.Cookie("session_id")
	if err == nil {
		h.oauthManager.DeleteSession(sessionCookie.Value)
	}

	// Clear auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

// GetCurrentUser returns the current authenticated user
func (h *KernelHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r)
	if !ok {
		// Not authenticated, return anonymous
		writeJSON(w, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	// Get user from database
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	writeJSON(w, map[string]interface{}{
		"authenticated": true,
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"provider":      user.OAuthProvider,
	})
}

// getOrCreateUser gets existing user or creates new one from OAuth2 info
func (h *KernelHandler) getOrCreateUser(userInfo *auth.UserInfo) (*storage.User, error) {
	// Try to find existing user by OAuth sub
	user, err := h.userRepo.GetByOAuth(userInfo.Provider, userInfo.ID)
	if err == nil {
		return user, nil
	}

	// Create new user
	userID := uuid.New()
	user = &storage.User{
		ID:            userID,
		Username:      userInfo.Username,
		Email:         userInfo.Email,
		OAuthProvider: userInfo.Provider,
		OAuthSub:      userInfo.ID,
	}

	if err := h.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// ConfigureOAuth2 allows configuring OAuth2 providers via API
func (h *KernelHandler) ConfigureOAuth2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg auth.OAuth2Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "invalid config", http.StatusBadRequest)
		return
	}

	// Encrypt client secret before storing
	if cfg.ClientSecret != "" {
		encryptor, err := encryption.NewEncryptor(os.Getenv("ENCRYPTION_KEY"))
		if err == nil {
			encrypted, err := encryptor.Encrypt(cfg.ClientSecret)
			if err == nil {
				cfg.ClientSecret = encrypted
			}
		}
	}

	if err := h.oauthManager.RegisterProvider(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{"status": "configured"})
}
