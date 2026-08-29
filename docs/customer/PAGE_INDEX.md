# Chimera — Complete page index

Every primary navigable dashboard route.

_Generated: 2026-08-29 · 10 routes_

Regenerate: `node scripts/customer-docs/generate-page-index.mjs`

## AUTH

| Page | Route | Purpose | Guide |
|------|-------|---------|-------|
| Login | `/login` | Authenticate the browser session before Command Center chrome or live data loads. | [Open](pages/command-center/login.md) |

## COMMAND CENTER

| Page | Route | Purpose | Guide |
|------|-------|---------|-------|
| Overview KPIs | `/overview` | Six live KPIs — requests, error rate, sessions, bytes, exports, latency — refreshed every two seconds. | [Open](pages/command-center/overview.md) |
| Topology | `/topology` | Datacenter → cluster → host → datastore map of the configured simulator estate. | [Open](pages/command-center/topology.md) |
| Top Activity | `/activity` | Donut and legend grouping infrastructure traffic by operation class (SOAP, SDK, NFC). | [Open](pages/command-center/activity.md) |
| Live requests | `/live-requests` | Newest infrastructure HTTP calls with method, path, status, and gateway duration. | [Open](pages/command-center/live-requests.md) |
| VM inventory | `/vm-inventory` | Searchable, paginated simulator VM table with power state, sizing, datastore, and fixture badge. | [Open](pages/command-center/vm-inventory.md) |
| VMDK Library | `/vmdk-library` | Fixture `.vmdk` inventory — upload from browser or pick staged host files; optional VM pin. | [Open](pages/command-center/vmdk-library.md) |
| Scenario Launcher | `/scenarios` | One-click presets: clean, slow fabric, flaky API, resume export. | [Open](pages/command-center/scenarios.md) |
| Fault Studio | `/fault-studio` | Administrative drawer for latency, status codes, next-N failures, NFC drops, bandwidth cap. | [Open](pages/command-center/fault-studio.md) |
| Settings | `/settings` | Read-only listen address plus live admin username/password rotation (no restart). | [Open](pages/command-center/settings.md) |

## Related

- [Customer docs home](README.md)
- [Page-by-page guides](pages/README.md)
