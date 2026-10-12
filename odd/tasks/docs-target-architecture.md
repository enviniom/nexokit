# Documentation target architecture

## Objective
Make docs/ describe intended structure, responsibilities, and contracts rather than implementation progress or missing wiring.

## Scope and constraints
- Documentation only; preserve existing English conventions and links.
- Preserve necessary operational/security caveats without implying target requirements are already implemented.
- No source changes or publishing. User explicitly authorized commits after verification.
- Reuse the completed read-only audit of all 18 Markdown files.

## Tasks
- [x] T1 (verified by writer): Rewrite the audited architecture, starter, and reference passages; align navigation and necessary runtime exceptions.
- [x] T2 (verified): Independently verify philosophy consistency, Markdown links, and diff integrity.
- [x] T3 (completed): Reconcile evidence and report outcomes and remaining limitations.

## Acceptance and checks
- Guides specify target design; no incidental progress inventory.
- Necessary operational caveats remain explicit and distinguishable from normative requirements.
- Existing technical contracts are preserved; no new architecture is invented.
- Read-back, local link check, targeted status-language scan, and git diff --check.
- Automated behavior tests: N/A, passive documentation only, no meaningful RED/GREEN.

## Evidence and continuity
- Initial worktree clean on main.
- Prior audit inspected 18 Markdown files and identified clear passages in nine files.
- Branch: docs/target-architecture.
- Documentation work-unit commit: 0ce3209 (docs: describe target architecture instead of implementation progress).
- Tracking evidence is committed separately; user authorized both commits. No push.
- Engram mirror: saved under project nexokit, observation 4146.
- T1: 12 authorized docs changed; writer diff check and 110 local links passed; no executable behavior changed.
- Native assessment unavailable due to untracked task file declaration; independent verification completed with PASS.
- T2: independent diff checks passed; docs local links 110/110 valid, root README included 121/121. No severe findings. Heuristic anchor validator, not a full Markdown renderer.
- Historical context.md has two intentionally stale links outside scope (unchanged); code-fence false positives excluded from docs results.
- Preserved optional legacy heading anchor for incoming compatibility; target cache semantics distinguished from operational caveat.

## Next step
Documentation adjustments and independent verification complete. Local commits prepared for user review; publishing remains the user's decision.
