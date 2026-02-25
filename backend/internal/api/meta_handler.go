package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type MetaHandler struct{}

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
		"cluster":       cluster,
		"resourceTypes": []string{"Deployment", "Service", "ConfigMap"},
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
		"menus": []map[string]any{
			{
				"id":     "workloads",
				"group":  "system",
				"title":  "Workloads",
				"source": "system",
				"order":  10,
			},
			{
				"id":     "crd-dynamic",
				"group":  "dynamic",
				"title":  "Custom Resources",
				"source": "dynamic",
				"order":  50,
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
