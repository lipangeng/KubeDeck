# Tech Plan Delta

- Add router mapping for `/api/iam/groups/rebuild-ids` (POST, iam:write).
- Add `IAMHandler.RebuildGroupIDs` method with conflict/error mapping.
- Emit audit event `iam.group.rebuild_ids` for success/denied.
- Keep rebuild logic in existing `rebuildGroupIDsForAllTenants`.
