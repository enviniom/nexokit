# Sync Report: change-10-tenant-roles

Status: **PASS**

## Executive Summary

Archive-time sync fallback was explicitly approved by the user and completed successfully. Canonical OpenSpec specs now reflect tenant-scoped/root-only role behavior before archiving.

## Domains Synced

| Domain | Action | Result |
| --- | --- | --- |
| `roles` | Applied delta operations to existing canonical spec | PASS |
| `tenant-scoped-roles` | Created new canonical spec from full change spec | PASS |

## Requirement Operations

### roles

ADDED:
- Reserved slug validation

MODIFIED:
- Role CRUD API (replaced ~45 canonical lines with ~46 delta lines)
- Role seeds (replaced ~22 canonical lines with ~28 delta lines)
- System role protection (replaced ~22 canonical lines with ~34 delta lines)

REMOVED:
- None

### tenant-scoped-roles

ADDED / CREATED FULL SPEC:
- Company-scoped role model
- Reserved slug protection
- Tenant-isolated role queries
- Root role protection via API
- Role DTO includes company context
- Seed only root role globally

## Active Same-Domain Change Warnings

None. No other active (non-archive) change spec touches `roles` or `tenant-scoped-roles`.

## Destructive / Semantic Merge Approval

User explicitly approved archive-time sync fallback and semantic/destructive canonical spec replacements for `openspec/specs/roles/spec.md`, including root-only/no-`role_permissions` behavior and creation/sync of `tenant-scoped-roles`.

## Validation

Canonical specs were updated before archive. Available validation was performed after archival.
