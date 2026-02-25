package api

import "kubedeck/backend/internal/core/audit"

var defaultAuditWriter audit.Writer = audit.NewMemoryWriter()

