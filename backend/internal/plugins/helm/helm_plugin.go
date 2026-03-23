package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"kubedeck/backend/pkg/sdk"
)

// HelmPlugin implements Helm functionality as a plugin
type HelmPlugin struct{}

// HelmProvider implements the SDK Provider interface
type HelmProvider struct{}

// Descriptor returns the plugin descriptor
func (p *HelmPlugin) Descriptor() sdk.PluginDescriptor {
	return sdk.PluginDescriptor{
		ID:          "core.helm",
		Name:        "Helm Chart Manager",
		Version:     "v1",
		Description: "Manage Helm charts and releases",
		Author:      "KubeDeck Team",
	}
}

// Capabilities returns the plugin capabilities
func (p *HelmPlugin) Capabilities() []sdk.Capability {
	return []sdk.Capability{
		{
			ID:          "helm.releases",
			Name:        "Helm Releases",
			Description: "List and manage Helm releases",
			Type:        "api",
			Actions: []sdk.Action{
				{ID: "list", Name: "List Releases"},
				{ID: "install", Name: "Install Chart"},
				{ID: "upgrade", Name: "Upgrade Release"},
				{ID: "uninstall", Name: "Uninstall Release"},
			},
		},
		{
			ID:          "helm.repos",
			Name:        "Helm Repositories",
			Description: "Manage Helm repositories",
			Type:        "api",
			Actions: []sdk.Action{
				{ID: "repo.list", Name: "List Repositories"},
				{ID: "repo.add", Name: "Add Repository"},
				{ID: "repo.remove", Name: "Remove Repository"},
			},
		},
		{
			ID:          "helm.charts",
			Name:        "Helm Charts",
			Description: "Search and browse Helm charts",
			Type:        "api",
			Actions: []sdk.Action{
				{ID: "search", Name: "Search Charts"},
			},
		},
	}
}

// Execute executes a plugin action
func (p *HelmPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (interface{}, error) {
	switch action {
	case "releases.list":
		return listReleases(params)
	case "releases.install":
		return installChart(ctx, params)
	case "releases.upgrade":
		return upgradeRelease(ctx, params)
	case "releases.uninstall":
		return uninstallRelease(ctx, params)
	case "repos.list":
		return listRepos()
	case "repos.add":
		return addRepo(params)
	case "repos.remove":
		return removeRepo(params)
	case "charts.search":
		return searchCharts(params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// HelmRelease represents a Helm release
type HelmRelease struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   int    `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

func listReleases(params map[string]interface{}) ([]HelmRelease, error) {
	namespace := "default"
	if ns, ok := params["namespace"].(string); ok {
		namespace = ns
	}

	cmd := exec.Command("helm", "list", "-n", namespace, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return []HelmRelease{}, nil // Return empty if helm not installed or no releases
	}

	var releases []HelmRelease
	if err := json.Unmarshal(output, &releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// InstallParams represents parameters for installing a chart
type InstallParams struct {
	Name      string            `json:"name"`
	Chart     string            `json:"chart"`
	Namespace string            `json:"namespace"`
	Values    map[string]string `json:"values,omitempty"`
}

func installChart(ctx context.Context, params map[string]interface{}) (map[string]string, error) {
	var p InstallParams
	data, _ := json.Marshal(params)
	json.Unmarshal(data, &p)

	if p.Name == "" || p.Chart == "" {
		return nil, fmt.Errorf("name and chart required")
	}

	if p.Namespace == "" {
		p.Namespace = "default"
	}

	args := []string{"install", p.Name, p.Chart, "-n", p.Namespace}
	for k, v := range p.Values {
		args = append(args, "--set", k+"="+v)
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]string{
			"success": "false",
			"error":   string(output),
		}, err
	}

	return map[string]string{
		"success": "true",
		"message": "Helm chart installed successfully",
	}, nil
}

func upgradeRelease(ctx context.Context, params map[string]interface{}) (map[string]string, error) {
	var p InstallParams
	data, _ := json.Marshal(params)
	json.Unmarshal(data, &p)

	if p.Name == "" || p.Chart == "" {
		return nil, fmt.Errorf("name and chart required")
	}

	if p.Namespace == "" {
		p.Namespace = "default"
	}

	args := []string{"upgrade", p.Name, p.Chart, "-n", p.Namespace}
	for k, v := range p.Values {
		args = append(args, "--set", k+"="+v)
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]string{
			"success": "false",
			"error":   string(output),
		}, err
	}

	return map[string]string{
		"success": "true",
		"message": "Helm release upgraded successfully",
	}, nil
}

func uninstallRelease(ctx context.Context, params map[string]interface{}) (map[string]string, error) {
	name, _ := params["name"].(string)
	namespace, _ := params["namespace"].(string)

	if name == "" {
		return nil, fmt.Errorf("name required")
	}

	if namespace == "" {
		namespace = "default"
	}

	cmd := exec.CommandContext(ctx, "helm", "uninstall", name, "-n", namespace)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]string{
			"success": "false",
			"error":   string(output),
		}, err
	}

	return map[string]string{
		"success": "true",
		"message": "Helm release uninstalled successfully",
	}, nil
}

// HelmRepo represents a Helm repository
type HelmRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func listRepos() ([]HelmRepo, error) {
	cmd := exec.Command("helm", "repo", "list", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return []HelmRepo{}, nil
	}

	var repos []HelmRepo
	if err := json.Unmarshal(output, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// AddRepoParams represents parameters for adding a repository
type AddRepoParams struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func addRepo(params map[string]interface{}) (map[string]string, error) {
	var p AddRepoParams
	data, _ := json.Marshal(params)
	json.Unmarshal(data, &p)

	if p.Name == "" || p.URL == "" {
		return nil, fmt.Errorf("name and url required")
	}

	cmd := exec.Command("helm", "repo", "add", p.Name, p.URL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]string{
			"success": "false",
			"error":   string(output),
		}, err
	}

	return map[string]string{
		"success": "true",
		"message": "Helm repository added successfully",
	}, nil
}

func removeRepo(params map[string]interface{}) (map[string]string, error) {
	name, _ := params["name"].(string)

	if name == "" {
		return nil, fmt.Errorf("name required")
	}

	cmd := exec.Command("helm", "repo", "remove", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]string{
			"success": "false",
			"error":   string(output),
		}, err
	}

	return map[string]string{
		"success": "true",
		"message": "Helm repository removed successfully",
	}, nil
}

// ChartInfo represents a Helm chart
type ChartInfo struct {
	Name        string `json:"name"`
	Chart       string `json:"chart"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}

func searchCharts(params map[string]interface{}) ([]ChartInfo, error) {
	repo, _ := params["repo"].(string)

	if repo == "" {
		return nil, fmt.Errorf("repo required")
	}

	cmd := exec.Command("helm", "search", "repo", repo, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return []ChartInfo{}, nil
	}

	var charts []ChartInfo
	if err := json.Unmarshal(output, &charts); err != nil {
		return nil, err
	}

	return charts, nil
}

// Initialize initializes the plugin
func (p *HelmPlugin) Initialize(ctx context.Context) error {
	// Check if helm is installed
	cmd := exec.Command("helm", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm not installed: %w", err)
	}
	return nil
}

// Shutdown cleans up plugin resources
func (p *HelmPlugin) Shutdown(ctx context.Context) error {
	return nil
}
