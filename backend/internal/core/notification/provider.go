package notification

import "errors"

var ErrSMSNotImplemented = errors.New("sms provider not implemented")

// Provider defines outbound notification capabilities used by IAM workflows.
type Provider interface {
	SendEmail(to, subject, body string) error
	SendSMS(phone, content string) error
}

