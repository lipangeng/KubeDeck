package k8splugin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	aiplugin "kubedeck/backend/internal/ai/plugin"
)

// KubernetesPlugin implements AIPlugin for Kubernetes operations
type KubernetesPlugin struct {
	kubeconfig string
	cluster    string
	namespace  string
}

// Command security levels
const (
	SecurityGreen  = "green"  // Read-only, no approval needed
	SecurityYellow = "yellow" // Write operations, approval needed
	SecurityRed    = "red"    // Dangerous, blocked or multi-approval
)

// NewKubernetesPlugin creates a new Kubernetes plugin
func NewKubernetesPlugin(kubeconfig, cluster, namespace string) *KubernetesPlugin {
	return &KubernetesPlugin{
		kubeconfig: kubeconfig,
		cluster:    cluster,
		namespace:  namespace,
	}
}

func (p *KubernetesPlugin) Name() string {
	return "kubernetes"
}

func (p *KubernetesPlugin) Description() string {
	return "Kubernetes cluster management via kubectl commands"
}

func (p *KubernetesPlugin) Commands() []aiplugin.CommandSpec {
	return []aiplugin.CommandSpec{
		// Green list - read-only
		{
			Name:             "get",
			Description:      "Get Kubernetes resources",
			Pattern:          "kubectl get {resource} {flags}",
			Args:             []string{"resource", "flags"},
			Security:         SecurityGreen,
			RequiresApproval: false,
		},
		{
			Name:             "describe",
			Description:      "Describe a Kubernetes resource",
			Pattern:          "kubectl describe {resource} {name} {flags}",
			Args:             []string{"resource", "name", "flags"},
			Security:         SecurityGreen,
			RequiresApproval: false,
		},
		{
			Name:             "logs",
			Description:      "Get logs from a pod",
			Pattern:          "kubectl logs {pod} {flags}",
			Args:             []string{"pod", "flags"},
			Security:         SecurityGreen,
			RequiresApproval: false,
		},
		{
			Name:             "top",
			Description:      "Show resource usage",
			Pattern:          "kubectl top {resource} {flags}",
			Args:             []string{"resource", "flags"},
			Security:         SecurityGreen,
			RequiresApproval: false,
		},
		{
			Name:             "explain",
			Description:      "Explain a Kubernetes resource",
			Pattern:          "kubectl explain {resource} {flags}",
			Args:             []string{"resource", "flags"},
			Security:         SecurityGreen,
			RequiresApproval: false,
		},

		// Yellow list - write operations
		{
			Name:             "create",
			Description:      "Create a Kubernetes resource",
			Pattern:          "kubectl create {resource} {flags}",
			Args:             []string{"resource", "flags"},
			Security:         SecurityYellow,
			RequiresApproval: true,
		},
		{
			Name:             "apply",
			Description:      "Apply a configuration to a resource",
			Pattern:          "kubectl apply {flags}",
			Args:             []string{"flags"},
			Security:         SecurityYellow,
			RequiresApproval: true,
		},
		{
			Name:             "delete",
			Description:      "Delete a Kubernetes resource",
			Pattern:          "kubectl delete {resource} {name} {flags}",
			Args:             []string{"resource", "name", "flags"},
			Security:         SecurityYellow,
			RequiresApproval: true,
		},
		{
			Name:             "scale",
			Description:      "Scale a deployment",
			Pattern:          "kubectl scale {resource} {flags}",
			Args:             []string{"resource", "flags"},
			Security:         SecurityYellow,
			RequiresApproval: true,
		},
		{
			Name:             "rollout",
			Description:      "Manage rollout of a deployment",
			Pattern:          "kubectl rollout {action} {resource} {flags}",
			Args:             []string{"action", "resource", "flags"},
			Security:         SecurityYellow,
			RequiresApproval: true,
		},

		// Red list - dangerous operations
		{
			Name:             "drain",
			Description:      "Drain a node",
			Pattern:          "kubectl drain {node} {flags}",
			Args:             []string{"node", "flags"},
			Security:         SecurityRed,
			RequiresApproval: true,
		},
		{
			Name:             "taint",
			Description:      "Taint a node",
			Pattern:          "kubectl taint {node} {flags}",
			Args:             []string{"node", "flags"},
			Security:         SecurityRed,
			RequiresApproval: true,
		},
		{
			Name:             "cordon",
			Description:      "Mark a node as unschedulable",
			Pattern:          "kubectl cordon {node} {flags}",
			Args:             []string{"node", "flags"},
			Security:         SecurityYellow,
			RequiresApproval: true,
		},
	}
}

func (p *KubernetesPlugin) Intents() []aiplugin.IntentHandler {
	return []aiplugin.IntentHandler{
		{
			Name:        "list_pods",
			Description: "List pods in a namespace",
			Patterns:    []string{"list pods", "show pods", "get pods", "查看 pod*", "显示 pod*"},
			Handler:     p.handleListPods,
		},
		{
			Name:        "list_deployments",
			Description: "List deployments",
			Patterns:    []string{"list deployments", "show deployments", "get deployments", "查看部署*"},
			Handler:     p.handleListDeployments,
		},
		{
			Name:        "get_logs",
			Description: "Get pod logs",
			Patterns:    []string{"logs for *", "show logs for *", "get logs from *", "查看*日志*"},
			Handler:     p.handleGetLogs,
		},
		{
			Name:        "navigate_to",
			Description: "Navigate to a page",
			Patterns:    []string{"open *", "show *", "go to *", "打开*", "显示*", "跳转到*"},
			Handler:     p.handleNavigate,
		},
		{
			Name:        "cluster_info",
			Description: "Get cluster information",
			Patterns:    []string{"cluster info", "cluster status", "集群信息*", "集群状态*"},
			Handler:     p.handleClusterInfo,
		},
	}
}

func (p *KubernetesPlugin) Initialize(ctx context.Context) error {
	// Verify kubectl is available
	cmd := exec.CommandContext(ctx, "kubectl", "version", "--client")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kubectl not available: %w", err)
	}
	return nil
}

func (p *KubernetesPlugin) Execute(ctx context.Context, cmdName string, args []string) (*aiplugin.CommandResult, error) {
	kubectlArgs := []string{}

	if p.kubeconfig != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", p.kubeconfig)
	}

	if p.cluster != "" {
		kubectlArgs = append(kubectlArgs, "--cluster", p.cluster)
	}

	kubectlArgs = append(kubectlArgs, cmdName)
	kubectlArgs = append(kubectlArgs, args...)

	// Execute kubectl command
	execCmd := exec.CommandContext(ctx, "kubectl", kubectlArgs...)

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()

	if err != nil {
		return &aiplugin.CommandResult{
			Success: false,
			Error:   stderr.String(),
		}, nil
	}

	return &aiplugin.CommandResult{
		Success: true,
		Output:  stdout.String(),
	}, nil
}

// Intent handlers

func (p *KubernetesPlugin) handleListPods(ctx context.Context, intent string, params map[string]string) (*aiplugin.IntentResult, error) {
	ns := p.namespace
	if ns == "" {
		ns = "default"
	}

	result, err := p.Execute(ctx, "get", []string{"pods", "-n", ns, "-o", "wide"})
	if err != nil {
		return &aiplugin.IntentResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list pods: %v", err),
		}, err
	}

	return &aiplugin.IntentResult{
		Success: true,
		Message: fmt.Sprintf("Pods in namespace %s:\n%s", ns, result.Output),
		Action:  "info",
		ActionData: map[string]interface{}{
			"namespace": ns,
			"resource":  "pods",
		},
	}, nil
}

func (p *KubernetesPlugin) handleListDeployments(ctx context.Context, intent string, params map[string]string) (*aiplugin.IntentResult, error) {
	ns := p.namespace
	if ns == "" {
		ns = "default"
	}

	result, err := p.Execute(ctx, "get", []string{"deployments", "-n", ns})
	if err != nil {
		return &aiplugin.IntentResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list deployments: %v", err),
		}, err
	}

	return &aiplugin.IntentResult{
		Success: true,
		Message: fmt.Sprintf("Deployments in namespace %s:\n%s", ns, result.Output),
		Action:  "info",
	}, nil
}

func (p *KubernetesPlugin) handleGetLogs(ctx context.Context, intent string, params map[string]string) (*aiplugin.IntentResult, error) {
	// Extract pod name from intent
	re := regexp.MustCompile(`(?:logs?|日志)\s+(?:for|from|的)?\s*(\S+)`)
	matches := re.FindStringSubmatch(intent)

	podName := "unknown"
	if len(matches) > 1 {
		podName = matches[1]
	}

	result, err := p.Execute(ctx, "logs", []string{podName, "-n", p.namespace, "--tail", "50"})
	if err != nil {
		return &aiplugin.IntentResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get logs: %v", err),
		}, err
	}

	return &aiplugin.IntentResult{
		Success: true,
		Message: fmt.Sprintf("Logs from pod %s:\n%s", podName, result.Output),
		Action:  "info",
	}, nil
}

func (p *KubernetesPlugin) handleNavigate(ctx context.Context, intent string, params map[string]string) (*aiplugin.IntentResult, error) {
	intent = strings.ToLower(intent)

	// Map intents to routes
	routes := map[string]string{
		"pod":        "/workloads",
		"pods":       "/workloads",
		"deploy":     "/workloads",
		"deployment": "/workloads",
		"workload":   "/workloads",
		"cluster":    "/clusters",
		"clusters":   "/clusters",
		"role":       "/roles",
		"rbac":       "/roles",
		"ingress":    "/ingress",
		"service":    "/services",
		"config":     "/configmaps",
		"secret":     "/secrets",
	}

	for keyword, route := range routes {
		if strings.Contains(intent, keyword) {
			return &aiplugin.IntentResult{
				Success: true,
				Message: fmt.Sprintf("Opening %s page", keyword),
				Action:  "navigate",
				ActionData: map[string]interface{}{
					"route": route,
				},
			}, nil
		}
	}

	return &aiplugin.IntentResult{
		Success: false,
		Message: "I'm not sure which page to open. Try saying 'open pods' or 'show clusters'.",
		Action:  "info",
	}, nil
}

func (p *KubernetesPlugin) handleClusterInfo(ctx context.Context, intent string, params map[string]string) (*aiplugin.IntentResult, error) {
	result, err := p.Execute(ctx, "cluster-info", []string{})
	if err != nil {
		return &aiplugin.IntentResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get cluster info: %v", err),
		}, err
	}

	return &aiplugin.IntentResult{
		Success: true,
		Message: fmt.Sprintf("Cluster information:\n%s", result.Output),
		Action:  "info",
	}, nil
}

// Security check helper
func CheckSecurityLevel(cmd string) string {
	cmd = strings.ToLower(cmd)

	// Red list
	redPatterns := []string{"drain", "taint", "delete --all", "delete all"}
	for _, pattern := range redPatterns {
		if strings.Contains(cmd, pattern) {
			return SecurityRed
		}
	}

	// Yellow list
	yellowPatterns := []string{"create", "apply", "delete", "scale", "rollout", "patch", "update"}
	for _, pattern := range yellowPatterns {
		if strings.Contains(cmd, pattern) {
			return SecurityYellow
		}
	}

	// Default to green
	return SecurityGreen
}

// EstimateCommandTime estimates how long a command might take
func EstimateCommandTime(cmd string, args []string) time.Duration {
	fullCmd := strings.Join(append([]string{cmd}, args...), " ")

	if strings.Contains(fullCmd, "logs") {
		return 5 * time.Second
	}
	if strings.Contains(fullCmd, "describe") {
		return 3 * time.Second
	}
	if strings.Contains(fullCmd, "get") {
		return 2 * time.Second
	}

	return 10 * time.Second
}
