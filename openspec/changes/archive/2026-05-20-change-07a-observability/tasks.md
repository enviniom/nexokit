# Tasks: Observability Infrastructure — Health Endpoints

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 200–265 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR (within budget) |
| Delivery strategy | auto-chain |
| Chain strategy | feature-branch-chain |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: feature-branch-chain
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Health types, interfaces, handlers + tests | PR 1 (feature/change-07a-observability) | Base = feature/tracker branch; self-contained with table-driven tests |
| 2 | Router wiring + bootstrap injection + regression tests | PR 2 (feature/change-07a-observability-router) | Base = PR 1 branch; wires health into existing flow |

## Phase 1: Foundation — Health Types and Interfaces

- [x] 1.1 Create `internal/server/health.go` with `HealthDeps` struct, `dbPinger` interface (`PingContext`), and `cacheGetter` interface (`Get`).
- [x] 1.2 Add response structs in `health.go`: `LiveResponse` (`Status string`) and `ReadyResponse` with `Dependencies` slice of `DepStatus{Name, Status, Error}`.

## Phase 2: Core Implementation — Health Handlers (TDD: RED → GREEN)

- [x] 2.1 **RED**: Write table-driven unit test `TestLiveHandler` in `internal/server/health_test.go` — assert `liveHandler` returns 200 with `{"status":"alive"}` for all cases.
- [x] 2.2 **GREEN**: Implement `liveHandler(c *gin.Context)` in `health.go` — writes JSON `{"status":"alive"}` with 200, no I/O.
- [x] 2.3 **RED**: Write table-driven unit test `TestReadyHandler` in `health_test.go` — cases: all healthy, DB fail, cache fail, cache disabled, nil DB, nil cache when enabled.
- [x] 2.4 **GREEN**: Implement `readyHandler(deps HealthDeps)` in `health.go` — pings DB via `PingContext`, probes cache via `Get` only when `CacheEnabled`, aggregates `DepStatus` records, returns 200 or 503.
- [x] 2.5 **RED**: Write unit test `TestHealthHandler` in `health_test.go` — assert existing `/health` envelope (`success=true`, `"API is healthy"`, `data.status="ok"`).
- [x] 2.6 **GREEN**: Implement `healthHandler(c *gin.Context)` in `health.go` — calls `response.Success(c, messages.MsgHealthy, map[string]string{"status":"ok"})`.

## Phase 3: Integration — Router Wiring and Bootstrap Injection

- [x] 3.1 Modify `internal/server/router.go`: update `NewRouter` signature to accept `HealthDeps`; register `/health`, `/health/live`, `/health/ready` on base engine before `/api/v1` group.
- [x] 3.2 Modify `internal/app/bootstrap.go`: extract `*sql.DB` from GORM via `database.DB()`, derive `CacheEnabled` from `cfg.Cache.Driver != "none"`, construct `HealthDeps`, pass to `NewRouter`.
- [x] 3.3 Update `internal/server/server_test.go`: fix `TestHealthEndpoint` to use new `NewRouter` signature; add `TestLiveEndpoint`, `TestReadyEndpoint` via `httptest` against `NewRouter`.
- [x] 3.4 Add regression test `TestNotFound` still passes with new router setup (verify existing test unchanged).

## Phase 4: Verification

- [x] 4.1 Run `go build ./...` — zero compilation errors.
- [x] 4.2 Run `go test ./...` — all tests pass, including new health tests and existing regression.
- [x] 4.3 Run `go vet ./...` — zero issues.
