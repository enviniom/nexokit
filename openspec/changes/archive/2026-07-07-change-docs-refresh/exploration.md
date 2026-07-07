# Exploration: change-docs-refresh

> **Goal**: Investigate the current `docs/` and `README.md` state for NexoKit, surface duplication and obsolescence, and recommend the shape of a top-level `docs/` index, the README layout, and the work-unit split that protects the 800-line review budget.
>
> This is exploration only. No code or doc edits are produced here; the proposal phase owns the implementation plan.

## Current State

### `README.md` (135 lines, May 26)

Located at the repo root. Covers:

- One-line product summary ("modular Go starter framework for building SaaS APIs").
- Prerequisites (Go 1.26+, Docker Compose, Make, `goose` install line).
- Quick Start with 5 numbered steps (env, docker compose up, `make migrate-up`, `make seed`, `make create-root`, `make dev`).
- Commands table covering the full `Makefile` target list.
- Pre-commit hooks section.
- Log Files section (gin.log, app.log, error.log).
- Shallow "Project Structure" tree of `internal/`.
- Short "Conventions" bullet list.

What is missing or weak:

- No link into `docs/`. The 14 markdown files in `docs/` are undiscoverable from the README.
- "Project Structure" tree stops at `internal/<top-level>`; it never mentions the four current modules (`auth`, `companies`, `iam`, `onboarding`) or the vertical-slice shape they use.
- No CLI section. `cmd/nexokit` is mentioned in the structure tree but the `nexokit` command (`serve`, `migrate up|down|status|reset|create`, `seed`, `create-root`, `make module|seed|migration`) is undocumented; the only place the CLI surface is described is `docs/cli.md` and it is not linked from anywhere.
- No "How to run in production" section. `make dev` is the only run command shown; `make build` produces `bin/api` and `bin/nexokit` but the README never says where those binaries go, how to launch them, or which env vars to set in prod.
- The Conventions section is a 4-bullet stub; the rich module/architecture contract lives in `docs/` and is unreferenced.
- "Tests" is absent. The Makefile exposes `test`, `test-unit`, `test-integration`, `test-coverage`, but the README never explains the unit/integration split or the CI workflow (`.github/workflows/ci.yml` runs `test`, `vet`, `fmt-check`, `module-errors-guard`).
- Tech stack is implicit; the README never spells out Go, Gin, GORM, PostgreSQL, Goose, slog, lumberjack, PASETO, argon2id, valkey/redis. The reader has to discover this from `go.mod`, `docker-compose.yml`, and the OpenSpec config.

### `docs/` inventory and condition

```
docs/
├── api-conventions.md               200 lines  May 20  (en)  CANONICAL
├── changes_sdd_nexokit_go_completo.md  2,733 lines  May 14  (en)  STALE / DUPLICATE
├── cli.md                            56 lines  May 16  (en)  CANONICAL but shallow
├── error-handling.md                205 lines  Jul  3  (en)  CANONICAL
├── module-error-conventions.md      109 lines  Jul  4  (en)  CANONICAL
├── modules.md                        80 lines  Jul  3  (en)  PARTIAL INDEX (only covers modules/)
├── modules/                          6 files  (en)  CANONICAL per-concern tutorials
│   ├── boundaries-and-dependencies.md
│   ├── module-structure.md
│   ├── queries-and-persistence.md
│   ├── testing.md
│   ├── validation-and-errors.md
│   └── vertical-slices.md
├── multitenancy.md                   64 lines  May 19  (en)  CANONICAL
├── nexokit-architecture.md          978 lines  Jul  3  (en)  STALE (pre-vertical-slice era)
├── nexokit-architecture-revised.md  572 lines  Jun  8  (es)  STALE + WRONG LANGUAGE
├── nexokit-architecture-v2.md       487 lines  Jun  8  (en)  STALE (pre-Change-15 architecture)
├── prompts/                          2 files  (es)  INTERNAL SDD PROMPTS
│   ├── _context.md                  236 lines  May 30
│   └── change_XX_vertical_slice.md   57 lines  Jul  3
├── request-flow.md                  217 lines  May 19  (en)  CANONICAL
├── testing.md                       192 lines  May 21  (en)  CANONICAL
└── vertical-slice-modules.md        226 lines  Jun  2  (en)  CANONICAL
```

Total: ~4,770 lines across 14 top-level files + 6 in `docs/modules/` + 2 in `prompts/`.

### Duplication and obsolescence candidates

1. **Three competing architecture docs** (`nexokit-architecture.md`, `nexokit-architecture-v2.md`, `nexokit-architecture-revised.md`) describe substantially the same folder map and import rules. None of them is the canonical reference today:
   - All three still describe a top-level `internal/domain/` package, but the repo has no such package — domain types live in `internal/modules/<m>/core/`. The canonical contract is `openspec/specs/vertical-slice-modules/spec.md` plus `docs/modules/module-structure.md` and `docs/modules/boundaries-and-dependencies.md`.
   - `nexokit-architecture-revised.md` is in Spanish ("Versión Optimizada") while the rest of the project docs and all OpenSpec specs are in English. It is a language outlier.
   - `nexokit-architecture-v2.md` is the shortest of the three and may be the most "current" predecessor, but it is still pre-vertical-slice-rules-audit (Change-15) and pre-Change-24.
2. **`docs/changes_sdd_nexokit_go_completo.md`** is a 2,733-line historical log of every SDD change in the project. That history already lives, in canonical form, under `openspec/changes/archive/2026-*` (27 archived change folders as of 2026-07-04). The doc is a parallel, older, and drift-prone copy of the same data.
3. **`docs/vertical-slice-modules.md`** is a complete per-concern tutorial on the vertical-slice rule. The same content is also split across `docs/modules/vertical-slices.md`, `docs/modules/module-structure.md`, `docs/modules/boundaries-and-dependencies.md`, `docs/modules/queries-and-persistence.md`, `docs/modules/validation-and-errors.md`, `docs/modules/testing.md`. `modules.md` is already an index over `modules/`, and `vertical-slice-modules.md` is mostly redundant with that index. It also notes itself as a "compatibility stub" in the `modules.md` table.
4. **`docs/modules.md` has dangling references**:
   - Line 5: `openspec/core_context.md` — this file does not exist (verified: no `core_context.md` anywhere under `openspec/`). The real source of non-negotiable decisions is `prompts/_context.md` (Spanish, internal SDD context), `openspec/config.yaml`, and the spec files themselves.
   - Line 5: `openspec/specs/backend/vertical-slice-modules/spec.md` — there is no `backend/` subdirectory under `openspec/specs/`. The actual spec is at `openspec/specs/vertical-slice-modules/spec.md`.
5. **Language mix**: 3 of 16 files in `docs/` are in Spanish (`nexokit-architecture-revised.md`, `prompts/_context.md`, `prompts/change_XX_vertical_slice.md`). The Spanish files are mostly internal SDD prompts, but `nexokit-architecture-revised.md` is exposed as project documentation. The user request explicitly asked for English-by-default project docs.
6. **`docs/cli.md`** says `create-root` is "persistence blocked until auth schema is wired" (line 20) and "the underlying storage persistence will be enabled by a future auth change" (line 56). Auth, IAM, and `create-root` have shipped (changes 03, 04, 21, 22, etc.) and `make create-root` is in the Makefile; the doc is stale.

### What is genuinely canonical today (do not duplicate)

- `docs/api-conventions.md` — DTO/envelope/query-param contract.
- `docs/error-handling.md` — `AppError` shape, helpers, `Wrap`, `ErrorLogger`, redaction.
- `docs/module-error-conventions.md` — per-sentinel table, paired with `core/errors_test.go`.
- `docs/modules.md` + `docs/modules/*` — the modular tutorial set, including the vertical-slice rule.
- `docs/multitenancy.md` — tenant field/scope rules.
- `docs/request-flow.md` — middleware chain and tenant resolution.
- `docs/testing.md` — project-wide test commands, integration setup, CI.
- `docs/vertical-slice-modules.md` — the long-form vertical-slice tutorial; safe to keep but should be demoted to a "deep dive" from `docs/modules.md` rather than a sibling.
- `docs/cli.md` — CLI command surface; needs the `create-root` line updated, but otherwise OK.

### What does not exist and should

- **No `docs/README.md` or `docs/index.md`**. `modules.md` is a partial index for the modules subdirectory only. There is no top-level entry point that says "if you want to understand the project, read this first; if you want to add a feature, read this; if you want to test, read this".
- **No "how to run in production" doc**. `make build` + binary execution + env-var expectations are not documented anywhere.
- **No deployment / observability / env-vars reference**. `.env.example` is the only inventory of env vars and is not indexed.
- **No "where things live" map** that ties the four current modules to the per-concern tutorials. A reviewer landing on the repo has to know that `internal/modules/iam/` follows the `docs/modules/vertical-slices.md` rule.

### What `prompts/` is and why it is not user-facing

`prompts/_context.md` and `prompts/change_XX_vertical_slice.md` are SDD working prompts used during change planning. They are not project documentation; they are an internal authoring tool. They happen to be in `docs/` because the project originally did not separate "docs for users" from "docs for the AI author". The user request says "improve project documentation under `docs/`", which implies that these internal prompts should be considered separately and probably moved or excluded from the public index.

## Affected Areas

| Path | Why it is affected |
|---|---|
| `README.md` | The user explicitly asked for a rewrite to best practices: stack, folder map, dev/prod run, link into `docs/`. |
| `docs/README.md` (new) | Will be the top-level index. New file. |
| `docs/nexokit-architecture.md` | Stale (pre-vertical-slice, references non-existent `internal/domain/`). Candidate for deletion or demotion to `docs/archive/`. |
| `docs/nexokit-architecture-v2.md` | Stale. Same deletion candidate. |
| `docs/nexokit-architecture-revised.md` | Stale, Spanish, language outlier. Strongest deletion candidate. |
| `docs/changes_sdd_nexokit_go_completo.md` | 2,733-line duplicate of `openspec/changes/archive/`. Move to `docs/archive/` or delete. |
| `docs/vertical-slice-modules.md` | Mostly duplicated by `docs/modules/*`. Either delete, demote, or repoint links from `docs/modules.md`. |
| `docs/modules.md` | Has two broken cross-references (`openspec/core_context.md`, `openspec/specs/backend/...`). Fix or rewrite. |
| `docs/cli.md` | Stale line about `create-root` being persistence-blocked. Update. |
| `prompts/_context.md`, `prompts/change_XX_vertical_slice.md` | Not user-facing; either keep hidden from the index or move out of `docs/`. |

## Approaches

### Approach A — "Index first, then trim" (single PR, indexed by `docs/README.md`)

1. Create `docs/README.md` as a short landing page with quick paths (add a module, fix a bug, run tests, deploy, understand the architecture).
2. Rewrite `README.md` to best practices: stack table, one-paragraph architecture summary, folder map, dev commands, prod commands, link to `docs/`.
3. In the same PR: delete or move the three stale architecture docs and the `changes_sdd_nexokit_go_completo.md` to `docs/archive/` (or a sibling `openspec/changes/archive/`-like folder under `docs/`).
4. Fix dangling references in `docs/modules.md` and `docs/cli.md` (`create-root` line).

- Pros: one pass, one PR, immediate cognitive load reduction, the index points at a clean tree.
- Cons: large diff (3 deletions of 500–2,700 lines plus rewrites); risks overshooting the 800-line review budget if not split.
- Effort: Medium.

### Approach B — "Two chained PRs: index + README, then archival"

PR 1 (`change-docs-refresh` part 1):
- Add `docs/README.md` index.
- Rewrite `README.md`.
- Fix `docs/cli.md` `create-root` line.
- Fix dangling references in `docs/modules.md`.

PR 2 (`change-docs-refresh` part 2, or a follow-up `change-docs-archive`):
- Move stale architecture docs to `docs/archive/`.
- Move `docs/changes_sdd_nexokit_go_completo.md` to `docs/archive/` (or delete).
- Resolve `docs/vertical-slice-modules.md` duplication.
- Decide what happens to `prompts/`.

- Pros: keeps each PR under the 400-line review budget, allows the team to merge the index first so the team can already navigate the docs while the archival lands.
- Cons: there is a window where `docs/` is "indexed but still messy"; the team must review two PRs.
- Effort: Medium.

### Approach C — "Rewrite everything in one large PR, no archive step"

1. Replace `docs/` wholesale with a clean tree and a fresh `docs/README.md` index.
2. Rewrite `README.md`.
3. Do not preserve any old architecture doc.

- Pros: cleanest end state, no legacy file names.
- Cons: destroys history, highest review load, biggest risk of losing useful content, breaks any external link to the old paths.
- Effort: High.

## Recommendation

**Approach B (chained PRs)**, for three reasons:

1. The 800-line review budget is the binding constraint. The README rewrite alone plus the index plus the dangling-reference fixes will already be 200–300 net additions. Adding 3 stale-doc deletions and a 2,733-line history move in the same PR would push it over.
2. The "index" half is independently useful and reviewable on its own; reviewers can confirm the index points at the canonical docs without simultaneously ruling on what to do with the obsolete ones.
3. The "archive" half is a deletion PR, which is exactly the kind of change that benefits from being reviewed in isolation — any reviewer concern about losing a paragraph of architecture can be addressed without blocking the index.

Concrete shape of PR 1 (`change-docs-refresh`):

- New: `docs/README.md` (top-level index, ~80 lines, scans in 30 seconds).
- Rewritten: `README.md` (rewritten to best practices, ~110–130 lines, with a stack table, architecture one-liner, folder map, dev / test / prod command tables, link to `docs/`).
- Fix: `docs/cli.md` (update the `create-root` line, ~3 lines).
- Fix: `docs/modules.md` (replace the dangling `openspec/core_context.md` link with the real source and fix `openspec/specs/backend/...` to `openspec/specs/...`).

Concrete shape of PR 2 (suggested name `change-docs-archive`):

- Move `docs/nexokit-architecture.md`, `docs/nexokit-architecture-v2.md`, `docs/nexokit-architecture-revised.md`, `docs/changes_sdd_nexokit_go_completo.md` to `docs/archive/` with a one-line note in `docs/archive/README.md` explaining why they were archived.
- Decide and act on `docs/vertical-slice-modules.md` (likely keep, demote to a "deep dive" sub-page from `docs/modules.md`).
- Decide and act on `prompts/` (likely keep as-is but exclude from the public index, or move to a `tools/sdd-prompts/` folder outside `docs/`).
- Net effect of this PR is mostly file moves plus a 5–10 line `docs/archive/README.md`.

### What the `docs/README.md` index should contain

Following the cognitive-doc-design skill (lead with the answer, progressive disclosure, recognition over recall, signposting):

```markdown
# NexoKit documentation

> One sentence: where to find what you need.

## Quick path
1. New here? Read `README.md` and `docs/architecture.md`.
2. Adding a module? Read `docs/modules/module-structure.md` and `docs/modules/vertical-slices.md`.
3. Fixing a bug? Read `docs/request-flow.md` and the relevant per-concern tutorial under `docs/modules/`.
4. Running tests? Read `docs/testing.md`.
5. Running in production? Read `docs/deployment.md` (added in this change).

## Documentation map
| If you want to... | Read this |
|---|---|
| Understand the architecture | `docs/architecture.md` (consolidated in this change) |
| Add or change a module | `docs/modules.md` → per-concern tutorials |
| Define or review an API | `docs/api-conventions.md` |
| Handle errors | `docs/error-handling.md` and `docs/module-error-conventions.md` |
| Understand tenant isolation | `docs/multitenancy.md` and `docs/request-flow.md` |
| Use the CLI | `docs/cli.md` |
| Write or run tests | `docs/testing.md` |
| Run in production | `docs/deployment.md` (added in this change) |
| Find historical architecture notes | `docs/archive/` (preserved for context) |
```

A consolidated `docs/architecture.md` is the natural place to land the "what used to live in three arch docs" content. It should be a short (~150 lines) one-pager that links into the per-concern tutorials for depth. **It is not strictly required by the user request, but the index is much weaker without it** — without a single architecture page, the three deleted architecture docs have no obvious successor and reviewers will keep asking "where is the architecture doc?". This is the strongest argument for making PR 2 include a 150-line `docs/architecture.md` consolidation, even though that pushes the PR slightly over the comfort line.

### What `README.md` should contain vs. what belongs in `docs/`

Following best practices for a Go SaaS starter README and the cognitive-doc-design skill:

| Section | In `README.md`? | In `docs/`? | Rationale |
|---|---|---|---|
| One-line description | Yes | No | README's job. |
| Tech stack table | Yes | No | Buyers / contributors scan it in 5 seconds. |
| One-paragraph architecture summary | Yes (1 paragraph) | Yes (deep dive in `docs/architecture.md`) | README hooks the reader; `docs/` carries the depth. |
| Folder map | Yes (one tree, top-level dirs only) | Yes (annotated, in `docs/architecture.md`) | README gives a 30-second map; `docs/` annotates each subtree. |
| Quick start (dev) | Yes | No | README's job. |
| How to run in production | Yes (brief, with link) | Yes (full env-var / binary / process doc in `docs/deployment.md`) | README gives the command; `docs/` gives the contract. |
| Test commands | Yes (one-liner per target) | Yes (deep dive in `docs/testing.md`) | Reader can find `make test` instantly; integration setup lives in `docs/`. |
| CLI summary | Yes (table) | Yes (full flags in `docs/cli.md`) | README gives `nexokit <command>`; `docs/` carries flag detail. |
| Pre-commit hooks | Yes (one paragraph) | No | README's job. |
| Conventions overview | Yes (5-bullet pointer) | Yes (full in per-concern tutorials) | README is the index; `docs/` is the substance. |
| Project status / scope | Yes (one line: "early-stage starter") | No | README's job. |
| License / contribution / contact | Yes | No | README's job. |
| Per-concern tutorials | No | Yes (under `docs/modules/`, `docs/`) | README does not duplicate depth. |
| Historical architecture | No | Yes (under `docs/archive/`) | README does not carry history. |

The result is a README in the 100–150 line range, scannable in under a minute, that points a reader at the right `docs/` page for depth. The current 135-line README has roughly the right length but the wrong content mix: too much space is given to a 5-step "Quick Start" that the user can also follow in the Makefile, and too little to the things only the README can answer (stack, architecture one-liner, production run, folder map, links into `docs/`).

## Risks

1. **Outdated architecture docs mislead new contributors.** Three competing "architecture" files, all referencing an `internal/domain/` package that no longer exists, are an attractive nuisance. A reader landing on the repo will follow one of them and form an incorrect mental model. Mitigation: PR 2 deletes or archives them, and `docs/architecture.md` (or `docs/modules/module-structure.md` plus `docs/modules/boundaries-and-dependencies.md`) becomes the single canonical source.
2. **Dev vs prod startup accuracy.** The current README only documents `make dev`. The Makefile exposes `make build` (produces `bin/api` and `bin/nexokit`) but no README line tells the operator which binary to launch, with which env vars, and how the CLI's `serve` subcommand differs from `make dev`. There is also no prod-style entry point documented (no `Dockerfile`, no systemd unit, no env-var matrix). The change needs a `docs/deployment.md` to fill this gap; otherwise the README's "production" line would be empty.
3. **Doc sprawl / review size.** 4,770 lines across 16 files, three of them Spanish, two of them broken-link sources, one of them 2,733 lines of duplicated history. A naive "rewrite everything" change would blow the 800-line review budget. Mitigation: chained PRs (Approach B) with the index first, archival second.
4. **Dangling cross-references break reading flow.** `docs/modules.md` already links to two non-existent files. Any new index must be cross-referenced against the actual filesystem before merging, otherwise the index itself becomes a source of broken links. Mitigation: include a verification step (`grep` for relative links in the new index) before review request.
5. **Language mix.** A Spanish architecture doc in an otherwise English project is a strong signal that no one is maintaining the docs. If the change removes the Spanish doc but leaves `prompts/_context.md` and `prompts/change_XX_vertical_slice.md` in Spanish, the change is half done. Mitigation: explicitly decide the fate of `prompts/` in the proposal.
6. **Lossy archival.** Deleting the three architecture docs and the 2,733-line history doc is destructive. Even when the same content is recoverable from `openspec/changes/archive/`, a reader who only looks at `docs/` will lose context. Mitigation: keep a `docs/archive/README.md` that points at the canonical source for every archived file, and never delete `openspec/changes/archive/` content.
7. **README rewrite may overshoot the 100–150 line target.** The temptation will be to dump the full `docs/` table into the README. The skill says "lead with the answer" — the README's job is the 30-second answer, not the table of contents. Mitigation: enforce a hard 150-line cap in the proposal.
8. **`prompts/` is internal authoring material, not user docs.** Keeping it under `docs/` and indexing it from `docs/README.md` would expose SDD working prompts to new contributors. Mitigation: either move it out of `docs/` or explicitly exclude it from the public index and label it as internal in the directory listing.

## Ready for Proposal

Yes. The change should be proposed as a chained-PR change with two phases:

- **Phase 1** (`change-docs-refresh`): new `docs/README.md` index, rewritten `README.md`, fixed `docs/modules.md` and `docs/cli.md`. Review budget forecast: ~300 changed lines.
- **Phase 2** (`change-docs-archive`, follow-up change): archive the three stale architecture docs and the historical changelog, write a short `docs/architecture.md` consolidator, decide the fate of `docs/vertical-slice-modules.md` and `prompts/`. Review budget forecast: ~500 changed lines (mostly file moves, so the *additions+deletions* count is the real signal — file moves count once on each side).

The orchestrator should tell the user:

1. The 800-line review budget is the binding constraint; the work must be split across two PRs.
2. PR 1 (this change) is the safe, high-value half: index + README + small fixes.
3. PR 2 (a follow-up change) is the destructive half: archive the obsolete files and consolidate architecture.
4. The user can choose to keep `prompts/` internal-only (index it as "internal" or move out of `docs/`); both options are listed as choices in the proposal.
