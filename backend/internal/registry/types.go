package registry

// ResourceType describes one API resource that can be listed and viewed.
type ResourceType struct {
	ID               string
	Group            string
	Version          string
	Kind             string
	Plural           string
	Namespaced       bool
	PreferredVersion string
	Source           string
}

// PageMeta describes page registration metadata from system/plugins.
type PageMeta struct {
	PageID         string
	Route          string
	PluginID       string
	ReplacementFor string
	Slots          []string
}

// SlotMeta describes extension points available on a page.
type SlotMeta struct {
	SlotID   string
	PageID   string
	Accepts  string
	Ordering string
}

// MenuItem describes one sidebar/menu item.
type MenuItem struct {
	ID              string
	Group           string
	TitleI18nKey    string
	TargetType      string
	TargetRef       string
	Source          string
	Order           int
	Visible         bool
	PermissionHints []string
}

// Snapshot is the composed registry payload returned by metadata API.
type Snapshot struct {
	ResourceTypes []ResourceType
	Pages         []PageMeta
	Slots         []SlotMeta
	Menus         []MenuItem
}
