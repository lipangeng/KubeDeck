package api

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"kubedeck/backend/internal/core/audit"
	"gopkg.in/yaml.v3"
)

type ResourceHandler struct{}

func NewResourceHandler() *ResourceHandler {
	return &ResourceHandler{}
}

type applyResponse struct {
	Status           string        `json:"status"`
	Cluster          string        `json:"cluster"`
	DefaultNamespace string        `json:"defaultNamespace"`
	Total            int           `json:"total"`
	Succeeded        int           `json:"succeeded"`
	Failed           int           `json:"failed"`
	Results          []applyResult `json:"results"`
}

type applyResult struct {
	Index     int    `json:"index"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
}

type manifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
}

func (h *ResourceHandler) Apply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	defaultNS := r.URL.Query().Get("defaultNs")
	if defaultNS == "" {
		defaultNS = "default"
	}
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	documents := splitYAMLDocuments(body)
	results := make([]applyResult, 0, len(documents))
	succeeded := 0
	failed := 0

	for i, doc := range documents {
		result := applyResult{
			Index:  i + 1,
			Status: "succeeded",
		}

		var m manifest
		if err := yaml.Unmarshal(doc, &m); err != nil {
			result.Status = "failed"
			result.Reason = "invalid yaml document"
			failed++
			results = append(results, result)
			continue
		}

		result.Kind = m.Kind
		result.Name = m.Metadata.Name
		result.Namespace = m.Metadata.Namespace

		if m.Kind == "" {
			result.Status = "failed"
			result.Reason = "kind is required"
			failed++
			results = append(results, result)
			continue
		}
		if m.APIVersion == "" {
			result.Status = "failed"
			result.Reason = "apiVersion is required"
			failed++
			results = append(results, result)
			continue
		}
		if m.Metadata.Name == "" {
			result.Status = "failed"
			result.Reason = "metadata.name is required"
			failed++
			results = append(results, result)
			continue
		}

		if !isClusterScopedKind(m.Kind) && result.Namespace == "" {
			result.Namespace = defaultNS
		}

		if strings.Contains(result.Name, "fail") {
			result.Status = "failed"
			result.Reason = "simulated apply failure"
			failed++
			results = append(results, result)
			continue
		}

		succeeded++
		results = append(results, result)
	}

	status := "success"
	if failed > 0 && succeeded > 0 {
		status = "partial"
	} else if failed > 0 {
		status = "failed"
	}

	if err := writeJSON(w, http.StatusOK, applyResponse{
		Status:           status,
		Cluster:          cluster,
		DefaultNamespace: defaultNS,
		Total:            len(results),
		Succeeded:        succeeded,
		Failed:           failed,
		Results:          results,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   "default",
		Action:     "resource.apply",
		TargetType: "manifest",
		Result:     status,
		Metadata: map[string]string{
			"cluster": cluster,
			"total":   strconv.Itoa(len(results)),
		},
	})
}

func splitYAMLDocuments(raw []byte) [][]byte {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	out := make([][]byte, 0, 4)
	for {
		var node yaml.Node
		if err := decoder.Decode(&node); err != nil {
			if err == io.EOF {
				break
			}
			out = append(out, []byte(":::decode-error:::"))
			break
		}
		if len(node.Content) == 0 {
			continue
		}
		docBytes, err := yaml.Marshal(node.Content[0])
		if err != nil {
			out = append(out, []byte(":::marshal-error:::"))
			continue
		}
		if len(bytes.TrimSpace(docBytes)) == 0 {
			continue
		}
		out = append(out, docBytes)
	}
	return out
}

func isClusterScopedKind(kind string) bool {
	switch kind {
	case "Namespace", "Node", "PersistentVolume", "CustomResourceDefinition", "ClusterRole", "ClusterRoleBinding":
		return true
	default:
		return false
	}
}
