# Workflow: iterative_feature (Supervisor)

Phases (delta-friendly):
1) PM -> PRD delta (scope/AC/edge)
2) UI -> UI spec delta
3) BE -> API contract delta
4) SA -> ARCH_SPEC delta + ADR if decision changes
5) DM -> TECH_PLAN delta + risks
6) FE/BE implement
7) QA focused regression + smoke
8) DE update demo runbook
9) SEC delta review (only impacted areas) + permission updates
10) OBS delta spec (new metrics/logs/alerts)
11) PERF delta plan (only if impacted path)
12) DOCS delta docs pack
13) DM release notes delta

Gates:
Same as greenfield; allow deltas if traceable.
