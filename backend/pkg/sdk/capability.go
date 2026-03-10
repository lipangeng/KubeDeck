package sdk

// TextRef carries user-facing metadata in a future-localizable shape.
type TextRef struct {
	Key         string
	Fallback    string
	Description string
}

// CapabilityDescriptor declares what one built-in capability or plugin provides.
type CapabilityDescriptor struct {
	ID      string
	Version string
	Pages   []PageDescriptor
	Menus   []MenuDescriptor
	Actions []ActionDescriptor
}

// CapabilityProvider exposes one capability descriptor to the kernel.
type CapabilityProvider interface {
	CapabilityDescriptor() CapabilityDescriptor
}
