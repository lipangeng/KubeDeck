# Workflow: bugfix_hotfix (Supervisor)

Phases:
1) PM or DM -> bug statement + expected behavior (mini-AC)
2) QA -> reproduction steps + minimal regression scope
3) SA (if needed) -> decision note if architecture impacted
4) FE/BE -> fix + add test
5) OBS -> ensure logs/metrics for detection (minimal)
6) SEC -> quick check (auth/data exposure) if security-adjacent
7) DE -> runbook smoke verification
8) DOCS -> changelog/known issue update if user-facing
9) DM -> release note + rollback note

Gates:
- Must have reproduction + fix verification
- Must have smoke verification
- If security-adjacent: SEC quick review required
