package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type MetaHandler struct{}

type resourceTypeDTO struct {
	ID               string `json:"id"`
	Group            string `json:"group"`
	Version          string `json:"version"`
	Kind             string `json:"kind"`
	Plural           string `json:"plural"`
	Namespaced       bool   `json:"namespaced"`
	PreferredVersion string `json:"preferredVersion"`
	Source           string `json:"source"`
}

type menuItemDTO struct {
	ID         string `json:"id"`
	Group      string `json:"group"`
	Title      string `json:"title"`
	TargetType string `json:"targetType"`
	TargetRef  string `json:"targetRef"`
	Source     string `json:"source"`
	Order      int    `json:"order"`
	Visible    bool   `json:"visible"`
}

func NewMetaHandler() *MetaHandler {
	return &MetaHandler{}
}

func (h *MetaHandler) Registry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	if err := writeJSON(w, http.StatusOK, map[string]any{
		"cluster": cluster,
		"resourceTypes": []resourceTypeDTO{
			{
				ID:               "apps.v1.deployments",
				Group:            "apps",
				Version:          "v1",
				Kind:             "Deployment",
				Plural:           "deployments",
				Namespaced:       true,
				PreferredVersion: "v1",
				Source:           "system",
			},
			{
				ID:               "v1.services",
				Group:            "",
				Version:          "v1",
				Kind:             "Service",
				Plural:           "services",
				Namespaced:       true,
				PreferredVersion: "v1",
				Source:           "system",
			},
		},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *MetaHandler) Clusters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	if err := writeJSON(w, http.StatusOK, map[string]any{
		"clusters": []string{"dev", "staging", "prod"},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *MetaHandler) Menus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	if err := writeJSON(w, http.StatusOK, map[string]any{
		"cluster": cluster,
		"menus": []menuItemDTO{
			{
				ID:         "workloads",
				Group:      "system",
				Title:      "Workloads",
				TargetType: "page",
				TargetRef:  "/workloads",
				Source:     "system",
				Order:      10,
				Visible:    true,
			},
			{
				ID:         "favorites",
				Group:      "user",
				Title:      "Favorites",
				TargetType: "page",
				TargetRef:  "/favorites",
				Source:     "user",
				Order:      20,
				Visible:    true,
			},
			{
				ID:         "crd-dynamic",
				Group:      "dynamic",
				Title:      "Custom Resources",
				TargetType: "resource",
				TargetRef:  "/resources/custom",
				Source:     "dynamic",
				Order:      50,
				Visible:    true,
			},
		},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	if err := writeJSON(w, status, map[string]string{"error": message}); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("{\"error\":\"internal server error\"}\n"))
	}
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}
