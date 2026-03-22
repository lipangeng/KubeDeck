package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaProvider implements AIProvider for Ollama
type OllamaProvider struct {
	config *ProviderConfig
	client *http.Client
}

// OllamaChatRequest represents an Ollama chat request
type OllamaChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// OllamaChatResponse represents an Ollama chat response
type OllamaChatResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// OllamaModelResponse represents an Ollama models response
type OllamaModelResponse struct {
	Models []struct {
		Name       string `json:"name"`
		Size       uint64 `json:"size"`
		Digest     string `json:"digest"`
		ModifiedAt string `json:"modified_at"`
	} `json:"models"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config *ProviderConfig) *OllamaProvider {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second
	}

	return &OllamaProvider{
		config: config,
		client: &http.Client{Timeout: timeout},
	}
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

func (p *OllamaProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	reqBody := OllamaChatRequest{
		Model:    p.config.Model,
		Messages: messages,
		Stream:   false,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/chat", bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error: %s", string(body))
	}

	var ollamaResp OllamaChatResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, err
	}

	return &Response{
		Content: ollamaResp.Message.Content,
		Model:   ollamaResp.Model,
	}, nil
}

func (p *OllamaProvider) Stream(ctx context.Context, messages []Message) (<-chan StreamChunk, error) {
	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	reqBody := OllamaChatRequest{
		Model:    p.config.Model,
		Messages: messages,
		Stream:   true,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/chat", bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	chunkChan := make(chan StreamChunk, 64)

	go func() {
		defer close(chunkChan)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var ollamaResp OllamaChatResponse
				if err := decoder.Decode(&ollamaResp); err != nil {
					if err == io.EOF {
						chunkChan <- StreamChunk{Done: true}
						return
					}
					return
				}

				chunkChan <- StreamChunk{
					Content: ollamaResp.Message.Content,
					Done:    ollamaResp.Done,
				}

				if ollamaResp.Done {
					return
				}
			}
		}
	}()

	return chunkChan, nil
}

func (p *OllamaProvider) Models(ctx context.Context) ([]Model, error) {
	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ollamaResp OllamaModelResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(ollamaResp.Models))
	for _, m := range ollamaResp.Models {
		models = append(models, Model{
			ID:       m.Name,
			Name:     m.Name,
			Provider: "ollama",
		})
	}

	return models, nil
}
