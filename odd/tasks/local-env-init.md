# Local-development environment initialization

## Objective and authorization
User confirmed extending interactive init/make init to copy adopter .env.example into new .env, ask essential PostgreSQL/app/CORS settings, generate random development PASETO key, redact summary, confirm writes, preserve unknown example options and never overwrite existing .env. Development-only; no database provisioning/migrations/remote GitHub/licensing/production promise. Existing explicit starter work-unit commit authorization retained; no push/PR.

## Exploration and design
- Existing interactive name/GitHub flow builds module and safe identity plan; preserve scripted flag contract/no stdin reads.
- Existing Plan snapshots .env.example and transactional identity updates but does not support new .env creation. Extend narrow optional creation, not general file updater.
- Current Makefile pending parent-owned make init alias skips .env for sole init; other goals use GNU make include/exports and migration shell guards. GNU make is not a dotenv parser: quoted/special password values and exports conflict with godotenv.Load, so remove raw dotenv interpretation and let Go commands load configuration. Preserve runtime/default target/alias, migrate command forwarding and external environment overrides. Add regression evidence.
- Existing godotenv v1.5.1; no x/term declared. Use existing parser for round-trip tests, preserve comments/unknown options. Hidden real-TTY password entry requires narrowly adding compatible golang.org/x/term (no raw platform termios tricks). Non-TTY synthetic/piped input may be consumed without echo; never silently visible terminal fallback. Single session reader must not lose buffered lines.
- Token layer hashes secret to 32 bytes; generate 32 crypto/rand bytes encoded hex (64 chars), not template/default/shared key. No need production generation support.
- DATABASE_URL overrides DB_*: generated .env explicitly sets DATABASE_URL empty so questionnaire DB fields apply.
- pgx modes: disable/allow/prefer/require/verify-ca/verify-full. Default disable explicitly local-development only.

## Product behavior
- Bare init asks name/GitHub then config only if .env absent. Existing regular .env is preserved without reading and config questions skipped; symlink/directory destination rejected. Already adopted module remains existing initializer error; no update mode.
- Ask PostgreSQL host default localhost, port 5432, database default normalized app slug, user default postgres, password hidden/no echoed default, SSL mode default disable.
- APP_ENV development; APP_NAME from identity. Ask APP_PORT default 8080, APP_URL default http://localhost:<chosen-port>, CORS default http://localhost:3000. Validate ports/URLs/origin list, no credentials/newline/control injection. Other example options kept byte-preserved where unchanged.
- Generate key internally, never display/log/dump password/key/template contents; preview path/field labels and nonsecret settings only. One combined default-No confirmation before ANY identity/env writes.
- .env mode 0600, creation exclusive/no symlink/race overwrite, temporary secret files mode restrictive and cleaned; failed create or later apply rolls back coherent unit, no partial identity/.env left or secret in error.
- Explicit flag form unchanged: no new config prompts or silent credential defaults; document wizard limited to interactive path.

## Tasks
- [ ] E1 — Make CLI wrappers dotenv-safe, preserving make init/dev/migration behavior with deterministic regressions and narrow docs. **In progress: writer and independent verifier PASS; prepare scoped tested-worktree commit, refreshing stale Makefile staging.**
- [ ] E2 — Extend initializer plan with exclusive transactional .env creation from synthetic template and preservation/escaping/rollback tests. **Pending.**
- [ ] E3 — Wire local configuration/password/key prompts and redacted confirmation; docs and real copied-source wizard/compatibility verification. **Pending.**

## Privacy and verification boundaries
NEVER read/copy/edit actual original .env OR .env.example. Prior tool denial unresolved; supplied example anchor only APP_NAME=nexokit. Runtime adopter file handling is requested product behavior; development verification uses invented synthetic example/options/secrets exclusively. Original-root make/init forbidden; any config-loading test command runs in narrow disposable source tree excluding hidden/env/Git/odd/log/tmp/vendor/generated artifacts/symlinks. Include source internal/infra/cache and internal/config/env.go, not classify by names alone. Keep parent existing Makefile alias change intact.
Test-first meaningful RED/GREEN per behavior unit; parent ASSESS after writer and exact risk-gated independent verification. Tests/fixtures remain with behavior/docs commit. Hosted CI/live PostgreSQL/production deployment excluded; exact dependency/service skips reported. Password cancellation at blocking stdin remains limited to before/after reads, no goroutine leaks.
Review burden: split into three coherent units instead of one broad transaction/prompt/Makefile change; report authored sizes, never minify/omit safety tests to meet advisory budget.

## Evidence
- E1 writer `mv6741ir-i-rwc5`: meaningful RED all seven synthetic wrapper scenarios failed on unexpected pre-load DB_PASSWORD exports; GREEN after removing raw include/exports/migration guards. Quotes/spaces/#/$ and literal command-like values preserved as data; unset/empty/external overrides, argv/stdin, init/dev/defaultdev/four migration forwards covered. New 114-line test.
- E1 independent verifier `mv678y2q-j-jfth`: PASS no blockers in narrow source-identical secret-free copy. Fresh seven-scenario test, tests/cli + commands, actual interactive source-copy smoke, build/vet, global format, guard and whitespace passed. Default/dev/init dry-run forwards correct. Three real-DB command tests skipped no DB configuration; inner external smoke recursion skips expected. Signal handling not separately exercised and full aggregate suite not repeated (focused scope + inner copied-source full smoke); no services/network/secret reads.
- ASSESS unassessable because untracked files => high-risk separate verifier path; RDD off, no native review started. Index contains stale intermediate Makefile include/exports/guards; refresh only scoped Makefile staging from tested worktree. No unrelated staged paths; hook absent. E1 commit pending.
- E2 RED/GREEN, transaction/serializer verification and commit: pending.
- E3 RED/GREEN, hidden input/redaction/wizard/copy verification and commit: pending.
- Legal terms untouched, no source updater or dependency upgrades beyond necessary compatible x/term addition.

## Next step
E1 verified; commit scoped tested Makefile/tests/docs/tracking and record hash, then E2 transactional .env foundation and E3 wizard single-threaded. Preserve unrelated staging/work. User explicitly requested computer shutdown AFTER all three units finish and checks/evidence are saved; do not shut down while pending/blocked, use normal authorized OS mechanism without privilege bypass only after completion.
