# Architecture decision records

Record why a consequential design was chosen; keep day-to-day implementation rules in the four core guides. ADRs explain tradeoffs, not duplicate tutorials.

## Location and lifecycle

The planned destination is [docs/adr/](adr/). This directory has not been created and no ADR records exist yet. Use sequential names such as `0001-short-title.md`, not year-based numbering.

| Status | Meaning |
|---|---|
| Proposed | Awaiting review |
| Accepted | Approved decision |
| Superseded | Replaced; link its successor |
| Deprecated | No longer applicable |

Keep accepted records as history; add a successor when the decision changes. Update living guides to reflect accepted outcomes and actual rollout status.

## Template

```markdown
# ADR-NNNN: Short title

- Status: Proposed
- Date: YYYY-MM-DD
- Authors: Name/role
- Related changes: links or identifiers

## Context
What problem and constraints require a decision?

## Options
What realistic alternatives were considered?

## Decision
What was chosen, why, and which invariants does it establish?

## Consequences
Benefits, costs, security/operational risks and rejected tradeoffs.

## Rollout
Compatibility, migration, tests and rollback; distinguish target from current state.
```

Use an ADR for consequential dependency, tenancy, persistence or deployment changes. Small fixes need not manufacture a design decision record.
