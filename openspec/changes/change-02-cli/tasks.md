# Tasks: Internal CLI and Developer Experience

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 800–1100 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (Foundation+Commands) → PR 2 (Generators) → PR 3 (Root+Tests+Docs) |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes  
Chained PRs recommended: Yes  
Chain strategy: stacked-to-main  
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Makefile + CLI dispatcher + serve/config/status/migrate | PR 1 | Base branch main; includes tests |
| 2 | Module/migration/seed generators + templates | PR 2 | Depends on PR 1 CLI contract |
| 3 | create-root + seed command + golden tests + docs | PR 3 | Depends on PR 1; root storage boundary ready for auth change |

## Phase 1: Foundation

- [x] 1.1 Add Makefile targets: `dev`, `build`, `test`, `migrate-up`, `migrate-down`, `migrate-create`, `seed`, `create-root`, `lint`, `fmt`.
- [x] 1.2 Create `cmd/nexokit/main.go` with exit-code handling and `cli.Execute` call.
- [x] 1.3 Create `internal/cli/cli.go` with `Command` interface, registry, dispatcher, and usage text.
- [x] 1.4 Create `internal/cli/stdio.go` with `Stdio` struct and helpers.

## Phase 2: Core Commands

- [x] 2.1 Implement `serve` command: load config, bootstrap app, start HTTP server.
- [x] 2.2 Implement `config` command: print resolved config, mask secrets.
- [x] 2.3 Implement `status` command: print app version, DB status, migration count.
- [x] 2.4 Implement `migrate up/down/status/reset` using Goose helpers in `internal/infra/db`.
- [x] 2.5 Implement `migrate create <name>` with snake_case validation and timestamped SQL in `migrations/`.

## Phase 3: Generators

- [x] 3.1 Create `internal/cli/generator/module.go` with name validation, directory creation, and idempotency check.
- [x] 3.2 Create `internal/cli/templates/module/*.tmpl` for handler, service, repository, dto, model, routes, validation.
- [x] 3.3 Implement `make module <name>` with flags `--crud`, `--migration`, `--tenant`.
- [x] 3.4 Implement `make migration <name>` command.
- [x] 3.5 Implement `make seed <name>` command.

## Phase 4: Root & Seed

- [ ] 4.1 Create `internal/cli/root/` package with safe root creation boundary and idempotency check.
- [ ] 4.2 Implement `create-root` command with email/password flags or prompt, validation, and transaction wrapper.
- [ ] 4.3 Implement `seed` command dispatcher for running seed files.

## Phase 5: Testing & Docs

- [ ] 5.1 Write unit tests for dispatcher, command validation, and module name normalization.
- [ ] 5.2 Write golden file tests for `make module` with `--crud`, `--migration`, `--tenant` in `tests/fixtures/cli/golden/`.
- [ ] 5.3 Write integration tests for migrate commands and root safety (skip if no DB env).
- [ ] 5.4 Create `docs/cli.md` documenting commands, flags, and out-of-scope items.
