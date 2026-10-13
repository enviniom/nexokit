# Starter bootstrap

## Objective
Make a local copy of NexoKit straightforward to initialize as an independent Go application today. Keep GORM, existing modules and developer CLI. No ongoing upstream synchronization promise.

## Authorized scope and rationale
User explicitly authorized completing starter preparation and committing verified work. Permissions prerequisite was verified and committed in `a662622`; evidence recorded in `1c7343d`.

The missing adoption step is safe application/module identity initialization. Go module renaming must cover imports, embedded templates, seed runner and generated test sources, not only go.mod. Tool identity (`cmd/nexokit` and CLI name) may remain NexoKit; app identity and module path must change without blindly replacing all provenance.

## Constraints
- One-time local initializer, no downloader/updater/sqlc/framework refactor.
- Explicit preview then apply; validate all planned changes before writing.
- Only known text surfaces; no .git/local .env/logs/vendor/temp/odd/task data or secrets touched.
- Preserve existing .env, go.sum/dependencies and module/domain/security behavior.
- Do not rename cmd/nexokit merely for cosmetic identity; documentation clearly distinguishes tool vs app.
- Avoid self-mutating initializer canonical template identity when rewriting import paths; repeated apply must safely reject already initialized project.
- No production secret generation, automatic DB provisioning or automatic seed/root-account execution.
- License is an owner decision; do not invent or assign terms. Local private reuse readiness is distinct from public distribution licensing.
- Keep writes single-threaded, commits scoped and verified; no push or PR.

## Tasks
- [ ] S1 — Implement safe local `init` CLI with preview/apply, identity rewriting, deterministic tests, adoption docs and renamed-copy smoke verification. **In progress: independent verification passed, privacy incident clarified and isolated guard passed; preparing scoped commit.**
- [ ] S2 — Enable automatic starter CI and clear known formatting baseline; validate full source and initialized copy, then commit. **Pending.**

## S1 acceptance
- `go run ./cmd/nexokit init --name <name> --module <module>` previews without writes; `--apply` applies deliberately; `--dry-run` explicit preview accepted; contradictory flags rejected.
- Invalid root/name/module, symlink escapes, unsafe/partial/repeated initialization rejected without unintended edits.
- go.mod/imports/embedded templates/seed runner/test fixture strings agree with new module; tests compile under new module without retrieving original project.
- README heading and APP_NAME example use requested app identity when the adopter runs init, while tool attribution/documentation retained. User supplied original example anchor `APP_NAME=nexokit`; do not read/copy/edit the original `.env.example` during development. Disposable smoke fixtures synthesize only this provided line; full configuration example inspection is not verified.
- All candidates preflighted and plan reports changed paths plus relevant remaining references; apply reports next configuration/check commands.
- Existing .env and excluded local files byte-identical; dependencies/go.sum untouched; no shell/download/Git side effects.
- Focused tests and disposable clean-copy init/help/build/full-test smoke succeed using cached dependencies offline when available.

## S2 acceptance
- Existing CI runs on push and pull requests as well as workflow_dispatch.
- Tests/vet/format/module boundary checks pass; baseline onboarding formatting fixed as mechanical readiness cleanup.
- New-project flow documentation describes PostgreSQL/configuration/migrations/root creation and smoke checks with no production-ready default claim.
- Public licensing remains explicitly unselected, not silently assigned.

## Checks and evidence
- Permissions prerequisite: full tests/build/vet/module guard passed; no endpoint-specific root override; user/system role grant policy as clarified.
- Initial writer stopped before edits: tool safety policy denied `read .env.example`; no tests/RED/GREEN or source writes performed. Do not use alternate commands to bypass denial.
- User selected resolving permission before implementation. Separate read-only tooling diagnosis found no documented per-file grant mechanism; Pi `/trust` does not authorize file access. Specific denial owner remains unknown. No policy changed and no denied read retried. User subsequently supplied `APP_NAME=nexokit` as the requested nonsecret fragment; this provides the identity anchor, not tool access to the original file.
- Writer `mv63mdpe-6-tl3y`: RED observed through not-implemented planner/command behavior and partial-source regression; GREEN focused projectinit/commands and generator/templates/CLI suites. Changed Go formatted; diff check passed.
- Writer synthetic renamed-copy smoke passed: read-only preview, explicit apply to `example.com/team/starter-smoke`, name `Starter Smoke`, help, repeat rejection, no original imports, offline full tests/build. Source whitelist excludes original environment/example, Git/odd/log/tmp/cache state; synthetic example contains only supplied APP_NAME line.
- Size: 917 new Go lines plus 109 tracked insertions/12 deletions, including safety/rollback and smoke tests; over advisory review-size heuristic, no added product scope.
- ASSESS returned `unassessable` because untracked files require explicit declaration; RDD off, conservative high-risk separate verifier required. Native review not started; do not interpret failure as low risk.
- Independent verifier `mv645u9j-7-ydtv`: PASS, no candidate-caused blockers; fresh focused/full tests, verbose renamed-copy smoke, build/vet and whitespace checks passed. Candidate Go formatted; onboarding baseline remains S2. Synthetic smoke correctly excluded original examples/environment and mutated only temporary copy.
- Privacy evidence incident: verifier ran `make check-module-errors` on original root while Makefile has `-include .env` and directory listing confirms private `.env` exists. Thus make implicitly read it; blanket no-private-env-read assertion is incorrect. No contents appeared in reported output and no file modification was reported. Do not repeat this command on original root. Separate verifier continuation `mv64b0ga-8-oher` confirmed that implicit read, retracted blanket privacy claim, and reran guard successfully in a narrowly copied disposable tree containing only Makefile/internal/modules, no env/hidden files or symlinks. Output exposed no env values. Test/build/vet root-env absence is structurally supported (package working directories and temporary chdir targets), not an exhaustive privacy proof. No source changes. Temporary isolated guard fixture `/tmp/nexokit-guard.msclaI` remains outside repository.
- S1/S2 commits: pending.
- Known baseline format debt: internal/modules/onboarding/slices/onboard_company/handler_test.go, to be fixed in S2.
- Existing user .env is private and will not be read or copied.

## Next step
Incident diagnosis complete and isolated guard passed. Commit S1 with tests/docs and record identity, then implement S2 automatic CI and baseline format cleanup. Verify uncertain commands in secret-free source copy only; no original env reads, including implicit Makefile loads, are permitted. Keep real-example/deployment limitations explicit. Pre-commit hook absence checked before commit. No push/PR.
