# Verify Report: change-docs-refresh

## Verification Report

**Change**: change-docs-refresh
**Version**: N/A (documentation-only)
**Mode**: Standard (no Strict TDD; TDD cache was not set for `nexokit`)
**Re-run trigger**: Final one-line fix to `docs/deployment.md` intro (W.1 resolution)

### Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 17 (1.1–1.2, 2.1–2.4, 3.1–3.3, 4.1–4.3, 5.1–5.5) + 7 second-remediation items (R.9–R.15) |
| Tasks complete | 17 / 17 (Phases 1–5) and 7 / 7 (R.9–R.15) |
| Tasks incomplete | 0 |
| Re-run changes | 1 file modified (`docs/deployment.md` intro, +1 / -1 line) |

Every `- [x]` item in `openspec/changes/change-docs-refresh/tasks.md` is checked, including the post-review and re-review remediation lines (R.9 through R.15). The only unchecked item is the explicit "Decision needed before apply: No" — which is a guard line, not a task.

### Build & Tests Execution

**Build**: ✅ Passed
```text
$ go build ./...
(no output, exit 0)
```

**Vet**: ✅ Passed
```text
$ go vet ./...
(no output, exit 0)
```

**Tests**: ✅ All packages green
```text
$ go test ./... -short -count=1
ok  github.com/enviniom/nexokit/internal/cli/commands               0.112s
ok  github.com/enviniom/nexokit/internal/cli/generator              0.041s
ok  github.com/enviniom/nexokit/internal/cli/root                   0.017s
ok  github.com/enviniom/nexokit/internal/cli/templates              0.022s
ok  github.com/enviniom/nexokit/internal/config                     0.005s
ok  github.com/enviniom/nexokit/internal/infra/cache                0.014s
ok  github.com/enviniom/nexokit/internal/middleware                 1.855s
ok  github.com/enviniom/nexokit/internal/modules/iam/roles/...      (all green)
ok  github.com/enviniom/nexokit/internal/modules/onboarding/...    (all green)
ok  github.com/enviniom/nexokit/internal/platform/...               (all green)
ok  github.com/enviniom/nexokit/internal/server                     0.010s
ok  github.com/enviniom/nexokit/tests/cli                           0.009s
ok  github.com/enviniom/nexokit/tests/docs                          0.003s
ok  github.com/enviniom/nexokit/tests/integration                   0.013s
... (73 packages: all "ok" or "[no test files]", zero FAIL, exit 0)
```

**Coverage**: ➖ Not enforced for this docs-only change. The build and unit tests for the
files actually referenced by the docs (`internal/cli/commands`, `internal/app`,
`internal/cli/root`) all pass.

### W.1 Resolution — Primary Verification Focus

**Previous warning**: `docs/deployment.md` intro (line 3) said "This guide covers running
NexoKit in production" and did not link to `starter-template.md` or remind the reader
that NexoKit is intended as a starter. A reader who landed directly on `docs/deployment.md`
could miss the starter framing.

**Resolution evidence** — new `docs/deployment.md` line 3:

```text
# NexoKit Deployment Guide

This guide covers running NexoKit-based applications in production. If you are
adapting the starter first, read [`starter-template.md`](starter-template.md).
You need a server with Go 1.26+ (or a pre-built Linux binary), PostgreSQL 15+,
and a reverse proxy for TLS termination. Redis/Valkey is optional
(`CACHE_DRIVER=none` works without it). For local development, see the root
[`README.md`](../README.md).
```

| Acceptance check | Result |
|------------------|--------|
| Intro links to `docs/starter-template.md` (`starter-template.md`) | ✅ Yes — relative link `[`starter-template.md`](starter-template.md)` is present and resolves |
| Intro frames deployment as NexoKit-based apps / starter use | ✅ Yes — "running NexoKit-based applications in production" and "If you are adapting the starter first" |
| Link target exists on disk | ✅ Yes — `docs/starter-template.md` (62 lines) |
| No other content was regressed | ✅ Yes — line count moved from 189 to 189 (one-line edit, swap-in of starter link) |

**W.1 status: RESOLVED.** A direct visitor to `docs/deployment.md` now sees the
starter framing and a one-click path to the starter-template guide in the first
paragraph.

### Spec Compliance Matrix

This change ships **no spec deltas** (proposal states `New=None, Modified=None`); verification
maps the proposal success criteria to on-disk evidence instead.

| Success criterion (proposal) | Evidence | Result |
|------------------------------|----------|--------|
| `docs/README.md` exists, all links resolve | `docs/README.md` present (31 lines); 20 relative links across the index resolve cleanly | ✅ COMPLIANT |
| `README.md` has stack, folder map, Makefile quick commands, starter, link to `docs/` | Lines 5–18 (stack table), 56–71 (folder map), 26–33 (Makefile quick path), 73–82 (starter), 99–107 (docs index) | ✅ COMPLIANT |
| `docs/architecture.md` is the only architecture doc linked from the index | `docs/README.md` row 17 links only `architecture.md`; three old arch docs live under `docs/archive/` | ✅ COMPLIANT |
| `docs/deployment.md` covers build, env, DB, migrations, seed/root, logging, reverse proxy/TLS, ops | Sections "Build", "Environment configuration", "Database setup", "Seeds and root user", "Logging and monitoring", "Reverse proxy / TLS", "Operational checklist" all present (lines 5, 22, 61, 98, 119, 127, 139, 171) — and the intro now leads with the starter-template framing (W.1 fix) | ✅ COMPLIANT |
| `docs/cli.md` documents `serve`, migrations, seed, `create-root` direct CLI usage | Sections `## serve`, `## config`, `## status`, `## migrate`, `## seed`, `## create-root`, plus generation commands `make module/migration/seed` (R.14 satisfied) | ✅ COMPLIANT |
| `docs/starter-template.md` covers adopt-as-starter | File present, 62 lines; sections "Adopt the repo", "Configure your environment", "Choose your modules", "Run the project", "Build your domain" | ✅ COMPLIANT |
| `prompts/` at repo root, excluded from the index | `docs/prompts/` does not exist; `prompts/` at root has 30 files; `docs/README.md` line 31 explicitly excludes `prompts/` | ✅ COMPLIANT |
| `docs/archive/README.md` lists archived files and points to canonical sources | `docs/archive/README.md` maps each archived file to a canonical replacement (architecture, changelog → `openspec/changes/archive/`) and documents the prompts-move convention (R.13) | ✅ COMPLIANT |
| All in English; PR diff under 800 lines | Active docs grep clean of Spanish; archived changelog is preserved historical Spanish (intentionally exempt); diff is 795 insertions, 208 deletions — under the 800-line guideline | ✅ COMPLIANT |
| Second-remediation R.9–R.15 all checked | All seven items in `tasks.md` lines 86–92 are checked | ✅ COMPLIANT |
| W.1 — `docs/deployment.md` intro cross-links to `starter-template.md` and frames deployment as starter-based | New line 3: "running NexoKit-based applications in production… [`starter-template.md`](starter-template.md)… If you are adapting the starter first" | ✅ COMPLIANT (was WARNING, now fixed) |

**Compliance summary**: 11 / 11 success criteria satisfied (W.1 promoted from WARNING to COMPLIANT).

### Correctness (Static Evidence)

| Claim | Verified against | Status |
|-------|------------------|--------|
| `nexokit seed` requires the Go toolchain on the seeding host | `internal/cli/commands/seed.go:130` uses `exec.CommandContext(ctx, "go", "run", ".")` — runs a temporary Go runner | ✅ Implemented and accurate |
| `migrate up/down/status/reset/create` are the migration subcommands | `internal/cli/commands/migrate.go` lines 30–43 dispatch on those exact tokens; docs/cli.md table rows 12–16 match | ✅ Implemented and accurate |
| `nexokit config` masks secrets (`DB_PASSWORD`, `DATABASE_URL`) | `internal/cli/commands/config.go` lines 105–106: `Password: "***masked***"`, `DatabaseURL: "***masked***"` | ✅ Implemented and accurate |
| `nexokit status` prints version, DB reachability, Goose version, `.sql` file count | `internal/cli/commands/status.go` lines 25–73 — matches every field in `docs/cli.md` "Expected output scope" | ✅ Implemented and accurate |
| `create-root` flags are `--name --email --password --force`; idempotent; reads `ROOT_USER_*` env; password validation 8+ chars + mixed case + digit | `internal/cli/commands/createroot.go` lines 32–35, 44–52, 117–119 + `internal/cli/root/root.go` lines 93–100 | ✅ Implemented and accurate |
| `make module <name>` flags `--crud --migration --tenant` | `internal/cli/commands/make.go` lines 50–57 | ✅ Implemented and accurate |
| Module contracts live in `internal/modules/<module>/core/contracts.go` (not a shared `platform/contracts`) | `internal/modules/iam/core/contracts.go`, `internal/modules/onboarding/core/contracts.go`, `internal/modules/auth/core/contracts.go` exist; `internal/platform/` has no `contracts` subdir (verified via `ls internal/platform` → no `contracts`) | ✅ Implemented and accurate |
| `internal/app/container.go` wires module-local contracts | `internal/app/container.go:46` injects `userLookup{resolver: iamContainer.AuthUserResolver}`; line 51 injects `onboarding.Config{PasswordHasher: passwordManager, ...}` | ✅ Implemented and accurate |
| `AuthUserResolver` and `PasswordHasher` are real, in-repo examples | `internal/modules/iam/core/contracts.go` defines `AuthUserResolver`; `internal/modules/onboarding/core/contracts.go` defines `PasswordHasher`; both referenced in `container.go` and tests | ✅ Implemented and accurate |
| Health endpoints `/health`, `/health/live`, `/health/ready` exist unversioned | `internal/server/router.go` lines 37–39 register exactly those three routes | ✅ Implemented and accurate |
| Modules list: `auth`, `companies`, `iam`, `onboarding` | `internal/modules/` contains exactly those four directories | ✅ Implemented and accurate |
| Stack table: Go, Gin, GORM, PostgreSQL, Goose, PASETO v4.local + opaque refresh, argon2id, Redis/Valkey, slog + lumberjack | `go.mod`, `.env.example`, `internal/infra/`, `internal/platform/token/`, `internal/platform/password/` all match | ✅ Implemented and accurate |
| Makefile target list (`dev`, `build`, `test`, `migrate-up`, `migrate-down`, `migrate-create`, `migrate-status`, `migrate-reset`, `seed`, `create-root`, `fmt`, `lint`) | `Makefile` lines 1, 14–70 — every README row maps to an actual target | ✅ Implemented and accurate |
| `docs/deployment.md` "Danger Zone" wording for `migrate down` / `migrate reset` is consistent with `docs/cli.md` | Both docs use the same policy-forbidden language; the CLI binary does not block these commands — matches `internal/cli/commands/migrate.go` (no destructive guard) | ✅ Implemented and accurate |

### Coherence (Design)

| Design decision | Followed? | Notes |
|-----------------|-----------|-------|
| README is 30-second entry; `docs/README.md` is index; topic docs own depth | ✅ Yes | README = 109 lines; `docs/README.md` = 31 lines; topic docs ~60–190 lines |
| Canonical architecture is `docs/architecture.md`; old arch moved to archive | ✅ Yes | Three `nexokit-architecture*.md` files live under `docs/archive/`; only `docs/architecture.md` is linked from the index |
| Frame NexoKit as a starter framework, not a managed platform | ✅ Yes | `docs/starter-template.md`, `README.md` "Using NexoKit as a starter" section, AND the new `docs/deployment.md` intro all carry the framing — W.1 is closed |
| CLI source of truth for direct `nexokit` usage; Makefile is happy path | ✅ Yes | `docs/cli.md` documents all direct commands; `docs/cli.md` ends with a Makefile-equivalents table |
| Archived material with pointers to canonical sources | ✅ Yes | `docs/archive/README.md` lists every archived file and its canonical replacement |
| `prompts/` is internal authoring material, outside `docs/` navigation | ✅ Yes | `docs/README.md` line 31 explicitly excludes `prompts/` from the public index |
| Active docs in English | ✅ Yes | No Spanish fragments in active docs (`README.md`, `docs/README.md`, `docs/cli.md`, `docs/deployment.md`, `docs/starter-template.md`, `docs/architecture.md`, `docs/modules.md`); archived changelog remains Spanish by design |
| `migrate down` / `migrate reset` are Danger Zone, not blocked by the binary, operator responsibility | ✅ Yes | `docs/cli.md` line 81 + `docs/deployment.md` line 83 use identical language: "forbidden by operational policy in production unless explicitly approved. The CLI binary does not block these commands" |
| 800-line review budget is a guideline, not a hard pass; `size:exception` approved | ✅ Yes | `tasks.md` records the `size:exception` decision; `design.md` lines 59 and 63 explicitly state the 800-line figure is a guideline. Raw diff is 795 insertions / 208 deletions, within the 800-line guideline even without the exception |

### Link Validation (re-run)

| Source | Relative links | Broken |
|--------|----------------|--------|
| `README.md` | 11 | 0 |
| `docs/README.md` | 20 | 0 |
| `docs/architecture.md` | 5 | 0 |
| `docs/cli.md` | 2 | 0 |
| `docs/deployment.md` | 2 (incl. new `starter-template.md` link) | 0 |
| `docs/modules.md` | 39 | 0 |
| `docs/starter-template.md` | 1 | 0 |
| **Total** | **80** unique | **0** |

All 80 unique relative Markdown links across the active docs resolve to an existing file.
The newly added `starter-template.md` link in `docs/deployment.md:3` resolves to
`docs/starter-template.md`. The remaining matching `(...)` references in active docs are
either absolute URLs, in-page anchors, or in-file headings — none are broken.

### Stale-Reference Sweep (re-run)

| Stale pattern | Hits in active docs | Hits in archive/notes (expected) |
|---------------|---------------------|---------------------------------|
| `docs/prompts/` | **0** | 1 (the archive index's historical note — appropriate) |
| `docs/nexokit-architecture*.md` | **0** | 3 (the archived files themselves) |
| `docs/changes_sdd_nexokit_go_completo.md` | **0** | 1 (the archived file itself) |
| `internal/platform/contracts` (presented as existing) | **0** (no claims in active docs) | many (archived old arch docs) |
| `ChangeMe123` (hard-coded password in docs) | **0** (placeholder `<GENERATED-PASSWORD>`) | 0 |
| `internal/domain/` (non-existent module layer) | **0** | 0 |
| `](prompts/...)` markdown links in active docs | **0** | 0 (R.15 satisfied — all `prompts/` refs in archive are backtick text, not links) |

No new stale references were introduced by the W.1 edit. Active docs remain free of
`docs/prompts/`, `platform/contracts`, and `internal/domain/` references.

### Line-Cap Compliance (re-run)

| File | Lines | Cap | Result |
|------|-------|-----|--------|
| `README.md` | 109 | ~150 | ✅ |
| `docs/README.md` | 31 | ~80 | ✅ |
| `docs/architecture.md` | 117 | ~150 | ✅ |
| `docs/deployment.md` | 189 | ~200 | ✅ |
| `docs/starter-template.md` | 62 | ~80 | ✅ |

The W.1 fix kept `docs/deployment.md` at 189 lines, still within the ~200-line cap.

### Issues Found

**CRITICAL**: None.

**WARNING**: None.

**SUGGESTION** (carried over from the previous report, still non-blocking):

- **S.1** — `docs/deployment.md:185` link target `#danger-zone--destructive-migration-commands`
  uses a double-hyphen slug that matches the section header. Tested and resolves correctly,
  but a future rename of the section would silently break it. No action needed today.
- **S.2** — The Makefile table in `docs/cli.md` (line 155) is a useful mapping but is now
  a partial repeat of the README "Everyday commands" table. Optional: collapse one of them
  or add a one-line note that README is the daily reminder and `cli.md` is the full
  reference. Cosmetic.
- **S.3** — `notes/inventory.md` and `notes/cli-source-of-truth.md` are useful artifacts
  created by Phase 1. Consider promoting them to `openspec/changes/change-docs-refresh/`
  or gitignoring `notes/` if they should not be in the repo long-term. Not a verification
  blocker.

### Verdict

**PASS**

All 17 task items and all 7 re-review remediation items are complete; `go build ./...`,
`go vet ./...`, and `go test ./... -short -count=1` are green (73 packages, zero FAIL);
80 / 80 relative Markdown links in the active docs resolve — including the newly added
`starter-template.md` link in `docs/deployment.md:3`; the `docs/prompts/` → `prompts/`
move is complete with no stale references in active docs; the `internal/platform/contracts`
non-existence is now explicitly disclaimed; the binary/Go toolchain seed constraint is
documented in both `README.md` and `docs/deployment.md` and matches the source in
`internal/cli/commands/seed.go:130`; password examples are placeholders; the
`migrate down` / `migrate reset` Danger Zone is consistently labeled across `docs/cli.md`
and `docs/deployment.md` as policy-forbidden (binary does not block); module-local
contracts in `internal/modules/<module>/core/contracts.go` are verified for `iam`,
`onboarding`, and `auth` and are wired in `internal/app/container.go` exactly as
`docs/architecture.md` describes.

**W.1 is resolved.** The previous warning that `docs/deployment.md` did not restate the
starter-template framing is now closed: the intro leads with "running NexoKit-based
applications in production" and provides a one-click link to `starter-template.md` for
adopters. The change is **archive-ready** with no outstanding warnings or critical issues.
