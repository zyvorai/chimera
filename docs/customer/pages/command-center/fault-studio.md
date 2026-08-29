# Fault Studio

## Purpose

Operator drawer for latency, HTTP failure status, next-N API/NFC failures, one-shot NFC stream drops, drop offset, and bandwidth caps.

## When to use it

- Fine-grained fault injection beyond Scenario Launcher presets
- Hardening Transiva or other clients against partial transfers and API errors
- Inspecting and clearing active fault policy before clean benchmark runs

## How to get there

- Drawer id: `#faultDrawer` (not a route — slides over the console)
- Nav: **Administration → Fault Studio** (sidebar ϟ), top bar **Fault Studio** (bolt icon), or **View Fault Studio →** on Scenario Launcher
- **Settings** (gear icon) is a separate drawer — do not confuse with Fault Studio

## Operate from the console (UX)

1. Sign in — mutating fault APIs require the admin session token.
2. Open Fault Studio from sidebar, top bar bolt, or scenario panel link.
3. Read summary chips at top: **Latency**, **Bandwidth**, **Status** (current policy snapshot).
4. Set fields as needed:
   - **Latency (ms)** — artificial delay on infrastructure requests.
   - **Failure HTTP status** — status returned for armed failures (default 503).
   - **Fail next requests** — count of generic API failures before auto-clear.
   - **Fail next NFC requests** — NFC-specific failure budget.
   - **Drop next NFC streams** / **Drop after (MiB)** — mid-transfer disconnect for resume testing.
   - **Bandwidth cap (MiB/s, 0 = unlimited)** — throttle NFC throughput.
5. Click **Apply policy** — toast **Fault policy active**; watch [Live requests](live-requests.md) for red statuses.
6. Click **Reset** for a clean traffic path without reloading the page (same as Clean scenario family).
7. **Unlock** / **Lock** toggles admin token in this tab; **API Tokens** sidebar entry uses the same control.
8. Close drawer with **×** or click the dimmed overlay.

**Empty / fail:** Drawer opens but **Apply** sends you to login → re-authenticate. Changes seem ignored → confirm toast success; some counters are consume-on-use (next-N). Bandwidth shows **∞** when cap is 0.

**Success:** Summary chips match inputs; injected errors appear in live request feed; **Reset** restores green statuses and nominal latency.

## Related pages

- [Scenario Launcher](scenarios.md)
- [Live requests](live-requests.md)
- [Overview KPIs](overview.md)
- [Workflows](../../workflows.md)
- [Admin basics](../../admin-basics.md)
- [Page index](../../PAGE_INDEX.md)
