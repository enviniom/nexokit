# IAM documentation compliance

## Objective and authorization
Implement the approved review plan: protect own HTTP contracts and reorganize IAM models/DTOs by entity with DTO-owned mapping, preserving behavior.

User approved the full plan and every necessary IAM module file. User explicitly approved the permission contract-reference update, two import-only repairs in tests/helpers/database_test.go and internal/app/container_test.go, and bounded Python edits for T4. Git mutations remain unauthorized: no branch changes, staging, commits, push or publication. Work-unit commits remain pending explicit authorization; checkmarks below mean implementation and verification, not committed delivery. Auth remains closed at user-supplied commit 11527e1; no Auth business changes.

## Constraints and final architecture
- Single writer. Entities: core/user, core/role, core/permission, each with model.go/dtos.go; no compatibility aliases.
- Root core retains module contracts/errors/constants and reserved-role identity policy.
- Entity dependency direction user -> role -> permission; no entity imports parent core.
- No new tests under core, containers or routes. Mapping/validation/serialization coverage lives in slices or queries/integration; existing core/errors_test.go unchanged.
- Existing JSON fields, public numeric foreign/audit fields, omitempty, nil/empty shapes, five legacy table names/GORM tags, preloads, tenant policies, guards and permission sync preserved.
- Own 400/422 envelopes and zero service calls protected; no invented validation or Gin-parser tests.
- No Auth normalization/session-policy, IAM business, schema or query redesign.

## Tasks
- [x] T0: Map consumers and model dependencies. Initial capped exploration followed by uncapped parent/writer reference scans, including aliases and outside-module consumers. Commit pending authorization.
- [x] T1: Protect own HTTP contracts and response safety. 13 handler test files, +506/-29 lines; all nine JSON-binding handlers inspected. Missing request 400/422 envelopes/field errors and zero-service-call checks added; success safe fields, optionals, nil/empty lists/groups/catalog/selection protected. Writer baseline/post-edit suites and independent 20-package focused verification passed. Commit pending authorization.
- [x] T2: Migrate permission model/DTOs and mapping. List/view/update use DTO FromModel; root catalog contract reference updated. Validation/table coverage relocated outside core. Missing FromModel compile RED observed, same focused test GREEN. Independent IAM/helpers/integration, global compile, mapper spot, formatting/diff passed. No aliases/stale root permission references. Commit pending authorization.
- [x] T3: Migrate role models/DTOs/catalog/assignment. All consumers migrated, root duplicates removed; sorting/selection policy remains services, DTO APIs own projection. View's pointer-to-empty company public ID versus omission in other operations preserved. Table/identity/validation coverage relocated. Mapper API RED then GREEN observed. Independent focused IAM/CLI/app/seeds/helpers/integration, all-package compile, 3 mapper tests/7 response subtests, formatting/diff passed. Commit pending authorization.
- [x] T4: Migrate user model/DTOs and six mappers/all consumers. RoleName still uses Role.Name; list values versus five pointer return paths preserved. Missing mapper RED then GREEN observed; safety/optionals/audit/reset tests passed. Root flat model.go/dto.go and emptied model/DTO/role-identity tests removed after relocation. Uncapped 459-Go-file scan no stale references/aliases/mappers. Independent IAM/CLI/helpers/integration, all-package compile, fresh mapper spot, formatting (154 changed/new Go files), diff passed. Commit pending authorization.
- [x] T5: Align six docs and independently verify full candidate. docs/README.md, architecture_and_vsa.md, api_validation_and_errors.md, testing_and_quality.md, system-map.md and normative_traceability_matrix.md aligned with current IAM and test ownership. Parent corrected stale Auth test-location wording only. Final independent functional and documentation checks passed; commit pending authorization.

## Verification evidence
- go test ./... -count=1: PASS, 71 tested packages / 20 no-test packages, no failures.
- go vet ./...: PASS, no output.
- go test ./tests/integration/... -count=1 -v: PASS, 6 top-level test functions / 32 runs passed, no skips or failures. SQLite test environment.
- Fresh three mapper slice suites: PASS (users/view_user, roles/view_role, permissions/list_permissions).
- Parent fresh go test ./internal/modules/iam/users/view_user -run '^TestUserResponseFromModel$' -count=1: PASS.
- gofmt -l across 154 existing changed/new Go files, deleted paths excluded: clean.
- git diff --check and git diff --cached --check: PASS; final documentation supplement diff check also PASS.
- Independent docs supplement: six diffs consistent, 11 cited code paths exist, local links/anchors resolve, no stale IAM-flat statements or contradictions. No Go changes after full functional checks, so full suite not repeated for final docs-only edits.
- Final read-only Git checks: branch main, HEAD b6f305b unchanged, no staged changes. Tracked workload: 151 files +1466/-1806; 15 untracked files (6 entity source files, 8 test files, this task document). Large review surface reflects relocation/direct-reference updates; review by T1/permission/role/user/docs units rather than as one undifferentiated diff.
- Native RDD off. ASSESS unassessable due undeclared untracked scope; conservative independent verification applied. No native review lineage started.

## Incidents and deviations
Role writer timed out after partial consumer edits. Read-only incident verification identified compile failures; bounded recovery fixed contract/association references and missing imports. Two outside-surface imports were explicitly approved and repaired; all final checks passed. No Git/config recovery or rollback.

T4 queries/user_model_test.go was created before its exact path was announced, within user-authorized IAM scope; writer disclosed this procedural deviation. No outside-IAM scope expansion. Existing-behavior characterization and passive docs had no meaningful RED; none manufactured. New mapping APIs had observed compile RED/GREEN.

## Limitations and next step
No live-server, production PostgreSQL locking/dialect or production Redis verification. Historical Redis limiter test-isolation failure was not observed in final checks; no Redis/config changes or retry masking. External downstream Go consumers not verified.

Prior entity migration completed and verified; evidence above predates the user's subsequent slice move and is historical, not proof of current compilation. User now moved all IAM use cases under slices/{users,roles,permissions,internal} and authorized validation and necessary adjustments. Preserve this entity grouping. New human-owned staged snapshot must remain byte-identical: ec52d73a5a19a695e4698b2fb787f786181ec2ae64ac1d30e88c88191150c2ec (git diff --cached --binary | sha256sum). Git mutations remain unauthorized.

## Slice relocation follow-up
- [x] T6: Inspect live moved tree and baseline compile. Old container/routes/integration imports and missing internal service imports found. Baseline IAM/integration and global compile failed. Go forbids root iam from importing slices/internal; staged hash unchanged. Stale CodeGraph map is not authoritative; current filesystem already contains moved usecases.
- [x] T7: Fix direct imports/wiring/current documentation preserving moved files. Status: writer complete and focused checks passed; commit pending authorization. Imports/routes/3 integration consumer imports repaired, slices/internal service imports restored, five docs updated. Focused IAM/integration and all-package compile passed, formatting/diff clean; staged checksum identical. Add a small slices-level capability assembler to import slices/internal legally and expose root core contracts; root container delegates capability construction. No aliases, business changes, container/routes tests or reverse root imports.
- [x] T8: Independently verify relocated candidate, docs and staged-index preservation. Status: independently verified, commit pending authorization. ASSESS unassessable due untracked scope, RDD off; independent high-risk fallback applied. Full tests passed (71 tested / 21 no-test packages), vet clean, IAM focused tests passed, integration passed (6 top-level / 26 subtests / 32 runs, no skips). Formatting clean across 185 existing changed/new Go files; working/cached diff checks passed. Staged checksum unchanged before/after and parent readback: ec52d73a5a19a695e4698b2fb787f786181ec2ae64ac1d30e88c88191150c2ec. HEAD b6f305b0c08b93bebccf66d34ba8b6038724ba94 and main unchanged. Of 144 relocated files, 140 match staged old-path bytes; remaining four change only imports. Five docs checked with eight existing code references. Parent fresh moved user mapper test passed. No aliases, cycles, duplicate construction, new forbidden tests, route/guard changes or blockers. No live-server/production PostgreSQL/production Redis/external downstream verification; local Redis limiter test passed with its test-owned namespace. New/moved files remain unstaged/untracked intentionally.

Next: User review of the committed IAM work. Auth and unrelated scope remain closed. Push, PR creation and merge remain unauthorized. Historical pending-commit and index-preservation statements above describe the earlier authorization boundary; the delivery record below supersedes them.

## Authorized delivery record
User explicitly authorized necessary commits after the slice relocation was independently verified. The user-staged snapshot contained all 209 intended paths, with no unstaged or untracked files; non-IAM consumers were import/type adaptations only. Independent precommit full tests and vet passed again. RDD mode remained globally off; no native review lineage was started.

- Feature branch: `refactor/iam-entity-slices`, created from `b6f305b0c08b93bebccf66d34ba8b6038724ba94`; main unchanged.
- T0–T8 implementation/test/documentation evidence: `712e304391f9abd125428dcf6191205bc89cc7cf` — `refactor(iam): organize entity models and vertical slices`.
- Boundary: models/DTO mapping, all direct consumers, slice move/wiring, characterization coverage and current documentation form one compile-coupled unit. Artificial entity or file-type splits would leave unresolved imports; the existing staged snapshot was committed intact rather than reconstructed.
- Review workload: 209 files, +2949/-2165 versus the base, including renames; oversized candidate explicitly surfaced. Review entity mapping, HTTP contracts/tenant policy, internal assembly, then docs. This commit is not a small PR slice; any later PR strategy requires a separate decision.
- Runtime evidence: full `go test ./... -count=1` and `go vet ./...` passed for the exact index/worktree snapshot; previously recorded integration tests passed with SQLite. No production PostgreSQL/Redis/live-server evidence.
- Rollback boundary: revert the coherent IAM commit and its consumer/doc adaptations together; do not revert individual entity models while retaining consumers. No rollback performed.
- Tracking-only follow-up commit records this authorization and commit identity; no source behavior changes and no meaningful test-first RED for this passive record.
- No push, PR or merge performed. Companies review is proposal-only; implementation not authorized.

## Persistence
Repository locator: odd/tasks/iam-documentation-compliance.md. Engram project nexokit, topic odd/iam-documentation-compliance/tasks, observation 4076; mirror read back after updates.
