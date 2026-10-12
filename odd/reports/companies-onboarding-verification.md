# Companies and onboarding: final verification review

## Outcome

Required source corrections are complete in four commits. Documentation now reflects
entity ownership and the onboarding provisioning slice. No blocking finding remains
in the supplied independent verification evidence; live-service and production HTTP
coverage remain explicit follow-up gaps, not completed verification.

This report uses the [task evidence](../tasks/companies-onboarding-verification.md)
and current source inspection. T4 is documentation-only and reruns no functional suite.
No auth/IAM behavior, schema, services or data changed. No push, PR or merge occurred.

## Review order and commits

| Commit | Completed unit | Review focus |
|---|---|---|
| `bacf005` | Persist companies request normalization | Pointer validation and domain regression cases |
| `6bce93b` | Companies entity model/DTO migration | Mapping, association, imports and consumer compatibility |
| `1d1fe85` | Reserve soft-deleted onboarding unique keys | Exact-key checks and atomic provisioning |
| `d67d5cc` | Local onboarding records and provisioning slice | Independent records, wiring and unchanged serialization |

Companies models/DTOs now live under `internal/modules/companies/core/<entity>/`.
Onboarding owns five local record packages under `internal/modules/onboarding/core/`;
its use case lives in `internal/modules/onboarding/slices/onboard_company/`.
Contracts/errors remain central. Internal aliases were removed and consumers migrated.
No new core/container/routes tests were added; regression tests are slice-owned.

## Decisions and compatibility

- Approved minimal `CompanyAssociation` preserves `CompanyDomain.Company`, preload,
  foreign-key semantics and the `companies` table without a bidirectional entity cycle.
  Current evidence: `internal/modules/companies/core/company_domain/model.go:7-21`.
- Pointer validation retains normalized domain mutations; company name/slug behavior
  remains unchanged. No additional normalization product rule was introduced.
- Onboarding local records remain independent of companies/IAM entity imports.
- Exact-key Unscoped availability queries match existing unconditional uniqueness;
  deleted keys remain reserved. No schema change or new key-reuse policy was added.
- HTTP fields, envelopes, optional values, collection shapes and tenant rules remain
  preserved within the verified scope. This is not full production HTTP verification.

## Prioritized findings

### Functional

**F1 — corrected: request normalization was lost.**
Historical `internal/modules/companies/core/dto.go:66,76,86,102` used value receivers.
Impact: normalized values did not persist on the caller request.
Observed evidence: normalization worker RED retained `  ACME.COM  ` in create/update;
GREEN passed after pointer receivers, including alternate normalized input,
name/slug preservation, empty-domain and primary-redirect cases.
Minimum remedy/status: pointer validation, completed in `bacf005`; current domain
validation is `internal/modules/companies/core/company_domain/dtos.go:36,51`.

**F2 — corrected: availability disagreed with deleted-key reservations.**
Historical onboarding `queries/check_domain_available.go:10-14` and
`queries/check_slug_available.go:10-14` used scoped availability checks.
Impact: reserved unique keys reached provisioning instead of early conflict handling.
Worker RED: four deleted-key cases returned nil; second RED reached email hashing.
GREEN: exact-key Unscoped checks passed all four cases, prevented hashing and preserved
six provisioning table counts. Injected late hashing failure proved transaction rollback.
Current evidence: `internal/modules/onboarding/queries/check_email_available.go:10-17`;
`internal/modules/onboarding/slices/onboard_company/service_test.go:255,307`.
Minimum remedy/status: three narrow reservation checks, completed in `1d1fe85`.

### Normative

**N1 — corrected: entity ownership and mapping layout drift.**
Historical `internal/modules/companies/model.go:5-23`, `dto.go:3-11` and onboarding
`core/dto.go:1-26` were flat/alias surfaces; those paths are no longer current sources.
Impact: ownership obscured and DTO placement diverged from architecture §4.
Evidence: current entity-owned records/DTOs, migrated consumers and central contracts.
Minimum remedy/status: internal migration without aliases, completed in `6bce93b`
and `d67d5cc`. Structural moves have no meaningful pre-implementation behavioral RED;
focused existing and added serialization/association tests supply compatibility evidence.

**N2 — corrected: stale documentation pointers.**
`docs/README.md:23` still described legacy flat Companies/onboarding layout;
`docs/references/infrastructure/domains.md:22` referenced removed companies `core/model.go`.
Impact: misleading current-state reading path and source lookup.
Evidence: current architecture §4 and entity/slice source files exist.
Minimum remedy/status: exact sentence/source replacements only, completed by T4.

### Optional / nonblocking

**O1 — deferred: generated-public-ID unique-error classification.**
Historical `internal/modules/onboarding/onboard_company/repository.go:132-135`
(now under `slices/onboard_company/`) raised a generic classification concern.
Impact: possible misleading classification outside the reproduced reservation cases.
Evidence: task explicitly records this as unproven, not a demonstrated new defect.
Minimum remedy/status: reproduce a concrete collision before proposing a narrow change;
no new product requirement or blocking correction is claimed.

**O2 — deferred: import grouping/status-literal consistency.**
`internal/modules/companies/core/company_domain/dtos.go:67-68` retains status literals.
Impact: maintenance consistency only; supplied independent review found no blocker.
Minimum remedy/status: optional scoped cleanup, not part of required source work.

**O3 — conditional: public host resolver/cache limitations (not a current private-route defect).**
`internal/modules/companies/resolver.go:24-27` does not explicitly filter joined
`companies.deleted_at`; `internal/middleware/tenant.go:104-158,232-252` has closure-local
public lookup caches without synchronization/invalidation. Impact if public tenant routes
are mounted: deleted-company host resolution, stale lookups and concurrent-map risk.
Read-only evidence: `internal/app/container.go:68-75` mounts private tenant guards only;
`PublicTenant` is not mounted. Private header lookup uses GORM company soft-delete scope
and intentionally is not the public active-domain gate. Minimum recommendation: before
public mounting, establish required revocation semantics and add focused deleted-company
and concurrent-cache regression coverage; no new product policy or dormant behavior
change introduced here. This is a documented conditional risk, not verified production safety.

## Verification evidence and limits

Independent verifier reported exit 0 for fresh companies/onboarding modules,
helpers/fixtures/integration, `go vet ./...`, `gofmt -l .`, working and cached
`git diff --check` / `git diff --cached --check`, and `make check-module-errors`.
Worker RED/GREEN evidence above is attributed to workers, not independent reruns.

`unshare --user --map-root-user --net go test -count=1 ./...` exited 0:
70 packages passed, 29 had no tests, zero failures. Ten isolated skips comprised
six Redis cache cases, one Redis rate-limiter case and three PostgreSQL CLI cases:
`TestMigrateCommand_Up`, `TestCreateRootCommand_IdempotentRealDB`,
`TestMigrateCommand_DBConnect`. Live Redis/PostgreSQL/migrations were NOT verified.

`internal/modules/onboarding/slices/onboard_company/wiring_test.go:24-44` exercises
real container/Register SQLite wiring with injected fake hasher and guard capabilities.
It does not verify production authentication/middleware. Companies production HTTP
wiring remains a gap. SQLite record assertions do not establish PostgreSQL migration fidelity.

## Verifiable correction plan and follow-up boundaries

| Unit | Status/checks | Rollback boundary |
|---|---|---|
| Normalization | Complete; observed worker RED/GREEN and independent focused checks | `bacf005` behavioral unit; preserve later entity layout |
| Companies ownership | Complete; module/integration, association/serialization, vet/guard | `6bce93b` cohesive types/imports/consumers unit |
| Reservations | Complete; four conflicts, early/no-write and late rollback evidence | `1d1fe85` exact-key checks and regressions |
| Onboarding layout | Complete; local records, slice tests and isolated suite | `d67d5cc` cohesive layout/consumer unit |
| Documentation | Exact stale references corrected; path inspection and diff check | Only the three T4 documentation files |

Follow-ups are separate: disposable PostgreSQL migration fidelity and production HTTP
wiring coverage. No live-service action or additional product behavior is authorized here.
Review risk is concentrated in import/type migration workload and live-service gaps.
T4 source/link targets were read successfully; `git diff --check` exited 0.
Parent read the report and replacement pointers; passive native assessment requires
structural readback only, with no additional functional suite. Final scope scan found
no auth/IAM/migration changes. Parent owns task reconciliation and documentation commit.
Shutdown was requested conditionally; parent decides only after saved report/commits
and no running agents. This report does not claim shutdown.
