package sdk

// Plugin represents a backend plugin registered in the kernel.
type Plugin interface {
	ID() string
}
