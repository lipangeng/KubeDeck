# UI Polish Plan

## Scope
- Target: `frontend/src/App.tsx` top status zone and login dialog.
- Constraint: style-only changes, no business logic changes.
- Goal: modern, concise, sci-fi visual language with blue glassmorphism, compatible with light and dark themes.

## Visual Style
- Theme direction: blue-first cyber control panel with translucent surfaces, subtle glow, and layered gradients.
- Surface tokens:
  - Use semi-transparent surfaces and stronger border contrast in dark mode.
  - Keep blur moderate (`8px-14px`) for readability and performance.
- Accent language:
  - Runtime and health signals use existing semantic chip colors.
  - Primary action remains gradient blue for visual focus.
- Motion and depth:
  - Prefer static depth (shadow + inset highlight) over heavy animation.

## Information Hierarchy
- Header hierarchy:
  - Level 1: product identity (`KubeDeck`).
  - Level 2: runtime status panel (`Runtime/Health/Ready/Checked`).
  - Level 3: controls (cluster/theme/language + actions).
- Login dialog hierarchy:
  - Level 1: dialog title.
  - Level 2: visual security context panel.
  - Level 3: form fields and actions.
- Keep high-frequency status information grouped into one compact panel.

## Navigation and Top Zone
- Group top health signals into a single glass panel to reduce scanning cost.
- Keep operational selectors (`cluster`, `theme`, `language`) in the same row for quick context switching.
- Preserve existing action order and semantics to avoid behavior regression.
- Maintain sticky top bar for persistent situational awareness.

## Form Usability
- Login form improvements:
  - Set `username` as autofocus.
  - Use compact (`size="small"`) input style for denser dialogs.
  - Add contextual helper line for tenant awareness.
- Error visibility:
  - Preserve existing inline error and OAuth diagnostics placement.
- Accessibility:
  - Keep label-driven input structure and dialog semantics unchanged.

## Mobile Experience
- Header layout wraps instead of overflowing.
- Status panel remains readable in narrow widths with chip-based compact display.
- Dialog sizing and spacing stay within `maxWidth="xs"` while preserving touchable actions.
- No interaction model changes between desktop and mobile.

## Implementation Checklist
- [x] Add top status glass panel container and compact chip group.
- [x] Improve app bar translucency behavior across light/dark modes.
- [x] Add login dialog visual shell and polish form controls.
- [x] Preserve all existing login flows and actions.
- [x] Add UI regression test anchors for key polished areas.

## Validation
- Unit/UI tests: `cd frontend && npm test -- --run`
- Build: `cd frontend && npm run build`
- Manual checks:
  - Light theme and dark theme readability.
  - Login dialog open/close and submit behavior unchanged.
  - Header still usable on narrow width.
