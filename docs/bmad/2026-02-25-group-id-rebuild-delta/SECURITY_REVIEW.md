# Security Review Delta

- Tenant-scoped IDs reduce cross-tenant collision and confused-deputy risks.
- Patch duplicate guard prevents hidden alias takeover through rename.
- Rebuild conflict detection prevents unsafe merges of canonical duplicate groups.
