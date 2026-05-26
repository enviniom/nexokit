# Archive Report: Company Onboarding

## Status

Archived

## Date

2026-05-26

## Summary

Closed `change-13-company-onboarding` after reconciling the already-landed implementation with OpenSpec/SDD artifacts.

## Canonical Specs Updated

- `openspec/specs/company-onboarding/spec.md` — added root-only onboarding, transactional provisioning, role permission assignment, and disabled direct company creation.
- `openspec/specs/companies-crud/spec.md` — removed direct `POST /api/v1/companies` creation contract and documented that direct creation returns 404.
- `openspec/specs/permissions/spec.md` — added automatic assignment of newly synced permissions to existing tenant admin roles.
- `openspec/specs/roles/spec.md` — documented admin role permission assignment lock.

## Verification

- `go test ./...` passed.
- `go build ./...` passed.
- LSP diagnostics had no blocking findings.

## Archive Location

`openspec/changes/archive/2026-05-26-change-13-company-onboarding/`
