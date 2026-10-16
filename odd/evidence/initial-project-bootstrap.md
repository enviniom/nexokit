# Initial project bootstrap: acceptance evidence

The complete one-time setup satisfies S1–S7 in [the feature specification](../tasks/initial-project-bootstrap.md). Independent final verdict: **PASS, no blockers**. This is functional verification, not native review authority or permission to publish.

## Candidate and isolation

- Candidate: `6477fcd61f0c8eea26e7882a789762467eb48b61`.
- Work-unit chain: `db0100e` → `62d666f` → `6477fcd`.
- Independent logical verifier: `mv8r3kwb-e-eevs`, reconstructed as `mvaeoi6u-1-62it`; final evidence assessment: `mvaezgds-2-7d8d`.
- Fresh source: 517 committed inputs copied from scoped Git objects, with an invented environment example; no original protected environment access in this final verification.
- Source manifest unchanged; HEAD before/after equals the candidate. `failures`, `unavailable`, and `sourceDrift` are empty.
- Scratch: `/tmp/nexokit-initial-setup-verify-01a131b8/unit-c-verify/`.
- The parent executed the verifier-authored script once in background (job-2), exit 0 after 4m 1s. No timeout retries.

## Observed checks

| Check | Exit | Time |
|---|---:|---:|
| `go build ./...` | 0 | 50.90s |
| `go vet ./...` | 0 | 4.29s |
| `go test -v ./... -count=1 -timeout=600s` | 0 | 61.07s |
| `go test -race ./internal/cli/commands ./internal/cli/projectinit -count=1 -timeout=600s` | 0 | 99.02s |
| Public CLI, component DB configuration | 0 | 5.26s |
| Public CLI, kept database URL | 0 | 5.52s |
| Public `make init`, component DB configuration | 0 | 5.47s |
| Public `make init`, kept database URL | 0 | 5.41s |
| Real PTY: hidden DB/root input, one default-No confirmation, cancellation without writes | 0 | Passed |

The four public cases run `TestInitPublicIsolatedPostgres` against four distinct empty disposable databases. They assert migrations → seeds → root → completion, stored root, and repeated-init rejection. PostgreSQL used the cached `postgres:16-alpine` image with `--pull=never`, tmpfs storage, and loopback port 32768 rather than 5432. Owner-label cleanup removed the container with exit 0; no user database was used.

The parent additionally ran, in the same isolated copy:

```text
go test ./internal/cli/commands -run '^TestInitSetup(FileFailureStopsStagesAndCloses|AdoptedRejectsBeforeReadingInput|KeptEnvironmentLifecycle)$' -count=1 -timeout=20s
```

Exit 0, package time 0.014s. This checks safe publication diagnostics, early one-time rejection, and ordered lifecycle/failure handling.

## Independent verdict by requirement

| Spec | Verdict | Evidence |
|---|---|---|
| S1 | PASS | Build/vet/suite and four public setup cases; environment/template preservation and preview regressions |
| S2 | PASS | Both real Make setup cases create the stored root after migrations |
| S3 | PASS | Public phase-order assertions require seeds before root |
| S4 | PASS | Public repeated-init refusal and once-only README/setup/CLI documentation |
| S5 | PASS | Kept URL cases, backup/rebuild and snapshot regressions, real hidden-password PTY cancellation |
| S6 | PASS | Authorized synthetic copy, local cached image, isolated empty DBs, verified cleanup |
| S7 | PASS | Unpublished A→B→C work-unit ancestry and stable candidate |

## Skips and limitations

The ordinary full suite explicitly skipped these seven tests; they are not reported as executed:

- `TestBootstrapIsolatedPostgres`
- `TestInitPublicIsolatedPostgres`
- `TestMigrateCommand_Up`
- `TestCreateRootCommand_IdempotentRealDB`
- `TestMigrateCommand_DBConnect`
- `TestRedisCache`
- `TestRedisLimiterIntegration`

The separate four-case PostgreSQL run executes **only** `TestInitPublicIsolatedPostgres`. It does not independently execute the dedicated adapter test or legacy migration/root integration assertions. Redis tests were skipped because no Redis service was available; Redis remains optional for initial setup.

No transaction spans files, migrations, arbitrary seeds and root creation. Database changes may persist after a failed stage; adopted projects use explicit maintenance commands for recovery, not another init. File checks assume a quiescent local copy, not filesystem locks.

## Historical failures and privacy disclosure

The feature log retains observed RED failures and all intermediate tooling failures: wrong test-copy root, missing persisted database-creation diagnostics, process-environment leakage from old tests, tmpfs exhaustion, a missing cache path, and disappearance of the previous temporary root. The last event required explicit reconstruction consent; the final independent proof above uses fresh artifacts, not missing historical logs. The cause of the disappearance is unknown.

The earlier Unit-B verifier consulted original environment filename/size/mtime metadata despite the exclusion. This was disclosed as a violation, not a privacy PASS. No content read was reported. The reconstructed final verifier respected the full exclusion, including metadata.

Native assessment was unavailable (`package-local-binary-missing`); the candidate was treated as high risk and independently verified. No native receipt or review-authority claim is made. Hosted CI, production deployment and publishing were not performed.

## Temporary artifacts

`verify-full-summary.json`, `source-manifest-before.txt`, `head-tree.txt`, and `logs/` under the scratch root contain the detailed execution evidence. They are temporary and may disappear; this report and the feature log retain the acceptance result and limitations without copying private configuration.
