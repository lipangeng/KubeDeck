# Workflow: greenfield_feature (Supervisor)

Phases:
1) PM -> PRD.md
2) UI -> UI_SPEC.md
3) BE -> API_CONTRACT.md
4) SA -> ARCH_SPEC.md + ADRs
5) DM -> TECH_PLAN.md
6) FE/BE -> Implement minimal runnable slice
7) QA -> TEST_PLAN.md + smoke suite
8) DE -> DEMO_RUNBOOK.md + demo checklist
9) SEC -> SECURITY_REVIEW.md + PERMISSION_MATRIX.md
10) OBS -> OBSERVABILITY_SPEC.md
11) PERF -> PERF_PLAN.md
12) DOCS -> DOCS_PACK.md
13) DM -> RELEASE_NOTES.md

Gates:
G0 PRD+AC
G1 API contract (or mock)
G4 Architecture spec + ADRs
G2 Demo runbook + smoke steps
G3 QA sign-off + AC coverage
G5 Security review + permission matrix
G6 Observability spec + minimal instrumentation
G7 Perf plan (always minimal; deeper only if sensitive)
G8 Docs pack complete
