# PRD: Invite Acceptance Membership Delta

## Problem / Background
- Accepting an invite previously changed invite status only; tenant membership was not created.

## Target Users
- Tenant admins sending invites.
- Invited users completing registration.

## Goals (In Scope)
- P0: Create or update tenant membership when invite is accepted.
- P0: Map `role_hint` to default tenant groups.

## Non-Goals (Out of Scope)
- Full external identity provisioning.
- New user directory service.

## Acceptance Criteria (AC)
- AC1: Given a valid invite, when accepted, then response includes created/updated membership.
- AC2: Given `role_hint=admin|owner|member`, then membership gets mapped tenant group (`admin|owner|viewer`).
