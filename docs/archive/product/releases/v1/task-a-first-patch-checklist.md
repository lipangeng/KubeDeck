# KubeDeck Task A First Patch Checklist

## 1. Purpose

This document defines the final checklist to review before starting and after completing the first patch for Task A: Shared Working Context Model.

Its purpose is to prevent the first patch from drifting beyond its structural goal.

## 2. Before-Start Checklist

All of the following should be true before implementation begins:

### Scope Checks

- [ ] The team accepts that the first patch is structural, not UI-complete.
- [ ] The team accepts that Homepage and Workloads migration is not part of this patch.
- [ ] The team accepts that old code may coexist temporarily.
- [ ] The team also accepts that old code is not the design source.

### Model Checks

- [ ] The shared working context model is accepted.
- [ ] The shared working context state schema is accepted.
- [ ] The events and update rules are accepted.
- [ ] The implementation mapping is accepted.
- [ ] The file responsibility draft is accepted.

### Boundary Checks

- [ ] V1 boundary still excludes AI.
- [ ] V1 boundary still excludes broad plugin workflow expansion.
- [ ] The first patch still excludes UI rewrites and visual redesign.

### Patch-Scope Checks

- [ ] The file set for the first patch is fixed.
- [ ] The first patch will create new `state/work-context` files first.
- [ ] The first patch will not delete historical page/state code yet.
- [ ] The first patch test scope is limited to state rules and event rules.

## 3. In-Progress Guardrails

While the first patch is being built, these conditions must remain true:

- [ ] No Homepage visual redesign work is added.
- [ ] No Workloads page rewrite is added.
- [ ] No Create / Apply UI work is added.
- [ ] No plugin system work is added.
- [ ] No backend API redesign is added.
- [ ] No AI-related scope enters the patch.
- [ ] No old file is deleted before the replacement structure exists.

If any of these becomes false, the patch scope has drifted.

## 4. Required Deliverables

The first patch must deliver:

- [ ] new `frontend/src/state/work-context/` file structure
- [ ] root shared state types
- [ ] cluster context logic
- [ ] namespace context logic
- [ ] workflow context logic
- [ ] action context logic
- [ ] list context structure
- [ ] event/update entry points
- [ ] selectors for future page consumption
- [ ] unit tests covering the core state rules

## 5. Required Validation

At minimum, validation after the patch should confirm:

- [ ] cluster switch state can be represented cleanly
- [ ] namespace scope modes can be represented cleanly
- [ ] execution target is structurally separate from browsing scope
- [ ] event handlers do not rewrite unrelated continuity state
- [ ] shared context files compile and tests pass

## 6. Drift Indicators

The first patch has gone off track if any of these happens:

- [ ] `App.tsx` becomes the place where the new model is redefined
- [ ] namespace browsing scope and execution target are collapsed together
- [ ] UI concerns dominate the new state files
- [ ] the patch introduces page migration before the shared structure is stable
- [ ] the patch starts cleaning up old code without replacement paths in place

## 7. Done Criteria

The first patch is done when:

- [ ] the new structural ownership exists
- [ ] core shared context concepts now have a canonical home
- [ ] the patch remains inside the first-patch scope
- [ ] no accidental UI migration has been mixed in
- [ ] the repository is ready for the second patch that wires pages onto the new context foundation

## 8. Next-Patch Gate

The next patch should not begin until the following are all true:

- [ ] the first patch is merged or otherwise accepted as the canonical base
- [ ] the shared context file set is no longer in dispute
- [ ] event/update ownership is no longer in dispute
- [ ] the team agrees that the next patch is about controlled page wiring, not more state redesign
