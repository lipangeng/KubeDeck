package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// RBACSyncer syncs RBAC roles between KubeDeck and Kubernetes
type RBACSyncer struct {
	clientset *kubernetes.Clientset
	namespace string
}

// PolicyRule represents a Kubernetes RBAC policy rule
type PolicyRule struct {
	APIGroups []string `json:"api_groups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// NewRBACSyncer creates a new RBAC syncer
func NewRBACSyncer(namespace string) (*RBACSyncer, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	if namespace == "" {
		namespace = "default"
	}

	return &RBACSyncer{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// getKubeConfig returns Kubernetes config from environment or file
func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home := os.Getenv("HOME")
		if home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// SyncRole syncs a KubeDeck role to Kubernetes ClusterRole/Role
func (s *RBACSyncer) SyncRole(ctx context.Context, name, description string, rulesJSON string, clusterScoped bool) error {
	var rules []PolicyRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return fmt.Errorf("failed to parse rules: %w", err)
	}

	// Convert to Kubernetes RBAC rules
	k8sRules := make([]rbacv1.PolicyRule, 0, len(rules))
	for _, rule := range rules {
		k8sRules = append(k8sRules, rbacv1.PolicyRule{
			APIGroups: rule.APIGroups,
			Resources: rule.Resources,
			Verbs:     rule.Verbs,
		})
	}

	// Create or update ClusterRole/Role
	if clusterScoped {
		return s.syncClusterRole(ctx, name, description, k8sRules)
	}
	return s.syncRole(ctx, name, description, k8sRules)
}

// syncClusterRole syncs to Kubernetes ClusterRole
func (s *RBACSyncer) syncClusterRole(ctx context.Context, name, description string, rules []rbacv1.PolicyRule) error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      map[string]string{"managed-by": "kubedeck"},
			Annotations: map[string]string{"description": description},
		},
		Rules: rules,
	}

	// Try to update existing
	_, err := s.clientset.RbacV1().ClusterRoles().Update(ctx, clusterRole, metav1.UpdateOptions{})
	if err != nil {
		// Create if not exists
		_, err = s.clientset.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ClusterRole: %w", err)
		}
	}

	return nil
}

// syncRole syncs to Kubernetes Role in namespace
func (s *RBACSyncer) syncRole(ctx context.Context, name, description string, rules []rbacv1.PolicyRule) error {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   s.namespace,
			Labels:      map[string]string{"managed-by": "kubedeck"},
			Annotations: map[string]string{"description": description},
		},
		Rules: rules,
	}

	// Try to update existing
	_, err := s.clientset.RbacV1().Roles(s.namespace).Update(ctx, role, metav1.UpdateOptions{})
	if err != nil {
		// Create if not exists
		_, err = s.clientset.RbacV1().Roles(s.namespace).Create(ctx, role, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Role: %w", err)
		}
	}

	return nil
}

// SyncRoleBinding syncs a KubeDeck role binding to Kubernetes
func (s *RBACSyncer) SyncRoleBinding(ctx context.Context, name, roleName, userName, userEmail string, clusterScoped bool) error {
	if clusterScoped {
		return s.syncClusterRoleBinding(ctx, name, roleName, userName, userEmail)
	}
	return s.syncRoleBinding(ctx, name, roleName, userName, userEmail)
}

// syncClusterRoleBinding syncs to Kubernetes ClusterRoleBinding
func (s *RBACSyncer) syncClusterRoleBinding(ctx context.Context, name, roleName, userName, userEmail string) error {
	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"managed-by": "kubedeck"},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     userEmail,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// Try to update existing
	_, err := s.clientset.RbacV1().ClusterRoleBindings().Update(ctx, binding, metav1.UpdateOptions{})
	if err != nil {
		// Create if not exists
		_, err = s.clientset.RbacV1().ClusterRoleBindings().Create(ctx, binding, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ClusterRoleBinding: %w", err)
		}
	}

	return nil
}

// syncRoleBinding syncs to Kubernetes RoleBinding
func (s *RBACSyncer) syncRoleBinding(ctx context.Context, name, roleName, userName, userEmail string) error {
	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: s.namespace,
			Labels:    map[string]string{"managed-by": "kubedeck"},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     userEmail,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// Try to update existing
	_, err := s.clientset.RbacV1().RoleBindings(s.namespace).Update(ctx, binding, metav1.UpdateOptions{})
	if err != nil {
		// Create if not exists
		_, err = s.clientset.RbacV1().RoleBindings(s.namespace).Create(ctx, binding, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create RoleBinding: %w", err)
		}
	}

	return nil
}

// DeleteRole deletes a role from Kubernetes
func (s *RBACSyncer) DeleteRole(ctx context.Context, name string, clusterScoped bool) error {
	if clusterScoped {
		err := s.clientset.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete ClusterRole: %w", err)
		}
	} else {
		err := s.clientset.RbacV1().Roles(s.namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete Role: %w", err)
		}
	}
	return nil
}

// DeleteRoleBinding deletes a role binding from Kubernetes
func (s *RBACSyncer) DeleteRoleBinding(ctx context.Context, name string, clusterScoped bool) error {
	if clusterScoped {
		err := s.clientset.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete ClusterRoleBinding: %w", err)
		}
	} else {
		err := s.clientset.RbacV1().RoleBindings(s.namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete RoleBinding: %w", err)
		}
	}
	return nil
}

// VerifyRoleBinding verifies if a user has a role in Kubernetes
func (s *RBACSyncer) VerifyRoleBinding(ctx context.Context, roleName, userEmail string, clusterScoped bool) (bool, error) {
	if clusterScoped {
		binding, err := s.clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, b := range binding.Items {
			if b.RoleRef.Name == roleName {
				for _, subject := range b.Subjects {
					if subject.Name == userEmail {
						return true, nil
					}
				}
			}
		}
	} else {
		binding, err := s.clientset.RbacV1().RoleBindings(s.namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, b := range binding.Items {
			if b.RoleRef.Name == roleName {
				for _, subject := range b.Subjects {
					if subject.Name == userEmail {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}
