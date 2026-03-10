package sdk

// PageDescriptor declares one frontend-visible workflow page.
type PageDescriptor struct {
	ID               string
	WorkflowDomainID string
	Route            string
	EntryKey         string
	Title            TextRef
	Description      *TextRef
}
