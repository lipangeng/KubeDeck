package aiplugin

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// CommandSpec defines a command that a plugin can execute
type CommandSpec struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Pattern          string   `json:"pattern"`  // Regex pattern for matching
	Args             []string `json:"args"`     // Expected arguments
	Security         string   `json:"security"` // green, yellow, red
	RequiresApproval bool     `json:"requires_approval"`
}

// IntentHandler handles natural language intents
type IntentHandler struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Patterns    []string `json:"patterns"` // Natural language patterns
	Handler     IntentHandlerFunc
}

// IntentHandlerFunc is the function signature for intent handlers
type IntentHandlerFunc func(ctx context.Context, intent string, params map[string]string) (*IntentResult, error)

// IntentResult represents the result of intent processing
type IntentResult struct {
	Success          bool                   `json:"success"`
	Message          string                 `json:"message"`
	Action           string                 `json:"action"` // navigate, execute, info
	ActionData       map[string]interface{} `json:"action_data"`
	RequiresApproval bool                   `json:"requires_approval"`
}

// AIPlugin defines the interface for AI plugins
type AIPlugin interface {
	// Name returns the plugin name
	Name() string

	// Description returns the plugin description
	Description() string

	// Commands returns the commands this plugin supports
	Commands() []CommandSpec

	// Intents returns the intent handlers this plugin supports
	Intents() []IntentHandler

	// Execute executes a command
	Execute(ctx context.Context, cmd string, args []string) (*CommandResult, error)

	// Initialize initializes the plugin
	Initialize(ctx context.Context) error
}

// CommandResult represents the result of command execution
type CommandResult struct {
	Success  bool        `json:"success"`
	Output   string      `json:"output"`
	Error    string      `json:"error,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

// Manager manages AI plugins
type Manager struct {
	plugins  map[string]AIPlugin
	commands map[string]*CommandSpec
	intents  []IntentHandler
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins:  make(map[string]AIPlugin),
		commands: make(map[string]*CommandSpec),
		intents:  []IntentHandler{},
	}
}

// Register registers a plugin
func (m *Manager) Register(plugin AIPlugin) error {
	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	m.plugins[name] = plugin

	// Register commands
	for _, cmd := range plugin.Commands() {
		cmdCopy := cmd
		key := fmt.Sprintf("%s:%s", name, cmd.Name)
		m.commands[key] = &cmdCopy
	}

	// Register intents
	for _, intent := range plugin.Intents() {
		intentCopy := intent
		m.intents = append(m.intents, intentCopy)
	}

	return nil
}

// GetPlugin returns a plugin by name
func (m *Manager) GetPlugin(name string) (AIPlugin, error) {
	plugin, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return plugin, nil
}

// ListPlugins returns all registered plugins
func (m *Manager) ListPlugins() []string {
	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}

// GetCommand returns a command spec
func (m *Manager) GetCommand(key string) (*CommandSpec, error) {
	cmd, ok := m.commands[key]
	if !ok {
		return nil, fmt.Errorf("command %s not found", key)
	}
	return cmd, nil
}

// ListCommands returns all registered commands
func (m *Manager) ListCommands() []CommandSpec {
	cmds := make([]CommandSpec, 0, len(m.commands))
	for _, cmd := range m.commands {
		cmds = append(cmds, *cmd)
	}
	return cmds
}

// ProcessIntent processes a natural language intent
func (m *Manager) ProcessIntent(ctx context.Context, input string) (*IntentResult, error) {
	input = strings.TrimSpace(input)

	// Try to match against intent patterns
	for _, intent := range m.intents {
		for _, pattern := range intent.Patterns {
			if m.matchPattern(pattern, input) {
				params := m.extractParams(pattern, input)
				return intent.Handler(ctx, input, params)
			}
		}
	}

	// Try to match against command patterns
	for _, cmd := range m.commands {
		if m.matchPattern(cmd.Pattern, input) {
			if cmd.RequiresApproval {
				return &IntentResult{
					Success:          true,
					Message:          fmt.Sprintf("Command %s requires approval", cmd.Name),
					Action:           "approve",
					RequiresApproval: true,
				}, nil
			}
		}
	}

	// No match found
	return &IntentResult{
		Success: false,
		Message: "I didn't understand that request. Try asking about Kubernetes resources or using commands like 'kubectl get pods'.",
		Action:  "info",
	}, nil
}

// ExecuteCommand executes a command
func (m *Manager) ExecuteCommand(ctx context.Context, pluginName, cmdName string, args []string) (*CommandResult, error) {
	plugin, ok := m.plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return plugin.Execute(ctx, cmdName, args)
}

// matchPattern checks if input matches a pattern
func (m *Manager) matchPattern(pattern, input string) bool {
	// Convert pattern to regex
	regexPattern := pattern
	regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "{", "(?P<")
	regexPattern = strings.ReplaceAll(regexPattern, "}", ">[^/]+)")

	matched, _ := regexp.MatchString("(?i)"+regexPattern, input)
	return matched
}

// extractParams extracts parameters from input based on pattern
func (m *Manager) extractParams(pattern, input string) map[string]string {
	params := make(map[string]string)

	// Simple extraction - can be enhanced
	re := regexp.MustCompile(`\{(\w+)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)

	for _, match := range matches {
		if len(match) > 1 {
			paramName := match[1]
			params[paramName] = ""
		}
	}

	return params
}
