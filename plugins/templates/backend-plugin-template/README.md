# Backend Plugin Template

## Create a New Plugin

1. Copy this folder to a new plugin directory.
2. Rename `pluginId`, `displayName`, and `version` in `plugin.manifest.json`.
3. Implement backend plugin logic in `src/index.go`.

## Contribution Patterns

- Expose one `sdk.CapabilityProvider`.
- Return kernel contributions through `CapabilityDescriptor`.
- Add `pages`, `menus`, `actions`, and `slots` only when the capability actually owns them.

## Run and Debug

Use backend scripts from repository root:

- `cd backend && go test ./...`
- `cd backend && go run ./cmd/kubedeck`
