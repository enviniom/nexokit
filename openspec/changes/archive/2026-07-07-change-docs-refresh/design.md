# Design: Documentation Refresh

## Technical Approach

This is a documentation-only consolidation. The implementation will rewrite navigation and operations docs from verified sources: `README.md`, `Makefile`, `.env.example`, `docker-compose.yml`, `docs/cli.md`, `docs/modules.md`, existing module guides, CLI command code, and OpenSpec config/spec context. No Go behavior, API contracts, migrations, or runtime configuration change.

Documentation flow:

```text
README.md -> docs/README.md -> topic docs
                         |-> docs/architecture.md
                         |-> docs/deployment.md
                         |-> docs/cli.md
                         |-> docs/starter-template.md
prompts/ stays outside public docs navigation
```

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Documentation information architecture | README is the 30-second entry point; `docs/README.md` is the docs index; topic docs own depth. | Put all detail in README; keep current scattered docs. | Reduces cognitive load and prevents README from duplicating long-form docs. |
| Canonical architecture | Create `docs/architecture.md` and make it the only architecture page linked from docs. | Keep three existing architecture versions; edit the least stale file. | Existing architecture docs reference non-existent `internal/domain/`; one canonical page prevents wrong mental models. |
| Starter positioning | Frame NexoKit as a starter framework/template, not a managed hosting platform. | Present it as deployable platform documentation. | Matches project intent and avoids overpromising production operations. |
| CLI source of truth | README uses Makefile-first quick commands; `docs/cli.md` documents direct `nexokit` commands. | Duplicate all CLI details in README. | Make targets are the happy path; direct CLI behavior belongs in one focused reference. |
| Historical material | Move obsolete architecture docs and SDD changelog to `docs/archive/` with `docs/archive/README.md` pointers. | Delete old docs outright. | Preserves context while removing stale docs from the active reading path. |

## File Changes

| File | Action | Description |
|---|---|---|
| `README.md` | Modify | Add stack table, docs index link, concise project map, starter-template adoption section, Makefile-first dev/test/build path, minimal production command path linking to deployment. |
| `docs/README.md` | Create | Top-level docs index: quick paths, documentation map, and canonical links only. |
| `docs/architecture.md` | Create | Consolidated architecture overview: current entrypoints, `internal/app`, `internal/modules/{auth,companies,iam,onboarding}`, platform/infra, request flow, module boundaries, and links to deep dives. |
| `docs/cli.md` | Modify | Make CLI source of truth for `serve`, `migrate up/down/status/reset/create`, `seed`, `create-root`, and generation commands; remove stale persistence-blocked root note. |
| `docs/deployment.md` | Create | Production starter guide: build (`make build`), secure env, DB/SSL, migrations, seed/root, logging files, reverse proxy/TLS handoff, process supervision, and operational checklist. |
| `docs/starter-template.md` | Create | How to adopt NexoKit as a starter: clone/fork, rename module/app, configure env, choose modules, run migrations/seeds/root, and replace defaults. |
| `docs/modules.md` | Modify | Fix broken OpenSpec links and point readers to current canonical architecture/spec paths. |
| `docs/prompts/*` -> `prompts/*` | Move | Move internal prompt template/context out of project docs; do not link from public docs index. |
| `docs/nexokit-architecture*.md` | Move | Archive under `docs/archive/` with pointers to `docs/architecture.md`. |
| `docs/changes_sdd_nexokit_go_completo.md` | Move | Archive as historical duplicate; point to `openspec/changes/archive/` as canonical history. |
| `docs/archive/README.md` | Create | Explain archived files and canonical replacements. |

## Interfaces / Contracts

No runtime interfaces change. Documentation contracts:
- All active project docs are in English.
- README does not duplicate topic depth; it links to `docs/README.md`, `docs/deployment.md`, `docs/cli.md`, and `docs/starter-template.md`.
- `docs/architecture.md` is the only active architecture document.
- `prompts/` is internal authoring material, outside `docs/` navigation.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Link validation | Relative Markdown links and moved paths | Extract `](...)` links from changed `.md` files; validate local targets exist; grep for old `docs/prompts/` and archived architecture links. |
| Command accuracy | README/CLI/deployment commands | Cross-check against `Makefile` targets and CLI code: `serve`, `migrate`, `seed`, `create-root`, `make module/migration/seed`. |
| English consistency | Active docs only | Grep active `README.md`/`docs/*.md` for obvious Spanish headings/phrases; archived docs and root `prompts/` are exempt. |
| Review size | 800-line guideline | Check `git diff --stat` and additions+deletions before review. The 800-line figure is a guideline, not a hard gate; the user-approved `size:exception` allows a single PR, and reviewers should use the documented work-unit sections to focus review. |

## Migration / Rollout

No data migration required. Rollout is a single documentation PR. The 800-line figure is a guideline. Because the workload decision is `size:exception`, a single PR is acceptable; reviewers should use the documented work-unit sections rather than the raw line count as a hard pass. If future changes exceed the guideline without an exception, split archive moves into a follow-up before apply/review.

## Open Questions

None.
