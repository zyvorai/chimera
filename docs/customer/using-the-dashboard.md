# Using the Dashboard

Chimera's **Command Center** is a single-page infrastructure console embedded in the Go binary (`internal/gateway/ui.go`). There is no Node/React build step — open `/__chimera/` on the same listener as the vSphere SDK.

## Shell layout

| Region | Role |
|--------|------|
| Left rail | Control Center nav, provider personas (vSphere active in UI; Nutanix/Hyper-V via `CHIMERA_PERSONA`; others roadmap), System Health ring |
| Top bar | Global search (⌘K), Scenario shortcut, Fault Studio (bolt), notifications bell → Live requests, Settings (gear), user menu (logout) |
| Main column | Overview KPIs, topology, Top Activity, Live requests, VM inventory, Scenario Launcher, VMDK Library |
| Right drawers | Fault Studio (fault injection) and Settings (listen address + admin credentials) |

## How to move around

1. Sign in — mutating controls (scenarios, faults, VMDK upload, credential save) require a dashboard session.
2. Use sidebar anchors or scroll — Overview is one long workbench page with hash targets (`#topology`, `#inventory`, `#vmdks`, `#requests`, `#activity`, `#scenarios`).
3. Click **↻ Refresh** after scenario or fault changes if you want an immediate telemetry pull (otherwise ~2s polling).
4. Open **View Full Topology** for the zoomable estate map (**Fit** / **＋**, Escape to exit).
5. Filter VMs by name or power state; paginate 5 per page; global search focuses inventory.
6. Open **Fault Studio** from the bolt icon (not the gear — that is Settings).

## Operate from the console (UX)

1. Land on Overview after login; read six KPI tiles and sidebar health ring.
2. Drive traffic with `selftest` or Transiva against `http(s)://<host>:8989/sdk` — dashboard polls do not inflate metrics.
3. Use **Scenario Launcher** for presets; open **Fault Studio** for manual latency/failure/NFC drop tuning.
4. Stage real disks via **VMDK Library** when testing conversion, not just synthetic export.
5. Bell icon jumps to **Live requests** when errors spike; toggle **View All** for 30-row buffer.
6. Rotate admin password under **Settings** before exposing `<host>` beyond a lab VLAN.

**Empty / fail:** See table below; page guides under [pages/](pages/README.md) have surface-specific checks.

| Symptom | Check |
|---------|--------|
| Login loop | Defaults `admin`/`admin`; or values set via Settings / `CHIMERA_ADMIN_*` |
| Empty inventory | Simulator counts in config; restart after config changes |
| No KPI movement | Clients must hit `/sdk`, not `/__chimera/` |
| Self-test fails | URL must be the SDK (`/sdk`), not Command Center |
| Fault Studio vs Settings | Bolt = faults; gear = settings |

**Success:** KPIs, activity donut, and live request feed update during exports; scenarios toast confirmation; VM/fixture badges reflect VMDK matching.

Next: [Page-by-page guides](pages/README.md) · [Workflows](workflows.md)
