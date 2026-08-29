# Page-by-page guides

Each guide follows: Purpose → When to use it → How to get there → Operate from the console (UX) → Related pages.

Every route is also listed in the [complete page index](../PAGE_INDEX.md).

## Command Center

| Page | What it covers |
|------|----------------|
| [Top Activity](command-center/activity.md) | Donut chart and legend grouping infrastructure traffic by operation class — vSphere SOAP, SDK, NFC download, NFC resume, and related gateway paths. |
| [Fault Studio](command-center/fault-studio.md) | Operator drawer for latency, HTTP failure status, next-N API/NFC failures, one-shot NFC stream drops, drop offset, and bandwidth caps. |
| [Live requests](command-center/live-requests.md) | Stream recent infrastructure HTTP calls: method, compact path, HTTP status, and end-to-end gateway duration. |
| [Login](command-center/login.md) | Authenticate the browser session before any Command Center chrome or data is shown. |
| [Overview KPIs](command-center/overview.md) | Show live gateway-derived health: requests, error rate, active sessions, bytes transferred, NFC exports, and average response latency. |
| [Scenario Launcher](command-center/scenarios.md) | One-click deterministic presets for clean, slow, flaky, and resume-export fault paths — without hand-tuning Fault Studio each run. |
| [Settings](command-center/settings.md) | Show the configured listen address (read-only) and change Command Center admin username/password live — no Chimera restart required. |
| [Topology](command-center/topology.md) | Visual datacenter → cluster → host → datastore map of the simulated estate. |
| [VM inventory](command-center/vm-inventory.md) | Searchable, paginated table of simulator VMs with power state, CPU/memory/disk, datastore, and Fixture badge (real VMDK vs synthetic). |
| [VMDK Library](command-center/vmdk-library.md) | List and manage `.vmdk` fixtures under configured fixture directories; upload from the browser or pick staged host files; optionally pin to a VM. |

---

10 guides. Regenerate: `node scripts/customer-docs/generate-guide-index.mjs`.
