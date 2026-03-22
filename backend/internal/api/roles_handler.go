package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"kubedeck/backend/internal/storage"
)

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Rules       []PolicyRuleInput `json:"rules"`
	IsBuiltin   bool              `json:"is_builtin"`
	ClusterIDs  []string          `json:"cluster_ids"`
	Namespaces  []string          `json:"namespaces"`
}

// PolicyRuleInput represents a Kubernetes policy rule
type PolicyRuleInput struct {
	APIGroups []string `json:"api_groups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

func (h *KernelHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roleRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, roles)
}

func (h *KernelHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	roleID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid role id", http.StatusBadRequest)
		return
	}

	role, err := h.roleRepo.GetByID(roleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, role)
}

func (h *KernelHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Convert rules to JSON
	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		http.Error(w, "invalid rules", http.StatusBadRequest)
		return
	}

	role := &storage.Role{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Rules:       string(rulesJSON),
		IsBuiltin:   req.IsBuiltin,
	}

	if err := h.roleRepo.Create(role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Create role binding when auth is integrated

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, role)
}

func (h *KernelHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	roleID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid role id", http.StatusBadRequest)
		return
	}

	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	role, err := h.roleRepo.GetByID(roleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		http.Error(w, "invalid rules", http.StatusBadRequest)
		return
	}

	role.Name = req.Name
	role.Description = req.Description
	role.Rules = string(rulesJSON)

	if err := h.roleRepo.Update(role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, role)
}

func (h *KernelHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	roleID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid role id", http.StatusBadRequest)
		return
	}

	if err := h.roleRepo.Delete(roleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *KernelHandler) ListRoleBindings(w http.ResponseWriter, r *http.Request) {
	bindings, err := h.roleBindingRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, bindings)
}

func (h *KernelHandler) CreateRoleBinding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoleID     string   `json:"role_id"`
		UserID     string   `json:"user_id"`
		ClusterIDs []string `json:"cluster_ids"`
		Namespaces []string `json:"namespaces"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		http.Error(w, "invalid role id", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	binding := &storage.RoleBinding{
		ID:         uuid.New(),
		RoleID:     roleID,
		UserID:     userID,
		ClusterIDs: marshalJSON(req.ClusterIDs),
		Namespaces: marshalJSON(req.Namespaces),
	}

	if err := h.roleBindingRepo.Create(binding); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, binding)
}

func marshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func getClaimsFromRequest(r *http.Request) (interface{}, bool) {
	// Implementation will be added when auth middleware is integrated
	return nil, false
}
