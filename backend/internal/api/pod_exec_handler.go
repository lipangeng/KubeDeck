package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"kubedeck/backend/internal/k8s"
)

// ExecMessage represents a message from/to the exec session
type ExecMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

// PodExecRequest represents a pod exec request
type PodExecRequest struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Container string   `json:"container,omitempty"`
	Command   []string `json:"command"`
	TTY       bool     `json:"tty,omitempty"`
}

// PodExecHandler handles pod exec requests
func (h *KernelHandler) PodExecHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	clientManager := k8s.GlobalClientManager()
	_, err = clientManager.GetCurrentClient()
	if err != nil {
		sendError(conn, "No cluster configured")
		return
	}

	var req PodExecRequest
	if err := conn.ReadJSON(&req); err != nil {
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	if len(req.Command) == 0 {
		req.Command = []string{"/bin/sh"}
	}

	// Send initial message
	initMsg := ExecMessage{
		Type: "output",
		Data: "Exec session established. Ready for commands.\n",
	}
	conn.WriteJSON(initMsg)

	// Handle WebSocket messages
	if err := handleExecSession(r.Context(), conn, req); err != nil {
		sendError(conn, "Exec failed: "+err.Error())
	}
}

func handleExecSession(ctx context.Context, conn *websocket.Conn, req PodExecRequest) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			var msg ExecMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			if msg.Type == "close" {
				return nil
			}

			// Echo back for mock
			response := ExecMessage{
				Type: "output",
				Data: "Command received: " + msg.Data + "\n",
			}
			if err := conn.WriteJSON(response); err != nil {
				return err
			}
		}
	}
}
