# Chimera

[![CI](https://github.com/zyvorai/chimera/actions/workflows/ci.yml/badge.svg)](https://github.com/zyvorai/chimera/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/zyvorai/chimera.svg)](https://pkg.go.dev/github.com/zyvorai/chimera)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

**Chimera** is a programmable infrastructure simulation engine for integration-testing migration, discovery, export and automation software without provisioning the real infrastructure platform.

The architecture is provider-persona based: **vSphere is the first fully implemented persona**, while the control plane and UX are designed to grow into Nutanix Prism, Proxmox VE, OpenStack, Hyper-V and cloud/API personalities.

> **One engine. Many infrastructure personalities.**

Today, the vSphere persona is deliberately much deeper than a simple HTTP mock. A real govmomi client can authenticate with username/password, establish a session, traverse inventory, resolve VMs, create OVF descriptors, call `ExportVm`, wait on an `HttpNfcLease`, download a VMDK fixture, retry a broken transfer, and resume with HTTP Range/206 semantics.

Chimera is a test and compatibility appliance. It is not VMware, Nutanix, Microsoft, Red Hat, Proxmox or cloud-vendor software, and it is not intended to host production workloads.

## What is included

### Chimera Command Center UX

Open `http://localhost:8989/__chimera/` after starting the engine.

The embedded, zero-dependency dashboard is a full infrastructure command center:

- premium dark control-plane shell with provider-persona navigation
- six live KPIs: requests, error rate, active sessions, transferred bytes, exports and response latency
- simulator health ring, uptime and build/version status
- visual datacenter → cluster → host → datastore topology
- traffic-activity donut driven by real gateway request classes
- live request feed with method, path, HTTP status and duration
- searchable/paginated VM inventory with power-state filters, export selection, and a per-VM Fixture badge (real file vs. synthetic)
- VMDK Library panel showing every fixture file and which VM it's assigned to, live-updated as files are added/removed
- deterministic scenario launcher for clean, slow, flaky and resume paths
- slide-over Fault Studio for latency, status codes, API/NFC failures, stream drops and bandwidth caps
- session-only admin-token unlock flow
- responsive desktop/tablet/mobile UI
- no Node/npm runtime and no external CDN dependencies

See [`docs/UX.md`](docs/UX.md).

### vSphere persona — implemented now

- vSphere SOAP/VIM endpoint at `/sdk`
- strict username/password authentication and session cookies
- service content + property collector
- datacenters, clusters, hosts, resource pools, datastores, networks and VMs
- VM discovery through govmomi
- `OvfManager.CreateDescriptor`
- `VirtualMachine.ExportVm`
- `HttpNfcLease` ready/completion/abort/progress behavior
- generated or operator-supplied VMDK fixture
- NFC download URLs
- HTTP `Range` / 206 resume behavior
- deterministic connection drops, 4xx/5xx faults, latency and bandwidth limits
- self-signed HTTPS option
- admin health/state/scenario APIs

### Persona roadmap

| Persona | Status | Target compatibility |
|---|---|---|
| VMware vSphere | **Implemented** | SOAP / VIM / OVF / HTTP NFC |
| Nutanix Prism | Planned | Prism Central / Element APIs |
| Proxmox VE | Planned | PVE REST API / tasks / storage |
| OpenStack | Planned | Keystone / Nova / Glance / Cinder |
| Microsoft Hyper-V | Planned | WinRM / WMI / virtualization management |
| Cloud APIs | Planned | AWS/Azure-style discovery and export test surfaces |

The future-provider contract is documented in [`docs/PROVIDER_ARCHITECTURE.md`](docs/PROVIDER_ARCHITECTURE.md).

## Why this matches Transiva

Transiva's vSphere provider uses govmomi for normal login and inventory. Its export flow creates an OVF descriptor, calls `vm.Export()`, waits for the lease, and downloads each lease item. When a partial local file exists, Transiva sends `Range: bytes=N-`. Chimera implements those exact boundaries so Transiva can exercise its production code path without requiring a physical vCenter.

## Install

**Prebuilt packages** (Linux amd64/arm64): grab the `.deb` or `.rpm` from the [latest release](https://github.com/zyvorai/chimera/releases/latest), then:

```bash
sudo apt install ./chimera_*.deb   # Debian/Ubuntu
sudo dnf install ./chimera-*.rpm   # Fedora/RHEL/Alma
sudo systemctl enable --now chimera
```

This installs the binary to `/usr/bin/chimera`, a systemd unit (`systemd/chimera.service`), and an env file at `/etc/chimera/chimera.env`.

**Remote host, from source**: `./scripts/deploy-remote.sh <host> [user]` cross-compiles and installs Chimera as a systemd service over SSH — no Go toolchain needed on the target. See `--help` for options.

**Build packages yourself**: `make package` (needs [nfpm](https://nfpm.goreleaser.com)) builds `.deb`/`.rpm` into `dist/` from `packaging/nfpm.yaml`.

## Quick start

Requirements: Go 1.25+.

```bash
go mod tidy
go build -o bin/chimera ./cmd/chimera
./bin/chimera serve -config config.example.json
```

Expected output:

```text
Chimera ready
  endpoint: http://localhost:8989/sdk
  username: administrator@vsphere.local
  password: vmware
  admin:    http://localhost:8989/__chimera/
  token:    chimera-admin
  sample VM path: /DC0/vm/DC0_C0_RP0_VM0
```

Open the Command Center:

```text
http://localhost:8989/__chimera/
```

The dashboard itself is public in the disposable lab. Mutating controls are protected with the admin token printed by `chimera serve`.

Run the end-to-end client probe:

```bash
./bin/chimera selftest \
  -url http://localhost:8989/sdk \
  -user administrator@vsphere.local \
  -pass vmware
```

The self-test performs login → datacenter discovery → VM inventory → `ExportVm` → lease wait → NFC read → lease complete. By default it only reads the first 4KB; pass `-vm <name>` to target a specific VM and `-save <path>` to download the complete disk instead (see `scripts/verify-real-fixture.sh` for a full worked example against a real fixture).

## Docker

```bash
docker compose up --build
```

If the client is in another container or host, set `CHIMERA_PUBLIC_HOST` to an address reachable **from the client**. Export lease URLs embed this value.

## Use with Transiva

```bash
./scripts/make-transiva-config.sh > /tmp/transiva-chimera.yaml

cd ../transiva
./scripts/build.sh
./bin/hyperexport --config /tmp/transiva-chimera.yaml
```

Choose `DC0_C0_RP0_VM0` or another VM shown by the Command Center. See [`docs/TRANSIVA.md`](docs/TRANSIVA.md) for detailed retry/resume recipes.

## Export fixture modes

### Generated transport fixture

With `fixture_vmdk` empty, Chimera creates a deterministic byte stream of `fixture_size_mb` and exposes it as `disk-0.vmdk`. This is ideal for testing authentication, inventory, leases, parallel transfers, retries, checkpoints and HTTP Range behavior.

The generated bytes are intentionally **not a valid VMDK**.

### Real VMDK fixture

To exercise downstream conversion paths such as qemu-img/hyper2kvm, supply a non-sensitive test VMDK:

```bash
CHIMERA_FIXTURE_VMDK=/lab/fixtures/ubuntu-test.vmdk \
./bin/chimera serve -listen 0.0.0.0:8989
```

Chimera never modifies the supplied fixture.

### Directory of VMDKs, one per VM

Point at a directory instead of a single file to give each simulated VM its own disk:

```bash
CHIMERA_FIXTURE_VMDK_DIR=/lab/fixtures \
./bin/chimera serve -listen 0.0.0.0:8989
```

Matching is two-pass: a file named after a VM (e.g. `DC0_C0_RP0_VM0.vmdk`) is assigned to that VM; any leftover files and VMs are then paired off in sorted order. VMs that still get nothing keep the default generated fixture. Mutually exclusive with `fixture_vmdk`. The directory is re-scanned automatically every 5 seconds — drop in (or remove) a file and it's picked up without restarting the server. See the VMDK Library panel in the Command Center, or `GET /__chimera/api/vmdks`, to see the resulting assignment.

Don't have a real VMDK handy? `make fixtures` fetches a small, official, checksummed Alpine Linux cloud image and converts it to VMDK automatically (see `scripts/fetch-sample-fixtures.sh`):

```bash
make fixtures
CHIMERA_FIXTURE_VMDK_DIR=./fixtures ./bin/chimera serve -listen 0.0.0.0:8989
```

`scripts/verify-real-fixture.sh` proves the whole path end-to-end: it exports the assigned VM's disk through the real NFC download (not just `selftest`'s 4KB probe) and confirms `qemu-img info` recognizes the result as a valid disk image.

## Configuration

| JSON | Environment | Default |
|---|---|---:|
| `listen` | `CHIMERA_LISTEN` | `127.0.0.1:8989` |
| `public_host` | `CHIMERA_PUBLIC_HOST` | listener address |
| `tls` | `CHIMERA_TLS` | `false` |
| `username` | `CHIMERA_USERNAME` | `administrator@vsphere.local` |
| `password` | `CHIMERA_PASSWORD` | `vmware` |
| `datacenters` | `CHIMERA_DATACENTERS` | `1` |
| `clusters` | `CHIMERA_CLUSTERS` | `1` |
| `hosts_per_dc` | `CHIMERA_HOSTS_PER_DC` | `0` |
| `hosts_per_cluster` | `CHIMERA_HOSTS_PER_CLUSTER` | `1` |
| `datastores` | `CHIMERA_DATASTORES` | `1` |
| `vms_per_pool` | `CHIMERA_VMS_PER_POOL` | `3` |
| `soap_delay_ms` | `CHIMERA_SOAP_DELAY_MS` | `0` |
| `admin_token` | `CHIMERA_ADMIN_TOKEN` | `chimera-admin` |
| `fixture_vmdk` | `CHIMERA_FIXTURE_VMDK` | generated fixture |
| `fixture_vmdk_dir` | `CHIMERA_FIXTURE_VMDK_DIR` | generated fixture |
| `fixture_size_mb` | `CHIMERA_FIXTURE_SIZE_MB` | `16` |

## Admin APIs

Public read endpoints:

```text
GET /__chimera/health
GET /__chimera/api/bootstrap
GET /__chimera/api/inventory
GET /__chimera/api/vmdks
GET /__chimera/api/telemetry
```

`GET /__chimera/api/telemetry` returns gateway-derived live metrics and a bounded recent-request feed. Dashboard polling under `/__chimera` is intentionally excluded from those counters so the UI does not create its own traffic metrics.

Protected endpoints require:

```text
Authorization: Bearer chimera-admin
```

Available protected operations:

```text
GET  /__chimera/state
POST /__chimera/faults
POST /__chimera/reset
POST /__chimera/scenario/clean
POST /__chimera/scenario/slow
POST /__chimera/scenario/flaky
POST /__chimera/scenario/resume
```

Example fault policy:

```bash
curl -X POST \
  -H 'Authorization: Bearer chimera-admin' \
  -H 'Content-Type: application/json' \
  http://localhost:8989/__chimera/faults \
  -d '{
    "latency_ms":250,
    "fail_next":1,
    "fail_status":503,
    "nfc_fail_next":1,
    "nfc_drop_next":1,
    "nfc_drop_after_bytes":2097152,
    "bandwidth_bytes_per_sec":1048576
  }'
```

Built-in scenarios:

```bash
./scripts/scenario.sh clean
./scripts/scenario.sh slow
./scripts/scenario.sh flaky
./scripts/scenario.sh resume
```

The `resume` scenario aborts the next export stream after 2 MiB. A compatible client can retry with `Range: bytes=N-` and continue from the partial offset.

## Architecture

```text
                  +------------------------------------+
                  |       Chimera Command Center       |
                  | personas · inventory · faults · UI |
                  +------------------+-----------------+
                                     |
                              /__chimera APIs
                                     |
                                     v
Client / migration tool ---> +---------------------------+
                             | Chimera public gateway    |
                             | auth path · URL rewrite   |
                             | Range/206 · faults · bw   |
                             +-------------+-------------+
                                           |
                         +-----------------+-----------------+
                         |                                   |
                         v                                   v
              +---------------------+             +--------------------+
              | vSphere persona     |             | future personas    |
              | govmomi simulator   |             | Nutanix / PVE /    |
              | SOAP/VIM inventory  |             | OpenStack / Hyper-V|
              +----------+----------+             +--------------------+
                         |
                         v
              +---------------------+
              | export compatibility|
              | OVF · ExportVm · NFC|
              +----------+----------+
                         |
                         v
                    fixture store
```

## Project layout

```text
cmd/chimera/             CLI: serve, selftest, print-config
internal/config/         config + environment overrides
internal/lab/            simulator lifecycle + TLS
internal/exportshim/     current vSphere OVF/ExportVm/NFC compatibility
internal/fixture/        generated/real VMDK fixture registry
internal/gateway/        public proxy, APIs, Range, faults and embedded UX
internal/faults/         deterministic scenario state
internal/selftest/       govmomi end-to-end probe
integration/             compatibility tests
docs/                    UX, architecture, Transiva guide, test matrix
scripts/                 config, scenario, smoke, deploy and packaging helpers
systemd/                 systemd unit (used by packages and deploy-remote.sh)
packaging/               nfpm config for .deb/.rpm builds
.github/workflows/       CI + tagged-release packaging
```

## Testing

```bash
make test
make build
```

Then:

```bash
make run
./bin/chimera selftest \
  -url http://localhost:8989/sdk \
  -user administrator@vsphere.local \
  -pass vmware
```

See [`docs/TEST_MATRIX.md`](docs/TEST_MATRIX.md) for the acceptance matrix.

## Contributing

Issues and PRs are welcome. `make verify` runs the same checks as CI (build, vet, tests, `gofmt`, and a syntax check on the embedded dashboard JS) — run it before opening a PR.

## License

Apache License 2.0 — see [`LICENSE`](LICENSE). Third-party dependencies are listed in [`THIRD_PARTY.md`](THIRD_PARTY.md).
