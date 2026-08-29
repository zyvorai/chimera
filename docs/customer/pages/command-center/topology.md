# Topology

## Purpose

Visual datacenter → cluster → host → datastore map of the simulated estate.

## When to use it

- Orient operators to inventory shape before export or Transiva wiring
- Confirm config counts (`datacenters`, `clusters`, `hosts_per_cluster`, `datastores`) match expectations
- Walk new teammates through the fake vSphere layout without opening the SDK

## How to get there

- Screen id: `#topology`
- Nav: **Control Center → Topologies** (sidebar ⌘) or scroll to **Infrastructure Topology** on Overview

## Operate from the console (UX)

1. From Overview, locate the **Infrastructure Topology** card — subtitle shows live counts (for example `1 datacenter · 2 clusters · 3 hosts`).
2. Read node labels: **DC0** (datacenter), **Cluster0/1**, **Host0–2** with lab IPs, **Datastore0–2** with fixture size from config.
3. Click **View Full Topology** to expand the same diagram full-viewport; use toolbar **Fit** to reset zoom or **＋** to step zoom in.
4. Dismiss fullscreen via **Exit Full Screen**, click outside the diagram, or press **Escape**.
5. Cross-check host/datastore count with [VM inventory](vm-inventory.md) sidebar badge and table footer.

**Empty / fail:** Subtitle stuck at generic text → bootstrap API may have failed; reload after login or check engine logs. Nodes show but datastore labels say `fixture` with wrong size → verify `fixture_size_mb` in config and restart if you changed estate shape.

**Success:** Topology subtitle matches your JSON config; fullscreen zoom works; footer **vSphere Persona** strip shows the public SDK endpoint clients should use.

## Related pages

- [VM inventory](vm-inventory.md)
- [Overview KPIs](overview.md)
- [Admin basics](../../admin-basics.md)
- [Getting Started](../../getting-started.md)
- [Page index](../../PAGE_INDEX.md)
