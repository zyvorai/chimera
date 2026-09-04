# Chimera Command Center UX

The Command Center is embedded directly into the Go executable and served at:

```text
/__chimera/
```

There is no React/Node runtime, package manager, CDN, web-font dependency or separate frontend process. The complete application lives in `internal/gateway/ui.go` and talks only to Chimera's local APIs.

## Visual direction

The UI follows an Apple TV Home–inspired layout: a sticky top nav, centered 980px content column, full-width hero on Home, and separate pages for Infrastructure, Inventory, Telemetry, and Lab. Typography uses the system stack (`-apple-system` / SF Pro), light surfaces, generous section padding, and tile cards that link between pages.

Design cues come from Apple.com product pages: large hero headlines, alternating white/gray sections, pill CTAs, and one focused surface per page instead of a cramped single dashboard.

## Main surfaces

### Six live KPIs

The overview header shows real gateway-derived values for:

- total infrastructure requests
- error rate
- active vSphere sessions
- transferred response bytes
- NFC/export starts
- average response latency

Dashboard polling under `/__chimera` is excluded from these counters, so Chimera does not inflate its own metrics.

### Infrastructure topology

A visual datacenter → cluster → host → datastore map gives the operator an immediate mental model of the current simulator estate. Labels are hydrated from the configured simulator counts and fixture size. "View Full Topology" expands the same live diagram into a full-viewport overlay (dismissed via its own button, clicking outside it, or Escape); the toolbar's `Fit`/`+` controls reset or step up the zoom level.

### Top activity

A live donut groups infrastructure traffic into gateway operation classes such as vSphere SOAP, vSphere SDK, NFC Download and NFC Resume.

### Live requests

The request feed shows method, compact path, HTTP status and end-to-end gateway duration for the newest infrastructure requests. "View All" toggles the visible row count between 10 and 30 (still within the server's buffered request history, no extra API call), flipping its own label between "View All" and "View Less." The topbar bell icon (badge driven by the current error count) scrolls straight to this panel — useful for jumping to failing calls shown as red status codes.

### VM inventory

The inventory table supports:

- search by VM name
- power-state filter
- pagination
- CPU, memory, disk and datastore columns
- a Fixture column showing whether each VM's export uses a real VMDK (`Matched`, `Round-robin`, `Shared`, `Manual`) or the default generated synthetic fixture (`Synthetic`), hover for the filename
- export target selection/copy action

Rows reflect the real govmomi simulator inventory (name, power state, datastore), not a fabricated count.

### VMDK Library

A dedicated card lists every `.vmdk` file found under `fixture_vmdk_dir` and any configured `fixture_vmdk_dirs` (recursing into subdirectories), with its size and which VM (if any) it was assigned to — including files that matched no VM, which don't otherwise show up anywhere in the VM-keyed inventory table. The directories are re-scanned automatically every 5 seconds, so adding or removing a file anywhere under them shows up here without restarting the server. See `docs/TRANSIVA.md`'s "Testing conversion, not just transfer" section for the matching rules.

The "Upload VMDK" button opens a modal with two ways to add a fixture without touching the host's shell:

- **Upload** a `.vmdk` from the browser straight into `fixture_vmdk_dir`.
- **Browse** files already staged on the host under any configured fixture root and pick one — this only ever lists what's inside those specific directories, never an arbitrary host path.

Either path can optionally pin the file to a specific VM, which overrides the automatic name/round-robin matching (shown as the `Manual` fixture badge) until cleared.

### Provider personas

The left rail treats infrastructure implementations as first-class personas. **vSphere** is active in Command Center today. **Nutanix Prism**, **Hyper-V**, **AWS**, and **Azure** are available as protocol personas (`CHIMERA_PERSONA=nutanix|hyperv|aws|azure`) but do not mount the Command Center yet — start a separate process for those endpoints. Proxmox VE and OpenStack remain standby roadmap providers.

### Scenario Launcher

One-click deterministic presets:

- Clean Environment
- Slow Fabric
- Flaky API
- Resume Export

### Fault Studio

A right-side administrative drawer controls:

- latency
- failure status
- next-N generic request failures
- next-N NFC failures
- one-shot NFC stream drops
- stream drop offset
- bandwidth cap

The drawer can be opened from the top command bar's gear icon, sidebar or scenario panel.

### Health and engine status

The sidebar health ring derives a simple health score from current error rate and shows uptime and version. The footer shows the active persona, public SDK endpoint, simulator administrator identity and local system time.

## Authentication model

The dashboard is a **full-page login gate**: without a valid session, the browser shows only a dedicated login screen (username/password, default `admin`/`admin`) and never renders or fetches dashboard data — no flash of inventory/telemetry content before the login check runs. This is a frontend-only gate. The underlying `/__chimera/*` read endpoints (health, bootstrap, inventory, telemetry, vmdks) remain reachable without a token exactly as before, so `deploy-remote.sh`'s health check, `transiva-smoke.sh`, and anything else scripting against the API directly are unaffected — only the browser UI's behavior changed.

Logging in calls `POST /__chimera/login`, which checks the credentials (constant-time comparison, no lockout/rate-limiting — this is a test-lab tool, not a hardened auth system) and hands back the same admin bearer token every protected write endpoint has always used; the dashboard then sends it as `Authorization: Bearer <token>` and loads the dashboard for the first time. Logging out (via the topbar user menu) clears the token and returns to the login page rather than leaving dashboard chrome visible behind a reopened dialog.

The browser stores the token only in `sessionStorage`, so it disappears when the browser session closes — reloading with a still-valid token goes straight to the dashboard, while a stale/expired one falls back to the login page. The admin login itself (not the token) can be changed live from the Settings panel, or via `CHIMERA_ADMIN_USERNAME`/`CHIMERA_ADMIN_PASSWORD` at startup — see the README's Configuration table.

### Settings panel

Opened from the sidebar's "Settings" or "Users & Auth" entries (both point at the same panel — Chimera has one shared admin login, not a multi-user directory). Shows the configured listen address read-only (changing it needs `CHIMERA_LISTEN` and a restart) and a form to change the admin username/password live, no restart needed.

## Backend endpoints used by the UX

```text
GET  /__chimera/health
GET  /__chimera/api/bootstrap
GET  /__chimera/api/inventory
GET  /__chimera/api/vmdks
GET  /__chimera/api/telemetry
POST /__chimera/login
GET  /__chimera/state
POST /__chimera/faults
POST /__chimera/reset
POST /__chimera/scenario/{clean|slow|flaky|resume}
POST /__chimera/admin/credentials
POST /__chimera/api/vmdks/upload
GET  /__chimera/api/vmdks/browse
POST /__chimera/api/vmdks/assign
```

## Responsive behavior

- Topology, Top Activity and Live Requests each render as a full-width stacked panel rather than competing for a narrow column, so every panel gets real reading room at any desktop width
- the sidebar can be manually folded to an icon-only rail via its own toggle (persisted in `localStorage`), independent of viewport size
- single-column workbench on tablets
- collapsed icon-only sidebar (automatic, not just the manual fold) and simplified controls on small screens

## Future UX additions

- Command Center shells for Nutanix / Hyper-V / AWS / Azure (today those personas are protocol-only)
- in-UI persona activation / hot switching
- drag-and-drop synthetic topology builder
- request trace and SOAP/REST inspector
- lease/export progress visualization
- inventory CRUD
- saved scenario composer
- provider-specific templates
- test run history and pass/fail reports
- OpenTelemetry trace exploration
