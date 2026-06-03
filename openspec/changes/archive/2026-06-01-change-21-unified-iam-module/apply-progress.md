## Apply Progress

**Change**: 21-unified-iam-module
**PR Slice**: PR 5 — app wiring + integration finalization (Phase 5 tasks 5.1–5.10)
**Mode**: Standard
**Persistence**: openspec

### Completed Tasks (cumulative)

- [x] Phase 1 tasks 1.1–1.9 (from PR1)
- [x] 2.1 users container + routes
- [x] 2.2 list_users slice
- [x] 2.3 create_user slice
- [x] 2.4 view_user slice
- [x] 2.5 update_user slice
- [x] 2.6 delete_user slice
- [x] 2.7 change_user_password slice
- [x] 2.8 assign_role_to_user slice
- [x] 2.9 toggle_user_status slice
- [x] 3.1 roles container + routes
- [x] 3.2 list_roles slice
- [x] 3.3 list_selectable_roles slice
- [x] 3.4 view_role slice
- [x] 3.5 create_role slice
- [x] 3.6 update_role slice
- [x] 3.7 delete_role slice
- [x] 3.8 view_role_permission_catalog slice
- [x] 3.9 assign_permissions_to_role slice
- [x] 4.1 permissions container + routes
- [x] 4.2 list_permissions slice
- [x] 4.3 view_permission slice
- [x] 4.4 update_permission slice
- [x] 4.5 internal resolve_auth_user slice
- [x] 4.6 internal resolve_user_permissions slice
- [x] 4.7 internal sync_permissions slice
- [x] 4.8 internal resolve_role_by_slug slice
- [x] 4.9 internal list_all_permissions slice
- [x] 5.1 app container uses `IAM *iam.Container`
- [x] 5.2 `RegisterModules` mounts IAM only via `iam.Register(...)`
- [x] 5.3 auth middleware lookup adapter delegates to IAM resolver
- [x] 5.4 role resolver adapter delegates to IAM role-by-slug contract
- [x] 5.5 bootstrap permission sync delegates to IAM sync contract
- [x] 5.6 app wiring test verifies 19 IAM endpoint paths are mounted
- [x] 5.7 app wiring test verifies only IAM route set is mounted (legacy mounts removed)
- [x] 5.8 adapter contract tests verify auth/authz wiring delegates to IAM
- [x] 5.9 legacy module compile check passes
- [x] 5.10 full test suite + IAM import boundary checks pass

### Tests Executed

- `go test ./internal/modules/iam/...` — passes (users + roles + permissions + internal slices)
- `go test ./...` — all pass, 0 failures
- `go build ./internal/modules/users/... ./internal/modules/roles/... ./internal/modules/permissions/...` — passes
- `go build ./...` — passes
- `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./internal/modules/iam/...` — no imports from legacy `users/roles/permissions/companies`

### Verification

- **Verdict**: PASS
- **Report**: `openspec/changes/21-unified-iam-module/verify-report.md`
- **Warnings**: None.

### Notes

- Query boundary rule preserved: no speculative files were added in `internal/modules/iam/queries/`.
- Single-use data access remains inside IAM role shared repository; no speculative shared query files were introduced.
- Maintainer query rule preserved: no new speculative shared query files were created; single-use SQL remains in owning repositories.
- Legacy modules remain on disk and compile; app runtime wiring now uses IAM as the primary users/roles/permissions boundary.
