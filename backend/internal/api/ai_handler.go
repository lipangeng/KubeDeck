package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"kubedeck/backend/internal/ai"
	aiplugin "kubedeck/backend/internal/ai/plugin"
	"kubedeck/backend/internal/ai/plugin/kubernetes"
)

// AIHandler handles AI-related API requests
type AIHandler struct {
	aiManager     *ai.Manager
	pluginManager *aiplugin.Manager
	jwtManager    *interface{} // Will be implemented
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message string      `json:"message"`
	Context ChatContext `json:"context,omitempty"`
	Stream  bool        `json:"stream,omitempty"`
}

// ChatContext represents the chat context
type ChatContext struct {
	Cluster   string `json:"cluster,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Page      string `json:"page,omitempty"`
	User      string `json:"user,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Message    string                 `json:"message"`
	Action     string                 `json:"action,omitempty"`
	ActionData map[string]interface{} `json:"action_data,omitempty"`
}

// ApprovalRequest represents a command approval request
type ApprovalRequest struct {
	Command   string `json:"command"`
	Plugin    string `json:"plugin"`
	Reason    string `json:"reason"`
	Security  string `json:"security"`
	RequestID string `json:"request_id"`
}

// NewAIHandler creates a new AI handler
func NewAIHandler() *AIHandler {
	// Initialize AI manager
	aiManager := ai.NewManager()

	// Initialize plugin manager
	pluginManager := aiplugin.NewManager()

	// Register Kubernetes plugin
	k8sPlugin := k8splugin.NewKubernetesPlugin("", "", "default")
	pluginManager.Register(k8sPlugin)

	return &AIHandler{
		aiManager:     aiManager,
		pluginManager: pluginManager,
	}
}

// ConfigureAI configures an AI provider
func (h *AIHandler) ConfigureAI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config ai.ProviderConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "invalid config", http.StatusBadRequest)
		return
	}

	var provider ai.AIProvider
	switch config.Type {
	case "ollama":
		provider = ai.NewOllamaProvider(&config)
	case "openai":
		provider = ai.NewOpenAIProvider(&config)
	default:
		http.Error(w, "unsupported provider type", http.StatusBadRequest)
		return
	}

	h.aiManager.Register(config.Type, provider)

	writeJSON(w, map[string]string{"status": "configured"})
}

// Chat handles chat requests
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Process intent through plugin manager
	intentResult, err := h.pluginManager.ProcessIntent(r.Context(), req.Message)
	if err != nil {
		writeJSON(w, ChatResponse{
			Message: "I encountered an error processing your request.",
		})
		return
	}

	// If intent was matched, return the result
	if intentResult.Success {
		response := ChatResponse{
			Message:    intentResult.Message,
			Action:     intentResult.Action,
			ActionData: intentResult.ActionData,
		}

		if intentResult.RequiresApproval {
			response.Action = "approve"
		}

		writeJSON(w, response)
		return
	}

	// Fall back to AI chat
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a helpful Kubernetes assistant. You can help users manage their clusters, view resources, and understand Kubernetes concepts. Be concise and helpful.",
		},
		{
			Role:    "user",
			Content: req.Message,
		},
	}

	resp, err := h.aiManager.Chat(r.Context(), messages)
	if err != nil {
		writeJSON(w, ChatResponse{
			Message: "AI service is not configured. Please configure an AI provider in settings.",
		})
		return
	}

	writeJSON(w, ChatResponse{
		Message: resp.Content,
	})
}

// ChatStream handles streaming chat requests
func (h *AIHandler) ChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Set up SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a helpful Kubernetes assistant.",
		},
		{
			Role:    "user",
			Content: req.Message,
		},
	}

	stream, err := h.aiManager.Stream(r.Context(), messages)
	if err != nil {
		sendSSE(w, "error", "AI service is not configured")
		return
	}

	for chunk := range stream {
		if chunk.Done {
			sendSSE(w, "done", "")
			break
		}
		sendSSE(w, "message", chunk.Content)
		flusher.Flush()

		select {
		case <-r.Context().Done():
			return
		default:
		}
	}
}

// RequestApproval requests approval for a command
func (h *AIHandler) RequestApproval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Store approval request (in production, use database)
	// For now, just return success

	writeJSON(w, map[string]string{
		"status":     "pending",
		"request_id": req.RequestID,
	})
}

// ApproveCommand approves a command
func (h *AIHandler) ApproveCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RequestID string `json:"request_id"`
		Approved  bool   `json:"approved"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Process approval (in production, use database)
	status := "approved"
	if !req.Approved {
		status = "rejected"
	}

	writeJSON(w, map[string]string{"status": status})
}

// ExecuteCommand executes an approved command
func (h *AIHandler) ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Plugin    string   `json:"plugin"`
		Command   string   `json:"command"`
		Args      []string `json:"args"`
		RequestID string   `json:"request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := h.pluginManager.ExecuteCommand(ctx, req.Plugin, req.Command, req.Args)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, result)
}

// ListPlugins returns registered plugins
func (h *AIHandler) ListPlugins(w http.ResponseWriter, r *http.Request) {
	plugins := h.pluginManager.ListPlugins()
	writeJSON(w, map[string][]string{"plugins": plugins})
}

// ListCommands returns available commands
func (h *AIHandler) ListCommands(w http.ResponseWriter, r *http.Request) {
	commands := h.pluginManager.ListCommands()
	writeJSON(w, map[string]interface{}{"commands": commands})
}

func sendSSE(w http.ResponseWriter, event, data string) {
	w.Write([]byte("event: " + event + "\n"))
	w.Write([]byte("data: " + data + "\n\n"))
}
