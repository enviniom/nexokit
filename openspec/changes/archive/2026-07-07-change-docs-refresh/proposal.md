# Proposal: Documentation Refresh

## Intent

Docs are undiscoverable and stale: README omits stack/modules/prod path; 3 arch docs reference a non-existent `internal/domain/`; a 2,733-line change-log duplicates `openspec/changes/archive/`; `docs/prompts/` mixes internal SDD with user docs. Consolidate into one English corpus and frame NexoKit as a starter framework.

## Scope

### In Scope
- New `docs/README.md` index.
- Rewrite `README.md` with Makefile-first quick commands; full reference in `docs/cli.md` and `docs/deployment.md`.
- New: `docs/architecture.md` (consolidates arch), `docs/deployment.md` (build/env/DB/migrations/seed/ops), `docs/starter-template.md`.
- Expand `docs/cli.md` for direct `nexokit` `serve`, migrations, seed, `create-root` (Makefile = happy path).
- Move `docs/prompts/` to repo root; fix `docs/modules.md` link.
- Archive stale arch docs + change-log in `docs/archive/`.

### Out of Scope
Code, API, test, or runtime changes; no spec deltas; no re-translation of `prompts/` Spanish.

## Capabilities

### New Capabilities
None — documentation-only.

### Modified Capabilities
None — no spec-level behavior changes.

## Approach

Single PR, tight tasks (single-pr-default) within the 800-line review budget. Source of truth: OpenSpec specs (read-only). All docs in English per cognitive-doc-design. Line caps: README ~150, arch ~150, deployment ~200, starter ~80.

## Affected Areas

- `README.md` — modified (rewritten; Makefile-first quick commands).
- New docs: `docs/README.md` (index), `docs/architecture.md`, `docs/deployment.md` (production guide), `docs/starter-template.md` (starter), `docs/archive/` (stale arch + change-log).
- `docs/cli.md` — modified (expanded: `serve`, migrations, seed, `create-root` direct CLI usage).
- `docs/modules.md` — modified (fix link).
- `prompts/` — new at root; 4 stale docs moved to `docs/archive/`.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Diff exceeds 800-line budget | Med | Slice tasks; chained commits. |
| Archive deletes lose context | Med | Archive README points at source. |
| New index has broken links | Med | Verify links before review. |
| README overshoots 150 lines | Low | Hard cap; depth in `docs/`. |
| `prompts/` move breaks references | Low | Grep first; update call sites. |

## Rollback Plan

Revert the PR (docs + folder moves only); `git revert` restores everything with no runtime side effects.

## Dependencies

OpenSpec specs are source of truth (read-only); no runtime deps.

## Success Criteria

- [ ] `docs/README.md` exists, all links resolve.
- [ ] `README.md` has stack, folder map, Makefile quick commands, starter, link to `docs/`.
- [ ] `docs/architecture.md` is the only architecture doc linked from the index.
- [ ] `docs/deployment.md` covers build, env, DB, migrations, seed/root, logging, reverse proxy/TLS, ops.
- [ ] `docs/cli.md` documents `serve`, migrations, seed, `create-root` direct CLI usage.
- [ ] `docs/starter-template.md` covers adopt-as-starter.
- [ ] `prompts/` at repo root, excluded from the index.
- [ ] `docs/archive/README.md` lists archived files and points to canonical sources.
- [ ] All in English; PR diff under 800 lines.
