# Page-by-page guides

Each guide follows: Purpose → When to use it → How to get there → Operate from the console (UX) → Related pages.

Every route is also listed in the [complete page index](../PAGE_INDEX.md).

## Command Center

| Page | What it covers |
|------|----------------|
| [Fault Studio](command-center/fault-studio.md) | Operator drawer for latency, status codes, next-N API/NFC failures, stream drops, drop offset, and bandwidth caps. |
| [Live requests](command-center/live-requests.md) | Stream recent infrastructure HTTP calls: method, path, status, duration. |
| [Login](command-center/login.md) | Authenticate the browser session before any Command Center chrome or data is shown. |
| [Overview KPIs](command-center/overview.md) | Show live gateway-derived health: requests, error rate, sessions, bytes, exports, latency. |
| [Scenario Launcher](command-center/scenarios.md) | One-click deterministic presets for clean, slow, flaky, and resume export paths. |
| [Settings](command-center/settings.md) | Show listen address and change Command Center admin username/password live (no restart). |
| [Topology](command-center/topology.md) | Visual datacenter → cluster → host → datastore map of the simulated estate. |
| [VM inventory](command-center/vm-inventory.md) | Searchable, paginated table of simulator VMs with power state, CPU/memory/disk, datastore, and Fixture badge. |
| [VMDK Library](command-center/vmdk-library.md) | List and manage `.vmdk` fixtures under configured fixture directories; upload or browse; optionally pin to a VM. |

---

9 guides. Regenerate: `node scripts/customer-docs/generate-guide-index.mjs`.
