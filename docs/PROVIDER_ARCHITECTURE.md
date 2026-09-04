# Chimera Provider Persona Architecture

Chimera evolves around provider **personas**, not around one vCenter simulator. A persona is a protocol-compatible control-plane surface with its own authentication, inventory model, operations, export semantics and fault extensions.

## Status today

| Persona | Package / entry | Endpoints | Command Center |
|---|---|---|---|
| vSphere | `internal/lab` + `exportshim` + govmomi simulator | `/sdk`, NFC, `/__chimera/` | Yes |
| Nutanix Prism | `internal/personas/nutanix` via `lab.StartHTTPPersona` | `/api/nutanix/v3` | No (protocol only) |
| Hyper-V | `internal/personas/hyperv` via `lab.StartHTTPPersona` | `/wsman` | No (protocol only) |
| Proxmox / OpenStack / Cloud | — | — | Planned |

Select with `persona` in JSON config or `CHIMERA_PERSONA` (`vsphere` default). Shared VM/task/disk seed lives in `internal/personas/common`.

### Nutanix coverage (Prism v3-compatible)

- Basic auth
- Cluster identity
- VM inventory list/detail
- Power-state task creation + task lookup
- Deterministic virtual-disk export bytes

### Hyper-V coverage (WS-Man)

- WS-Man Identify
- Enumerate / Pull of `Msvm_ComputerSystem`
- `RequestStateChange`
- Deterministic VM inventory

Next iteration (not yet): Nutanix v4, categories/projects/images, Hyper-V WMI association traversal, snapshots/checkpoints, VHDX byte-range export, SCVMM persona.

## Proposed contract

```go
type Persona interface {
    ID() string
    DisplayName() string
    Capabilities() Capabilities
    Mount(router Router) error
    Inventory(context.Context) (Inventory, error)
    Reset(context.Context) error
    Close(context.Context) error
}
```

Suggested capabilities include discovery, VM export, snapshots, change tracking, storage, networks, tasks, events, images, templates and migration-specific features.

## Directory evolution

```text
internal/
  core/
    engine.go
    persona.go
    inventory.go
    scenario.go
  gateway/
  faults/
  fixture/
  personas/
    vsphere/
      simulator.go
      auth.go
      inventory.go
      export.go
      nfc.go
    common/          # shared Store (landed)
    nutanix/         # landed
    hyperv/          # landed
    proxmox/
    openstack/
    cloud/
```

Nutanix and Hyper-V already land under this layout. They intentionally skip the Command Center gateway for now.

The current `internal/exportshim` and govmomi simulator logic remain the seed of `personas/vsphere` and can be moved there when the full `internal/core` Persona interface is introduced.

## Persona responsibilities

Each persona should own:

- protocol endpoints
- authentication behavior
- object naming / IDs
- inventory hierarchy
- task and async-operation semantics
- export/import behavior
- vendor-specific error contracts
- version personality
- provider-specific scenarios

## Shared Chimera responsibilities

The core engine should own:

- HTTP/TLS lifecycle
- admin API and Command Center
- global scenario/fault engine
- fixture storage
- request telemetry
- deterministic clocks/randomness
- persona registry
- scenario persistence
- observability
- test harness APIs

## Goal

A migration product should be able to point its real provider adapter at Chimera and believe it is talking to the target platform closely enough to exercise production discovery, planning, export, retry, resume and error-handling code paths.
