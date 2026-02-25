package notification

import (
	"os"
	"strings"
)

func NewProviderFromEnv() Provider {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("KUBEDECK_NOTIFICATION_PROVIDER")))
	switch provider {
	case "webhook":
		webhookURL := strings.TrimSpace(os.Getenv("KUBEDECK_NOTIFICATION_WEBHOOK_URL"))
		if webhookURL == "" {
			return NewEmailStubProvider()
		}
		return NewWebhookProvider(webhookURL)
	case "", "stub":
		return NewEmailStubProvider()
	default:
		return NewEmailStubProvider()
	}
}
