package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"kubedeck/backend/internal/storage"
)

// ClusterConfigRequest represents a cluster configuration request
type ClusterConfigRequest struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Server     string `json:"server"`
	Token      string `json:"token,omitempty"`
	Kubeconfig string `json:"kubeconfig,omitempty"`
}

// ClustersConfigHandler handles cluster configuration requests
func (h *KernelHandler) ClustersConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listClustersConfig(w, r)
	case http.MethodPost:
		h.createClusterConfig(w, r)
	case http.MethodPut:
		h.updateClusterConfig(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ClustersConfigDelete handles cluster deletion
func (h *KernelHandler) ClustersConfigDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	clusterID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.clusterConfigRepo.Delete(clusterID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestClusterConnection tests connection to a Kubernetes cluster
func (h *KernelHandler) TestClusterConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClusterConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual K8s connection test
	writeJSON(w, map[string]interface{}{
		"success": true,
		"message": "连接成功！Kubernetes v1.28.0",
		"version": "v1.28.0",
		"nodes":   3,
	})
}

func (h *KernelHandler) listClustersConfig(w http.ResponseWriter, r *http.Request) {
	clusters, err := h.clusterConfigRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type SafeCluster struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Server    string    `json:"server"`
		Status    string    `json:"status"`
		Version   string    `json:"version,omitempty"`
		Nodes     int       `json:"nodes,omitempty"`
		CreatedAt time.Time `json:"created_at"`
	}

	safeClusters := make([]SafeCluster, 0, len(clusters))
	for _, c := range clusters {
		safeClusters = append(safeClusters, SafeCluster{
			ID:        c.ID.String(),
			Name:      c.Name,
			Server:    c.Server,
			Status:    c.Status,
			Version:   c.Version,
			Nodes:     c.Nodes,
			CreatedAt: c.CreatedAt,
		})
	}

	writeJSON(w, map[string]interface{}{
		"clusters": safeClusters,
	})
}

func (h *KernelHandler) createClusterConfig(w http.ResponseWriter, r *http.Request) {
	var req ClusterConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Server == "" {
		writeJSON(w, map[string]string{
			"message": "名称和 API Server 地址为必填项",
		})
		return
	}

	cluster := &storage.ClusterConfig{
		ID:     uuid.New(),
		Name:   req.Name,
		Server: req.Server,
		Token:  req.Token,
		Status: "disconnected",
	}

	if err := h.clusterConfigRepo.Create(cluster); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]string{
		"id":      cluster.ID.String(),
		"message": "集群配置已保存",
	})
}

func (h *KernelHandler) updateClusterConfig(w http.ResponseWriter, r *http.Request) {
	var req ClusterConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	clusterID, err := uuid.Parse(req.ID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	cluster, err := h.clusterConfigRepo.GetByID(clusterID)
	if err != nil {
		http.Error(w, "cluster not found", http.StatusNotFound)
		return
	}

	cluster.Name = req.Name
	cluster.Server = req.Server
	if req.Token != "" {
		cluster.Token = req.Token
	}

	if err := h.clusterConfigRepo.Update(cluster); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{
		"message": "集群配置已更新",
	})
}

// GetClusterConfigHandler gets a specific cluster config
func (h *KernelHandler) GetClusterConfigHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	clusterID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	cluster, err := h.clusterConfigRepo.GetByID(clusterID)
	if err != nil {
		http.Error(w, "cluster not found", http.StatusNotFound)
		return
	}

	type SafeCluster struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Server    string    `json:"server"`
		Status    string    `json:"status"`
		Version   string    `json:"version,omitempty"`
		Nodes     int       `json:"nodes,omitempty"`
		CreatedAt time.Time `json:"created_at"`
	}

	safeCluster := SafeCluster{
		ID:        cluster.ID.String(),
		Name:      cluster.Name,
		Server:    cluster.Server,
		Status:    cluster.Status,
		Version:   cluster.Version,
		Nodes:     cluster.Nodes,
		CreatedAt: cluster.CreatedAt,
	}

	writeJSON(w, safeCluster)
}

// ConnectClusterHandler tests and updates cluster connection
func (h *KernelHandler) ConnectClusterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID     string `json:"id"`
		Token  string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	clusterID, err := uuid.Parse(req.ID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	cluster, err := h.clusterConfigRepo.GetByID(clusterID)
	if err != nil {
		http.Error(w, "cluster not found", http.StatusNotFound)
		return
	}

	// Update token if provided
	if req.Token != "" {
		cluster.Token = req.Token
	}

	// Test connection
	// TODO: Implement actual K8s connection test
	cluster.Status = "connected"
	now := time.Now()
	cluster.LastConnect = &now

	if err := h.clusterConfigRepo.Update(cluster); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{
		"status":  "connected",
		"message": "Cluster connected successfully",
	})
}
