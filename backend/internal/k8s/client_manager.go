package k8s

import (
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientManager manages Kubernetes client connections
type ClientManager struct {
	mu        sync.RWMutex
	clients   map[string]*kubernetes.Clientset
	configs   map[string]*rest.Config
	current   string
}

// ClusterClient holds client and info for a cluster
type ClusterClient struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	Name      string
	Server    string
}

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*kubernetes.Clientset),
		configs: make(map[string]*rest.Config),
	}
}

// AddCluster adds a cluster configuration
func (m *ClientManager) AddCluster(name, server, token string, insecure bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config := &rest.Config{
		Host: server,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: insecure,
		},
		BearerToken: token,
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	m.clients[name] = clientset
	m.configs[name] = config

	if m.current == "" {
		m.current = name
	}

	return nil
}

// GetClient gets a client for a cluster
func (m *ClientManager) GetClient(name string) (*kubernetes.Clientset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[name]
	if !ok {
		return nil, fmt.Errorf("cluster %s not found", name)
	}

	return client, nil
}

// GetCurrentClient gets the current cluster client
func (m *ClientManager) GetCurrentClient() (*kubernetes.Clientset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.current == "" {
		return nil, fmt.Errorf("no cluster selected")
	}

	return m.clients[m.current], nil
}

// SetCurrent sets the current cluster
func (m *ClientManager) SetCurrent(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[name]; !ok {
		return fmt.Errorf("cluster %s not found", name)
	}

	m.current = name
	return nil
}

// LoadFromKubeconfig loads clusters from kubeconfig
func (m *ClientManager) LoadFromKubeconfig(kubeconfig string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.clients["default"] = clientset
	m.configs["default"] = config
	m.current = "default"
	m.mu.Unlock()

	return nil
}

// Global client manager instance
var globalClientManager *ClientManager
var globalManagerOnce sync.Once

// GlobalClientManager returns the global client manager
func GlobalClientManager() *ClientManager {
	globalManagerOnce.Do(func() {
		globalClientManager = NewClientManager()
	})
	return globalClientManager
}
