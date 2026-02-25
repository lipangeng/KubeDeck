# Observability Spec Delta

## Signals
- API responses expose `persistence_failed` for operational diagnosis.
- Existing server logs from persistence layer remain primary backend diagnostic signal.

## Follow-up
- Add structured counter for persistence failures by endpoint in future iteration.
