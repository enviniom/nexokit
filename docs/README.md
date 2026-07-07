# NexoKit Documentation

This is the documentation index for NexoKit. Start with the root [`README.md`](../README.md) for the 30-second overview, then use the guides below for depth.

## Quick paths

1. **New to the codebase?** Read [`architecture.md`](architecture.md) first.
2. **Deploying?** Follow [`deployment.md`](deployment.md).
3. **Using as a starter template?** See [`starter-template.md`](starter-template.md).
4. **Running commands?** Use [`../Makefile`](../Makefile) for daily tasks; [`cli.md`](cli.md) is the source of truth for direct `nexokit` usage.
5. **Adding a module or slice?** Read [`modules.md`](modules.md) and [`request-flow.md`](request-flow.md).

## Documentation map

| Document | What it covers |
|----------|----------------|
| [`architecture.md`](architecture.md) | Canonical architecture: entrypoints, layers, module structure, request flow, boundaries |
| [`deployment.md`](deployment.md) | Build, environment, database, migrations, seed/root, logging, reverse proxy/TLS, ops checklist |
| [`cli.md`](cli.md) | Direct `nexokit` CLI usage: `serve`, `config`, `status`, migrations, `seed`, `create-root`, generation commands |
| [`starter-template.md`](starter-template.md) | How to clone/fork, rename, configure, and adopt NexoKit as a project starter |
| [`request-flow.md`](request-flow.md) | End-to-end HTTP path through auth, tenant, permissions, and modules |
| [`modules.md`](modules.md) | Module layout, vertical-slice rules, and links to per-concern tutorials |
| [`api-conventions.md`](api-conventions.md) | DTO naming, list queries, soft delete, tenant scope, response envelopes |
| [`error-handling.md`](error-handling.md) | Error classification, `apperror`, `response.HandleError`, debug exposure |
| [`module-error-conventions.md`](module-error-conventions.md) | Per-module error vocabulary and conventions |
| [`multitenancy.md`](multitenancy.md) | Tenant scope rules for repositories and handlers |
| [`testing.md`](testing.md) | Test commands, integration setup, CI reproduction |

## Archived documentation and internal material

Old architecture versions and the historical SDD changelog live in [`archive/README.md`](archive/README.md). Root-level `prompts/` contains internal SDD authoring material and is excluded from this public index.
