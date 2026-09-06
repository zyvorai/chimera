# Overview KPIs

## Purpose

Show live gateway-derived health: requests, error rate, active sessions, bytes transferred, NFC exports, and average response latency.

## When to use it

- First screen after login — confirm Chimera is receiving infrastructure traffic
- Before and after scenario or fault changes to watch error rate and latency move
- When Transiva or `selftest` runs but you need a quick “is anything happening?” check

## How to get there

- Screen id: `#overview` (Command Center home)
- Nav: **Control Center → Overview** (sidebar ⌂) or land here immediately after login

## Operate from the console (UX)

1. Sign in at `/__chimera/` if the login gate is showing.
2. Read the six metric tiles: **Requests**, **Error Rate**, **Active Sessions**, **Data Transfer**, **Exports**, **Avg Response**.
3. Note trend arrows (↑/↓) on the second poll — dashboard polling under `/__chimera` is excluded from these counters, so flat zeros mean no client traffic yet.
4. Optionally change **Last 15 minutes** / **Last hour** in the page head (visual filter; telemetry still polls every ~2s).
5. Click **↻ Refresh** to force an immediate telemetry pull if you just started Transiva or `selftest`.
6. Use the top bar **Scenario** dropdown shortcut to scroll to Scenario Launcher when you want presets without hunting the page.

**Empty / fail:** All KPIs stay at zero after export tests → confirm clients hit `/sdk`, not `/__chimera/`. **Error Rate** climbs with red statuses in [Live requests](live-requests.md) — open Fault Studio or run **Clean Environment**. Health ring in the sidebar reads **Unavailable** → check `GET /__chimera/health` on the host.

**Success:** Requests and sessions increment during `selftest` or Transiva export; exports tick up on NFC lease starts; error rate stays near 0% on clean runs.

## Related pages

- [Live requests](live-requests.md)
- [Top Activity](activity.md)
- [Topology](topology.md)
- [Scenario Launcher](scenarios.md)
- [Getting Started](../../getting-started.md)
- [Page index](../../PAGE_INDEX.md)
