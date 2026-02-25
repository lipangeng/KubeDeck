package notification

// EmailStubProvider is a no-op provider for MVP flow wiring.
type EmailStubProvider struct{}

func NewEmailStubProvider() *EmailStubProvider {
	return &EmailStubProvider{}
}

func (p *EmailStubProvider) SendEmail(_, _, _ string) error {
	return nil
}

func (p *EmailStubProvider) SendSMS(_, _ string) error {
	return ErrSMSNotImplemented
}

