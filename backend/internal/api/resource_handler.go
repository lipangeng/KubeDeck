package api

import "net/http"

type ResourceHandler struct{}

func NewResourceHandler() *ResourceHandler {
	return &ResourceHandler{}
}

func (h *ResourceHandler) Apply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	if err := writeJSON(w, http.StatusOK, map[string]any{
		"status":  "accepted",
		"message": "resource apply stub",
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
	}
}
