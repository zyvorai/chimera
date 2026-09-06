# Live requests

## Purpose

Stream recent infrastructure HTTP calls: method, compact path, HTTP status, and end-to-end gateway duration.

## When to use it

- Debug govmomi / Transiva traffic and spot failing calls (red status codes)
- Correlate Fault Studio or scenario changes with immediate API/NFC errors
- Audit-style tail of the newest simulator-facing requests (also linked as **Sessions** / **Audit Logs** in the sidebar)

## How to get there

- Screen id: `#requests`
- Nav: **Control Center → Sessions** or **Audit Logs** (sidebar), or **Live Requests** panel on Overview
- Shortcut: top bar bell (badge = current error count) scrolls directly here

## Operate from the console (UX)

1. Sign in and open **Live Requests** (or click the bell when the badge is non-zero).
2. Watch rows populate: **GET**/**POST** method chip, truncated path, green status (or red for ≥400).
3. Compare duration column (ms) before/after **Slow Fabric** scenario or Fault Studio latency.
4. Click **View All** to expand from 10 to 30 buffered rows (label flips to **View Less**); no extra API call — same server buffer.
5. When debugging resume exports, look for NFC paths returning **206** after **Resume Export** scenario.
6. Cross-check with [Fault Studio](fault-studio.md) when statuses match your injected `fail_status`.

**Empty / fail:** **Waiting for client traffic…** during an export → client may be pointed at wrong host/port or TLS mismatch. All red rows → run **Clean Environment** scenario, then **Reset** in Fault Studio. Bell badge stuck high after clean run → click **↻ Refresh** on Overview.

**Success:** SOAP/SDK/NFC paths appear with 2xx/206 statuses on happy path; error badge clears after clean reset.

## Related pages

- [Fault Studio](fault-studio.md)
- [Overview KPIs](overview.md)
- [Top Activity](activity.md)
- [Workflows](../../workflows.md)
- [Page index](../../PAGE_INDEX.md)
