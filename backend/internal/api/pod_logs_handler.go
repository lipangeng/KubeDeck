package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// PodLogRequest represents a pod log streaming request
type PodLogRequest struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Container string `json:"container,omitempty"`
	TailLines int64  `json:"tailLines,omitempty"`
	Follow    bool   `json:"follow,omitempty"`
}

// PodLogsHandler handles pod log streaming requests
func (h *KernelHandler) PodLogsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	config, err := rest.InClusterConfig()
	if err != nil {
		sendError(conn, "K8s not configured: "+err.Error())
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		sendError(conn, "Failed to create client: "+err.Error())
		return
	}

	var req PodLogRequest
	if err := conn.ReadJSON(&req); err != nil {
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	if err := streamPodLogs(r.Context(), conn, clientset, req); err != nil {
		sendError(conn, "Failed to stream logs: "+err.Error())
	}
}

func streamPodLogs(ctx context.Context, conn *websocket.Conn, clientset *kubernetes.Clientset, req PodLogRequest) error {
	tailLines := req.TailLines
	if tailLines == 0 {
		tailLines = 100
	}

	logReq := clientset.CoreV1().Pods(req.Namespace).GetLogs(req.Name, &corev1.PodLogOptions{
		Container:  req.Container,
		Follow:     req.Follow,
		TailLines:  &tailLines,
		Timestamps: true,
	})

	stream, err := logReq.Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	buf := make([]byte, 2000)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, err := stream.Read(buf)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}

			if n > 0 {
				message := map[string]string{
					"type": "log",
					"data": string(buf[:n]),
				}
				if err := conn.WriteJSON(message); err != nil {
					return err
				}
			}
		}
	}
}

func sendError(conn *websocket.Conn, message string) {
	errMsg := map[string]string{
		"type":  "error",
		"error": message,
	}
	conn.WriteJSON(errMsg)
}

// GetPodLogsHandler handles single pod log retrieval
func (h *KernelHandler) GetPodLogsHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}
	name := r.URL.Query().Get("name")
	container := r.URL.Query().Get("container")
	tailLines := int64(100)

	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		writeJSON(w, map[string]string{
			"error": "K8s not configured",
		})
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
		Container:  container,
		TailLines:  &tailLines,
		Timestamps: true,
	})

	stream, err := req.Stream(r.Context())
	if err != nil {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}
	defer stream.Close()

	buf := make([]byte, 10000)
	n, err := stream.Read(buf)
	if err != nil && err != io.EOF {
		writeJSON(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, map[string]string{
		"logs": string(buf[:n]),
	})
}

// GetPodHandler handles pod detail retrieval
func (h *KernelHandler) GetPodHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}
	name := r.URL.Query().Get("name")

	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		writeJSON(w, map[string]string{
			"error": "K8s not configured",
		})
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	type SafePod struct {
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Status    string            `json:"status"`
		Phase     string            `json:"phase"`
		IP        string            `json:"ip"`
		Node      string            `json:"node"`
		CreatedAt string            `json:"created_at"`
		Labels    map[string]string `json:"labels,omitempty"`
	}

	safePod := SafePod{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Phase:     string(pod.Status.Phase),
		IP:        pod.Status.PodIP,
		Node:      pod.Spec.NodeName,
		CreatedAt: pod.CreationTimestamp.Time.Format(time.RFC3339),
		Labels:    pod.Labels,
	}

	writeJSON(w, safePod)
}
