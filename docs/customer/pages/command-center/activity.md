# Top Activity

## Purpose

Donut chart and legend grouping infrastructure traffic by operation class — vSphere SOAP, SDK, NFC download, NFC resume, and related gateway paths.

## When to use it

- See *what kind* of traffic dominates during Transiva export or discovery runs
- Confirm NFC vs SOAP mix after resume/retry scenarios
- Quick sanity check when KPIs move but you need operation-level breakdown

## How to get there

- Screen id: `#activity`
- Nav: **Control Center → Telemetry** (sidebar ∿) or scroll to **Top Activity** on Overview

## Operate from the console (UX)

1. Sign in and scroll to the **Top Activity** card below the topology panel.
2. Read the center total — it tracks overall request count from telemetry (same family as the Requests KPI).
3. Inspect colored legend rows: each line shows operation name, count, and percentage share.
4. Run `./bin/chimera selftest` or a Transiva export, then watch segments populate within ~2s polling cycles.
5. If the donut is a flat ring with **No traffic yet.**, drive clients against `/sdk` — dashboard fetches do not count.

**Empty / fail:** **No traffic yet.** persists during active exports → verify Transiva `vcenter_url` points at `<host>:8989/sdk`, not localhost from a remote runner. Legend shows one class at 100% unexpectedly → check Fault Studio for forced failures skewing retries.

**Success:** Multiple operation classes appear with plausible percentages; total matches rising **Requests** KPI during a self-test or export.

## Related pages

- [Overview KPIs](overview.md)
- [Live requests](live-requests.md)
- [Scenario Launcher](scenarios.md)
- [Workflows](../../workflows.md)
- [Page index](../../PAGE_INDEX.md)
