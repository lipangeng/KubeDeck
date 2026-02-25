package notification

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProviderFromEnvDefaultsToStub(t *testing.T) {
	t.Setenv("KUBEDECK_NOTIFICATION_PROVIDER", "")
	t.Setenv("KUBEDECK_NOTIFICATION_WEBHOOK_URL", "")

	provider := NewProviderFromEnv()
	if _, ok := provider.(*EmailStubProvider); !ok {
		t.Fatalf("expected EmailStubProvider, got %T", provider)
	}
}

func TestNewProviderFromEnvWebhook(t *testing.T) {
	t.Setenv("KUBEDECK_NOTIFICATION_PROVIDER", "webhook")
	t.Setenv("KUBEDECK_NOTIFICATION_WEBHOOK_URL", "http://127.0.0.1:18080/hook")

	provider := NewProviderFromEnv()
	if _, ok := provider.(*WebhookProvider); !ok {
		t.Fatalf("expected WebhookProvider, got %T", provider)
	}
}

func TestWebhookProviderSendEmailAndSMS(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	provider := NewWebhookProvider(server.URL)
	if err := provider.SendEmail("user@example.com", "subject", "body"); err != nil {
		t.Fatalf("send email: %v", err)
	}
	if err := provider.SendSMS("+123456", "hello"); err != nil {
		t.Fatalf("send sms: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("expected 2 webhook requests, got %d", requestCount)
	}
}

func TestWebhookProviderFailsOnErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	provider := NewWebhookProvider(server.URL)
	if err := provider.SendEmail("user@example.com", "subject", "body"); err == nil {
		t.Fatal("expected error for webhook 502 status")
	}
}
