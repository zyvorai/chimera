# Login

## Purpose

Authenticate the browser session before any Command Center chrome or data is shown.

## When to use it

- Every new browser tab or after **Log out** from the top-right user menu
- Before Scenario Launcher, Fault Studio writes, VMDK upload, or credential changes

## How to get there

- URL: `http(s)://<host>:8989/__chimera/`
- Screen id: login gate (full page — no sidebar until authenticated)

## Operate from the console (UX)

1. Open `http(s)://<host>:8989/__chimera/` in a browser.
2. Confirm you see **Log in to Chimera** — not the overview metrics (unauthenticated users never flash inventory).
3. Enter username and password (default **admin** / **admin** unless changed in Settings or via `CHIMERA_ADMIN_*`).
4. Click **Log in** (or press Enter in the password field).
5. Wait for the toast **Logged in — Scenarios and fault injection are enabled.** The sidebar, KPI row, and inventory load.

**Empty / fail:** **Login failed** toast → verify credentials in Settings (if you still have another session) or env vars on the host. Reload loops usually mean wrong password or stale token — use **Log out** from the user chip, then sign in again.

**Success:** Full Command Center shell appears; sidebar health ring shows uptime/version; footer shows active vSphere persona and SDK endpoint.

Console: `http(s)://<host>:8989/__chimera/` · vSphere SDK: `http(s)://<host>:8989/sdk` · Never publish lab IPs — use `<host>`.

## Related pages

- [Settings](settings.md)
- [Overview KPIs](overview.md)
- [Admin basics](../../admin-basics.md)
- [Getting Started](../../getting-started.md)
- [Page index](../../PAGE_INDEX.md)
