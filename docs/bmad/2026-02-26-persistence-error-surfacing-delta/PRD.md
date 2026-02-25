# PRD: Persistence Error Surfacing Delta

## Goals
- P0: Make rebuild API return explicit 500 when persistence init/write fails.
- P0: Keep backward compatibility for existing non-strict persist callers.

## Acceptance Criteria
- AC1: rebuild endpoint returns `rebuild_failed` on persistence failure.
- AC2: strict persistence helpers return concrete errors.
- AC3: existing legacy persist wrappers still log failures for compatibility.
