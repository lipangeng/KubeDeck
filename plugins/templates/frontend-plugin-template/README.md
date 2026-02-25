# Frontend Plugin Template

## Create a New Plugin

1. Copy this folder to a new plugin directory.
2. Rename `pluginId`, `displayName`, and `version` in `plugin.manifest.json`.
3. Implement contributions in `src/index.ts`.

## Contribution Patterns

- Replacement page: return page metadata with `replacementFor`.
- Slots extension: provide `registerExtensions` entries.
- Menus: provide `registerMenus` entries.

## Run and Debug

Use the host frontend app scripts:

- `npm run dev`
- `npm run test`
