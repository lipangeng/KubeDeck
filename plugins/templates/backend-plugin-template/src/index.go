package backendplugintemplate

import "kubedeck/backend/pkg/sdk"

type Plugin struct {
	id string
}

func New() sdk.Plugin {
	return &Plugin{id: "example-backend-plugin"}
}

func (p *Plugin) ID() string {
	return p.id
}
