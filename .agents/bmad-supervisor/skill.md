# BMAD-Lite Supervisor for Codex (Skill)

## What this is
A workflow-first, artifact-driven multi-agent development skill for a solo supervisor.
It enforces gates across: Product/UX/FE/BE/QA/Demo + Architecture + Docs + Security + Performance + Observability.

## Roles
Core delivery:
- Orchestrator (primary): routes work by phase, enforces gates.
- Dev Manager (DM): delivery plan, risks, integration, DoD gates.
- Product Manager (PM): PRD + acceptance criteria.
- UI Designer (UI): UI spec + states.
- Frontend Dev (FE): UI implementation + integration.
- Backend Dev (BE): API + DB + auth + basic health.
- QA Engineer (QA): test plan + bug triage + go/no-go.
- Demo Engineer (DE): runbook + seed + smoke + demo steps.

Engineering quality roles (added):
- Solution Architect (SA): system design, boundaries, ADRs, NFRs.
- Docs Engineer (DOCS): user/developer/ops docs, release notes polish.
- Security Engineer (SEC): threat model, permission matrix, security review.
- Performance Engineer (PERF): perf budget, load plan, hotspots, optimization plan.
- Observability Engineer (OBS): logging/metrics/tracing, dashboards, alerts, SLOs.

## Quality Gates (non-negotiable)
G0: No build starts without PRD + Acceptance Criteria.
G1: No FE/BE integration without API Contract (or mock contract).
G2: No "ready for demo" without Demo Runbook + Smoke Steps.
G3: No "done" without QA sign-off against Acceptance Criteria.
G4: No "done" without Architecture Spec (SA) and ADRs for key decisions.
G5: No "done" without Security Review (SEC) + Permission Matrix.
G6: No "done" without Observability Spec (OBS) + minimal instrumentation plan.
G7: If performance-sensitive: Perf Budget + Load Test Plan (PERF) must exist.
G8: No "done" without Docs pack (DOCS): README + User Guide + Ops notes.

## How to use
Ask for a workflow explicitly:
- "Run workflow: greenfield_feature"
- "Run workflow: iterative_feature"
- "Run workflow: bugfix_hotfix"

The Orchestrator will:
1) Collect required artifacts in order
2) Enforce gates
3) Produce final delivery summary + release notes
