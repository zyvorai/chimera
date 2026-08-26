# Chimera Provider Persona Architecture

Chimera should evolve around provider **personas**, not around one vCenter simulator. A persona is a protocol-compatible control-plane surface with its own authentication, inventory model, operations, export semantics and fault extensions.

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
    nutanix/
    proxmox/
    openstack/
    hyperv/
    cloud/
```

The current `internal/exportshim` and govmomi simulator logic are the seed of `personas/vsphere` and can be moved there when the second persona is implemented.

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
