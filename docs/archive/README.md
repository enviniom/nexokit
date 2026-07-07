# Archived Documentation

This directory contains older documentation that has been removed from the active reading path. The files are preserved for historical context. For current guidance, use the canonical sources listed below.

## Archived files

| File | Reason archived | Canonical replacement |
|------|-----------------|-----------------------|
| `nexokit-architecture.md` | Stale; references a non-existent `internal/domain/` layout | [`docs/architecture.md`](../architecture.md) |
| `nexokit-architecture-v2.md` | Superseded by the consolidated architecture guide | [`docs/architecture.md`](../architecture.md) |
| `nexokit-architecture-revised.md` | Stale Spanish-language architecture doc | [`docs/architecture.md`](../architecture.md) |
| `changes_sdd_nexokit_go_completo.md` | Historical SDD changelog; duplicated by `openspec/changes/archive/` | [`openspec/changes/archive/`](../../openspec/changes/archive/) |

## Prompts move

Internal SDD prompt templates previously lived in `docs/prompts/` and have moved to the repository root `prompts/`. They are intentionally excluded from the public documentation navigation.

The archived changelog previously contained markdown links to `prompts/`. Those direct links have been removed and replaced with plain path references so public docs navigation does not expose prompts. The archived historical content may still contain preserved examples (e.g., placeholder tokens, sample JSON) that were part of the original planning material.

## Current documentation entry points

- [`docs/README.md`](../README.md) — docs index
- [`docs/architecture.md`](../architecture.md) — canonical architecture
- [`docs/deployment.md`](../deployment.md) — production guide
- [`docs/cli.md`](../cli.md) — CLI reference
- [`docs/starter-template.md`](../starter-template.md) — starter template adoption
