package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type WebhookProvider struct {
	url    string
	client *http.Client
}

func NewWebhookProvider(url string) *WebhookProvider {
	return &WebhookProvider{
		url: strings.TrimSpace(url),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (p *WebhookProvider) SendEmail(to, subject, body string) error {
	payload := map[string]string{
		"type":    "email",
		"to":      to,
		"subject": subject,
		"body":    body,
	}
	return p.send(payload)
}

func (p *WebhookProvider) SendSMS(phone, content string) error {
	payload := map[string]string{
		"type":    "sms",
		"phone":   phone,
		"content": content,
	}
	return p.send(payload)
}

func (p *WebhookProvider) send(payload map[string]string) error {
	if p.url == "" {
		return fmt.Errorf("notification webhook url is empty")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, p.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification webhook returned status %d", resp.StatusCode)
	}
	return nil
}
