# Tech Plan Delta

- Implement `groupIDForTenantName(tenantID, groupName)` using canonical-name hash.
- Keep per-tenant duplicate-name protection and add patch duplicate conflict guard.
- Add `rebuildGroupIDsForAllTenants` to rebuild groups and rewrite membership refs.
- Add rollback-on-conflict by restoring in-memory snapshots.
