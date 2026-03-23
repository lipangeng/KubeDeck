package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

var execUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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
	conn, err := execUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	config, err := rest.InClusterConfig()
	if err != nil {
		sendExecErrorMessage(conn, "K8s not configured: "+err.Error())
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		sendExecErrorMessage(conn, "Failed to create client: "+err.Error())
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

	if err := execIntoContainer(r.Context(), conn, config, clientset, req); err != nil {
		sendExecErrorMessage(conn, "Exec failed: "+err.Error())
	}
}

func execIntoContainer(ctx context.Context, conn *websocket.Conn, config *rest.Config, clientset *kubernetes.Clientset, req PodExecRequest) error {
	reqURL := clientset.CoreV1().RESTClient().
		Post().
		Namespace(req.Namespace).
		Resource("pods").
		Name(req.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   req.Command,
			Container: req.Container,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       req.TTY,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", reqURL.URL())
	if err != nil {
		return err
	}

	// Send initial message
	initMsg := ExecMessage{
		Type: "output",
		Data: "Exec session established. Connected to " + req.Name + ".\n",
	}
	conn.WriteJSON(initMsg)

	// Create handler for streaming
	handler := &wsStreamHandler{
		conn: conn,
		ctx:  ctx,
	}

	// Start exec session
	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  handler,
		Stdout: handler,
		Stderr: handler,
		Tty:    req.TTY,
	}); err != nil {
		return err
	}

	return nil
}

type wsStreamHandler struct {
	conn *websocket.Conn
	ctx  context.Context
}

func (h *wsStreamHandler) Read(p []byte) (int, error) {
	h.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	_, message, err := h.conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	var msg ExecMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		return 0, err
	}

	if msg.Type == "resize" {
		return 0, nil
	}

	if msg.Type == "close" {
		return 0, fmt.Errorf("session closed")
	}

	copy(p, []byte(msg.Data))
	return len(msg.Data), nil
}

func (h *wsStreamHandler) Write(p []byte) (int, error) {
	msg := ExecMessage{
		Type: "output",
		Data: string(p),
	}

	if err := h.conn.WriteJSON(msg); err != nil {
		return 0, err
	}

	return len(p), nil
}

func sendExecErrorMessage(conn *websocket.Conn, message string) {
	errMsg := ExecMessage{
		Type: "error",
		Data: message,
	}
	conn.WriteJSON(errMsg)
}
