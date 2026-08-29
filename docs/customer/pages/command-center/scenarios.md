# Scenario Launcher

## Purpose

One-click deterministic presets for clean, slow, flaky, and resume-export fault paths — without hand-tuning Fault Studio each run.

## When to use it

- Repeatable Transiva or client hardening tests
- Reset to baseline (**Clean Environment**) between scenarios
- Simulate NFC drop mid-transfer for Range/206 resume behavior

## How to get there

- Screen id: `#scenarios`
- Nav: **Control Center → Scenarios** (sidebar ◈) or top bar **Scenario** button (scrolls here)

## Operate from the console (UX)

1. Sign in — scenarios POST to `/__chimera/scenario/*` and require the admin bearer token.
2. Scroll to **Scenario Launcher** (right column on Overview, below VM inventory on wide layouts).
3. Click a card:
   - **Clean Environment** — reset faults and counters (happy path baseline).
   - **Slow Fabric** — high latency and low bandwidth pressure.
   - **Flaky API** — random failures and timeouts on API calls.
   - **Resume Export** — drop NFC connection mid-transfer (~2 MiB) for Range retry / HTTP 206.
4. Active card highlights with an inset ring; toast shows **Scenario applied** with the preset name.
5. Open [Fault Studio](fault-studio.md) to inspect the underlying latency/failure fields the preset set.
6. Click **View Fault Studio →** at the bottom of the card for the same drawer.
7. Re-run Transiva export or `./bin/chimera selftest`, then watch [Live requests](live-requests.md) and KPI error rate.

CLI equivalent (same token as API): `./scripts/scenario.sh clean|slow|flaky|resume`

**Empty / fail:** Click does nothing / login page returns → session expired; sign in again. **Scenario failed** toast → check engine logs; verify `Authorization` path with `CHIMERA_ADMIN_TOKEN` if using curl. Resume test never shows 206 → confirm client sends Range headers (see [Workflows](../../workflows.md)).

**Success:** Toast confirms preset; Fault Studio summary chips update; traffic reflects scenario (latency, errors, or dropped NFC) until **Clean Environment** or **Reset**.

## Related pages

- [Fault Studio](fault-studio.md)
- [Live requests](live-requests.md)
- [Overview KPIs](overview.md)
- [Workflows](../../workflows.md)
- [Page index](../../PAGE_INDEX.md)
