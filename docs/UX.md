# Chimera Command Center UX

The Command Center is embedded directly into the Go executable and served at:

```text
/__chimera/
```

There is no React/Node runtime, package manager, CDN, web-font dependency or separate frontend process. The complete application lives in `internal/gateway/ui.go` and talks only to Chimera's local APIs.

## Visual direction

The UI is intentionally closer to a production infrastructure control plane than a developer mock: dense dark navigation, high-contrast telemetry, a topology workbench, compact data tables, scenario cards, and a slide-over Fault Studio. The generated design reference used for this iteration is included at `docs/chimera-ux-reference.png`.

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

A visual datacenter → cluster → host → datastore map gives the operator an immediate mental model of the current simulator estate. Labels are hydrated from the configured simulator counts and fixture size.

### Top activity

A live donut groups infrastructure traffic into gateway operation classes such as vSphere SOAP, vSphere SDK, NFC Download and NFC Resume.

### Live requests

The request feed shows method, compact path, HTTP status and end-to-end gateway duration for the newest infrastructure requests.

### VM inventory

The inventory table supports:

- search by VM name
- power-state filter
- pagination
- CPU, memory, disk and datastore columns
- a Fixture column showing whether each VM's export uses a real VMDK (`Matched`, `Round-robin`, `Shared`) or the default generated synthetic fixture (`Synthetic`), hover for the filename
- export target selection/copy action

Rows reflect the real govmomi simulator inventory (name, power state, datastore), not a fabricated count.

### VMDK Library

A dedicated card lists every `.vmdk` file found under `CHIMERA_FIXTURE_VMDK_DIR` (when configured), with its size and which VM (if any) it was assigned to — including files that matched no VM, which don't otherwise show up anywhere in the VM-keyed inventory table. See `docs/TRANSIVA.md`'s "Testing conversion, not just transfer" section for the directory-matching rules.

### Provider personas

The left rail treats infrastructure implementations as first-class personas. vSphere is active today; Nutanix Prism, Proxmox VE, OpenStack, Hyper-V, AWS-style and Azure-style personalities are represented as standby roadmap providers.

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

The drawer can be opened from the top command bar, sidebar or scenario panel.

### Health and engine status

The sidebar health ring derives a simple health score from current error rate and shows uptime and version. The footer shows the active persona, public SDK endpoint, simulator administrator identity and local system time.

## Authentication model

The dashboard and read-only health/bootstrap/inventory/telemetry/vmdks endpoints are public inside the disposable lab. Mutating controls require the admin bearer token printed by `chimera serve`.

The browser stores the token only in `sessionStorage`, so it disappears when the browser session closes.

## Backend endpoints used by the UX

```text
GET  /__chimera/health
GET  /__chimera/api/bootstrap
GET  /__chimera/api/inventory
GET  /__chimera/api/vmdks
GET  /__chimera/api/telemetry
GET  /__chimera/state
POST /__chimera/faults
POST /__chimera/reset
POST /__chimera/scenario/{clean|slow|flaky|resume}
```

## Responsive behavior

- full three-column command center on large desktops
- topology/activity + full-width request feed on medium screens
- single-column workbench on tablets
- collapsed icon-only sidebar and simplified controls on small screens

## Future UX additions

- actual persona activation/hot switching
- drag-and-drop synthetic topology builder
- request trace and SOAP/REST inspector
- lease/export progress visualization
- inventory CRUD
- saved scenario composer
- provider-specific templates
- test run history and pass/fail reports
- OpenTelemetry trace exploration
