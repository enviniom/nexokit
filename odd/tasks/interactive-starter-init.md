# Interactive starter initialization

## Objective and authorization
User requested interactive init, asking for GitHub username and automatically deriving the module. Focused product decision resolved: ask application display name AND GitHub username, not infer application identity from the folder. Existing explicit starter work-unit commit authorization applies to this bounded starter extension; no push/PR/merge.

Baseline: clean `feat/tenant-admin-access` at `c06e9cf`; previous initializer `e0c834f`, CI `1ab41b3`.

## Scope and design
- Bare `go run ./cmd/nexokit init` enters interactive mode.
- Ask display name and GitHub username; derive repository slug by lowercasing ASCII name and replacing runs of spaces/dots/underscores/hyphens with one hyphen, trim separators, reject empty/unsafe result. Keep original trimmed display name for application identity.
- Validate GitHub username locally (1–39 ASCII alphanumeric/hyphen, no leading/trailing or consecutive hyphens); no credential prompts, network, existence lookup or remote repository creation.
- Construct `github.com/<username>/<slug>` automatically; show derived identity and existing plan before asking apply confirmation, default No.
- Explicit flag form keeps current behavior: --name/--module required, preview default, --apply explicit; missing/invalid/partial flags stay errors and never unexpectedly prompt.
- Reuse existing planner/preflight/symlink/rollback protections rather than duplicate them or broaden rewrite surfaces.
- No new dependencies, Git automation, DB provisioning, production-key generation, licensing or updater.
- Technical prompts/docs/tests default English like current project.

## Tasks
- [x] I1 — Implement interactive init with automated identity derivation, safety tests and adoption docs; independently verify and close as one work-unit commit. **Verified and committed in `1769aa8` (`feat(cli): derive starter identity through interactive init`).**

## Acceptance
- Bare init prompts for name and GitHub username only for identity, derives expected module, then previews and asks explicit confirmation.
- Yes applies exactly the verified plan; No/blank/EOF at confirmation leaves all files unchanged.
- Missing input, invalid names/usernames, input errors, canceled context or invalid root fail safely without writes/panic.
- Flag callers never read stdin and retain default preview/apply/dry-run/duplicate/unknown/contradictory flag behavior.
- New stdin logic uses one reader for complete session and supports deterministic in-memory stdin tests.
- Existing .env and excluded files preserved in SYNTHETIC fixtures; original protected .env/.env.example must never be read/copied/edited here.
- Docs distinguish automatic LOCAL module derivation from remote GitHub repository creation and tool name from application identity.
- Disposable offline copied-source smoke exercises real no-arg CLI stdin, deny/default-no and confirm, renamed build/tests and repeat rejection; original worktree unchanged.

## Verification and privacy
Meaningful test-first RED/GREEN required for behavior. Parent ASSESS after writer; obey risk plan and independent verifier if required.
Known incidents from previous feature: original-root make implicitly reads private .env; blanket generated-cache filter dropped required internal/infra/cache source. Do not repeat either.
All functional runners that could load configuration run only in narrow disposable source copy including cache application package, excluding all original env files/example, hidden/Git/log/temp/odd/vendor/generated cache files and symlinks. Synthesized example anchor only `APP_NAME=nexokit` supplied by user. No CLI init/preview or make on original root.
Previous Redis-dependent subtests require localhost Redis and may skip; report exact skips. Hosted Actions, live database/provisioning/PASETO/production configuration and public licensing remain outside verification/scope.

## Evidence
- Writer `mv65xja4-c-6izp`: meaningful RED, valid no-argument sessions returned flag usage errors before interactive implementation; GREEN focused TestInit and slug/source-copy tests. Five focused packages commands/projectinit/generator/templates/testscli PASS; scoped formatting and whitespace checks PASS.
- Real synthetic copied-source CLI stdin smoke PASS: no/blank preserve identity/source; yes applies `github.com/user-name/my-api` + display name `My API`; help succeeds, repeat refused; renamed offline full build/tests pass with 120s subprocess bounds.
- Writer's initial fixture attempts failed (missing CLI identity anchor, required env.go wrongly excluded, tests clearing process env). Corrected and rerun PASS; no source regression established. Preserve source cache/env.go while excluding actual environment/generated files.
- Worker reported real-DB migrate/root tests and recursive nested source-copy smokes skipped; Redis tests passed. Independent outer skips: TestMigrateCommand_Up, TestCreateRootCommand_IdempotentRealDB, TestMigrateCommand_DBConnect (no database configuration). Nested run additionally skips TestInteractiveSourceCopy/TestInitializedSourceCopy only to prevent recursion; remaining unit tests run. Redis tests passed with available service; no services provisioned by agents, no hosted deployment checks.
- Authored diff 458 changed lines, above advisory review budget to preserve safety/runtime tests and docs; no multi-area product expansion.
- ASSESS unassessable on untracked paths, RDD off => high-risk separate verifier. `mv669tq9-d-p4jz` independently PASS: fresh secret-free source-copy full verbose tests (536 top-level PASS, zero FAIL), build/vet/global format/actual module guard. Real stdin no/blank preserved identity hashes; yes applied expected module/name, help and repeated-init rejection worked; renamed build/short tests passed. Flag caller with closed stdin and panic-on-read tests passed. Original Git status stable, no dependencies/planner changes.
- Work-unit commit: `1769aa8`, eight scoped files +499/-7 including tests/docs/tracking.
- Rollback boundary: interactive branch/input helpers/tests and matching adoption docs only; existing planner/flag functionality remains independently valid.

## Next step
Feature verified and committed, no blockers. Use bare `go run ./cmd/nexokit init` in a fresh application copy. Prompts request application name/GitHub username, derive module and default-No confirmation; no remote creation. Real PostgreSQL migrate/root/provisioning/PASETO/production configuration and hosted CI not verified. Synchronous blocked stdin cannot be immediately canceled; checks run before/after reads. Per-command smoke timeout 120s passed here but could be fragile on slow hosts. Pre-commit hook absent; no push/PR.
