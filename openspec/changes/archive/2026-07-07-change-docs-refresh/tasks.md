# Tasks: Documentation Refresh

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 600-9500 (raw); 600-750 with `--find-renames` |
| 400-line budget risk | High — waived by approved `size:exception` |
| Chained PRs recommended | No — exception approved for single PR |
| Suggested split | N/A — single PR under approved exception |
| Delivery strategy | single-pr-default |
| Resolved workload decision | `size:exception` approved by user; single PR accepted |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High — waived by approved `size:exception`

> Archive moves (3 arch docs ≈2,037 lines + 2,733-line changelog) inflate `git diff --stat`. The `size:exception` workload decision is already approved; there is no pending apply decision.

### Suggested Work Units

| Unit | Goal | PR |
|------|------|----|
| 1 | Docs inventory + Makefile/CLI source-of-truth | PR 1 |
| 2 | Root `README.md` rewrite (≤150 lines) | PR 1 |
| 3 | `docs/README.md` index (~80 lines) | PR 1 |
| 4 | `docs/cli.md` expansion | PR 1 |
| 5 | `docs/modules.md` link fix | PR 1 |
| 6 | English normalization of active docs | PR 1 |
| 7 | `docs/architecture.md` consolidation (~150 lines) | PR 2 |
| 8 | `docs/deployment.md` production guide (~200 lines) | PR 2 |
| 9 | `docs/starter-template.md` (~80 lines) | PR 2 |
| 10 | Move `docs/prompts/*` → `./prompts/` | PR 2 |
| 11 | Archive 3 arch docs + changelog; new `archive/README.md` | PR 2 |
| 12 | Verification suite | Both |

## Phase 1: Inventory (PR 1)

- [x] 1.1 Build `docs/` inventory at `notes/inventory.md` (paths, line counts, language, status).
- [x] 1.2 Parse `Makefile` and `cmd/nexokit/*.go`; record canonical commands in `notes/cli-source-of-truth.md`.

## Phase 2: Index and Root README (PR 1)

- [x] 2.1 Rewrite `README.md` (≤150 lines): stack table, architecture paragraph, folder map, Makefile quick path, prod pointer, `docs/` link.
- [x] 2.2 Create `docs/README.md` index (~80 lines): quick paths, doc map, canonical links.
- [x] 2.3 Fix `docs/modules.md`: replace `openspec/core_context.md` and `openspec/specs/backend/...` with real paths.
- [x] 2.4 Expand `docs/cli.md`: `serve`, `migrate up|down|status|reset|create`, `seed`, `create-root`; drop "persistence-blocked" line.

## Phase 3: Architecture and Operations (PR 2)

- [x] 3.1 Create `docs/architecture.md` (~150 lines): entrypoints, `internal/app`, `internal/modules/{auth,companies,iam,onboarding}`, flow.
- [x] 3.2 Create `docs/deployment.md` (~200 lines): build, env, DB/SSL, migrations, seed/root, logging, reverse proxy/TLS, ops checklist.
- [x] 3.3 Create `docs/starter-template.md` (~80 lines): clone/fork, rename, env, module selection, run.

## Phase 4: Archive and Prompts (PR 2)

- [x] 4.1 Move `docs/nexokit-architecture*.md` (3) and `changes_sdd_nexokit_go_completo.md` to `docs/archive/`.
- [x] 4.2 Create `docs/archive/README.md` listing archived files; point to canonical sources.
- [x] 4.3 Move `docs/prompts/*` to `./prompts/`; exclude from `docs/README.md`.

## Phase 5: Verification (Both)

- [x] 5.1 Validate all `](...)` links in changed `.md` resolve.
- [x] 5.2 Cross-check commands in `README.md`, `docs/cli.md`, `docs/deployment.md` against `Makefile` and `cmd/nexokit/`.
- [x] 5.3 Grep active `README.md`/`docs/*.md` (excl. `archive/`, `prompts/`) for Spanish fragments.
- [x] 5.4 Run `git diff --stat`; verify diff with `--find-renames`. The `size:exception` workload decision is approved; reviewers should use the documented work-unit sections rather than the 800-line figure as a hard pass.
- [x] 5.5 Grep for stale `docs/prompts/` references; update or remove.

## Post-review remediation notes

Fixed findings from the fresh-context review of `change-docs-refresh`:

- `README.md` and `docs/deployment.md` now state that `nexokit seed` requires the Go toolchain on the seeding host (or a build/admin environment with DB access), because `internal/cli/commands/seed.go` runs seeds via a temporary `go run` runner.
- `docs/cli.md` no longer shows `--password ChangeMe123`; it uses a placeholder and warns against passing production passwords via CLI args.
- `docs/deployment.md` and `docs/cli.md` label `migrate down`/`migrate reset` as Danger Zone, default-prohibited in production, and requiring backup + explicit approval.
- `docs/archive/changes_sdd_nexokit_go_completo.md` link labels updated from `docs/prompts/...` to `prompts/...`; `docs/archive/README.md` documents the historical label convention.
- `README.md` and `docs/architecture.md` no longer describe `internal/platform/contracts` as the cross-module pattern; they now describe module-local contracts in `internal/modules/<module>/core/contracts.go` wired in `internal/app/container.go`.
- `docs/cli.md` added `config` and `status` sections with purpose, env needs, invocation, and expected output scope.
- `docs/deployment.md` added concise alerting thresholds and rollback/fix-forward criteria.
- `tasks.md` no longer treats 800 changed lines as a hard pass; it records the user-approved `size:exception`.

## Second remediation (re-review)

- [x] R.9 `README.md` production command sequence now includes `./bin/nexokit seed` before `./bin/nexokit create-root`, and the Go-toolchain requirement is stated clearly.
- [x] R.10 `docs/cli.md` and `docs/deployment.md` now state that `migrate down`/`migrate reset` are forbidden by operational policy in production unless explicitly approved; they no longer imply the CLI binary blocks these commands.
- [x] R.11 `tasks.md` guard lines and forecast updated to reflect the approved `size:exception` (no pending decision, `Chain strategy: size-exception`, 800-line figure waived); `design.md` review-size wording treats 800 lines as a guideline, not a hard gate.
- [x] R.12 `openspec/changes/change-docs-refresh/exploration.md` relative markdown links converted to plain path references so they are not broken from the OpenSpec folder.
- [x] R.13 `docs/archive/README.md` updated to reflect that archived prompt links were de-linked and that preserved historical examples remain.
- [x] R.14 `docs/README.md` CLI summary now includes `config` and `status`.
- [x] R.15 `docs/archive/changes_sdd_nexokit_go_completo.md` prompt markdown links de-linked and credential/token examples sanitized.
