package sdk

import "context"

// PluginDescriptor describes a plugin
type PluginDescriptor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Homepage    string `json:"homepage,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// Capability represents a plugin capability
type Capability struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"` // api, ui, webhook
	Actions     []Action `json:"actions,omitempty"`
	Pages       []Page   `json:"pages,omitempty"`
}

// Action represents a plugin action
type Action struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// Page represents a plugin page
type Page struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
}

// Plugin interface that all plugins must implement
type Plugin interface {
	Descriptor() PluginDescriptor
	Capabilities() []Capability
	Execute(ctx context.Context, action string, params map[string]interface{}) (interface{}, error)
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
