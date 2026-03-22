package ai

import (
	"context"
	"errors"
)

var (
	ErrProviderNotConfigured = errors.New("AI provider not configured")
	ErrStreamingNotSupported = errors.New("streaming not supported by this provider")
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // Message content
}

// Response represents an AI response
type Response struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	Usage        *Usage `json:"usage,omitempty"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// Model represents an available AI model
type Model struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	MaxTokens  int    `json:"max_tokens"`
	ContextLen int    `json:"context_len"`
}

// AIProvider defines the interface for AI backends
type AIProvider interface {
	// Chat sends a chat request and returns a response
	Chat(ctx context.Context, messages []Message) (*Response, error)

	// Stream sends a chat request and returns a stream of responses
	Stream(ctx context.Context, messages []Message) (<-chan StreamChunk, error)

	// Models returns available models
	Models(ctx context.Context) ([]Model, error)

	// Name returns the provider name
	Name() string
}

// ProviderConfig holds provider configuration
type ProviderConfig struct {
	Type    string            `json:"type"` // ollama, openai, gemini, azure
	BaseURL string            `json:"base_url"`
	APIKey  string            `json:"api_key"`
	Model   string            `json:"model"`
	Timeout int               `json:"timeout"` // seconds
	Extra   map[string]string `json:"extra"`
}

// Manager manages AI providers
type Manager struct {
	providers map[string]AIProvider
	defaultID string
}

// NewManager creates a new AI manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]AIProvider),
	}
}

// Register registers an AI provider
func (m *Manager) Register(id string, provider AIProvider) {
	m.providers[id] = provider
	if m.defaultID == "" {
		m.defaultID = id
	}
}

// GetProvider returns a provider by ID
func (m *Manager) GetProvider(id string) (AIProvider, error) {
	provider, ok := m.providers[id]
	if !ok {
		return nil, ErrProviderNotConfigured
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (m *Manager) GetDefaultProvider() (AIProvider, error) {
	if m.defaultID == "" {
		return nil, ErrProviderNotConfigured
	}
	return m.providers[m.defaultID], nil
}

// Chat sends a chat request using the default provider
func (m *Manager) Chat(ctx context.Context, messages []Message) (*Response, error) {
	provider, err := m.GetDefaultProvider()
	if err != nil {
		return nil, err
	}
	return provider.Chat(ctx, messages)
}

// Stream sends a chat request using the default provider
func (m *Manager) Stream(ctx context.Context, messages []Message) (<-chan StreamChunk, error) {
	provider, err := m.GetDefaultProvider()
	if err != nil {
		return nil, err
	}
	return provider.Stream(ctx, messages)
}

// Models returns all available models from all providers
func (m *Manager) Models(ctx context.Context) ([]Model, error) {
	var allModels []Model
	for id, provider := range m.providers {
		models, err := provider.Models(ctx)
		if err != nil {
			continue
		}
		for _, model := range models {
			model.Provider = id
			allModels = append(allModels, model)
		}
	}
	return allModels, nil
}

// SetDefault sets the default provider
func (m *Manager) SetDefault(id string) error {
	if _, ok := m.providers[id]; !ok {
		return ErrProviderNotConfigured
	}
	m.defaultID = id
	return nil
}
