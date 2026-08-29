# Using the Dashboard

Chimera's **Command Center** is a single-page infrastructure console embedded in the Go binary (`internal/gateway/ui.go`). There is no Node/React build step.

## Shell layout

| Region | Role |
|--------|------|
| Left rail | Provider personas (vSphere active; others roadmap) + health ring |
| Top bar | Refresh, notifications (jumps to Live requests), Fault Studio gear |
| Main column | Overview KPIs, topology, activity donut, live requests, VM inventory, VMDK Library, scenarios |
| Right drawer | Fault Studio (latency, failures, NFC drops, bandwidth) |

## How to move around

1. Sign in — mutating controls require a dashboard session.
2. Use **Refresh** after scenario or fault changes.
3. Open **View Full Topology** for the zoomable estate map.
4. Filter VMs by name or power state; paginate as needed.
5. Open Fault Studio from the gear icon when you need deterministic failures for Transiva retries.

## Empty / fail tips

| Symptom | Check |
|---------|--------|
| Login loop | Defaults `admin`/`admin`; or values set via Settings / `CHIMERA_ADMIN_*` |
| Empty inventory | Simulator counts in config; restart after config changes |
| No KPI movement | Dashboard polling is excluded from counters — drive traffic via `/sdk` or Transiva |
| Self-test fails | URL must be the SDK (`/sdk`), not `/__chimera/` |

Next: [Page-by-page guides](pages/README.md) · [Workflows](workflows.md)
