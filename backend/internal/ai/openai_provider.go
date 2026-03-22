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

// OpenAIProvider implements AIProvider for OpenAI-compatible APIs
type OpenAIProvider struct {
	config *ProviderConfig
	client *http.Client
}

// OpenAIChatRequest represents an OpenAI chat request
type OpenAIChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// OpenAIChatResponse represents an OpenAI chat response
type OpenAIChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenAIStreamChunk represents an OpenAI stream chunk
type OpenAIStreamChunk struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// OpenAIModel represents an OpenAI model
type OpenAIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// OpenAIModelsResponse represents an OpenAI models response
type OpenAIModelsResponse struct {
	Data []OpenAIModel `json:"data"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *ProviderConfig) *OpenAIProvider {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return &OpenAIProvider{
		config: config,
		client: &http.Client{Timeout: timeout},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	reqBody := OpenAIChatRequest{
		Model:    p.config.Model,
		Messages: messages,
		Stream:   false,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
		return nil, fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(body))
	}

	var openaiResp OpenAIChatResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, err
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	chatResp := &Response{
		Content:      openaiResp.Choices[0].Message.Content,
		Model:        openaiResp.Model,
		FinishReason: openaiResp.Choices[0].FinishReason,
	}

	if openaiResp.Usage != nil {
		chatResp.Usage = &Usage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		}
	}

	return chatResp, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, messages []Message) (<-chan StreamChunk, error) {
	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	reqBody := OpenAIChatRequest{
		Model:    p.config.Model,
		Messages: messages,
		Stream:   true,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")

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
				var chunk OpenAIStreamChunk
				if err := decoder.Decode(&chunk); err != nil {
					if err == io.EOF {
						chunkChan <- StreamChunk{Done: true}
						return
					}
					return
				}

				if len(chunk.Choices) > 0 {
					content := chunk.Choices[0].Delta.Content
					done := chunk.Choices[0].FinishReason != ""

					if content != "" || done {
						chunkChan <- StreamChunk{
							Content: content,
							Done:    done,
						}
					}

					if done {
						return
					}
				}
			}
		}
	}()

	return chunkChan, nil
}

func (p *OpenAIProvider) Models(ctx context.Context) ([]Model, error) {
	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var openaiResp OpenAIModelsResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(openaiResp.Data))
	for _, m := range openaiResp.Data {
		models = append(models, Model{
			ID:       m.ID,
			Name:     m.ID,
			Provider: "openai",
		})
	}

	return models, nil
}
