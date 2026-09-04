# Chimera

[![CI](https://github.com/zyvorai/chimera/actions/workflows/ci.yml/badge.svg)](https://github.com/zyvorai/chimera/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/zyvorai/chimera.svg)](https://pkg.go.dev/github.com/zyvorai/chimera)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

**Chimera** is a programmable infrastructure simulation engine for integration-testing migration, discovery, export and automation software without provisioning the real infrastructure platform.

The architecture is provider-persona based: **vSphere is the deepest persona** (Command Center + govmomi), with **Nutanix Prism v3** and **Hyper-V WS-Man** available as protocol surfaces. Proxmox VE, OpenStack, and cloud/API personalities remain on the roadmap.

> **One engine. Many infrastructure personalities.**

Today, the vSphere persona is deliberately much deeper than a simple HTTP mock. A real govmomi client can authenticate with username/password, establish a session, traverse inventory, resolve VMs, create OVF descriptors, call `ExportVm`, wait on an `HttpNfcLease`, download a VMDK fixture, retry a broken transfer, and resume with HTTP Range/206 semantics.

Chimera is a test and compatibility appliance. It is not VMware, Nutanix, Microsoft, Red Hat, Proxmox or cloud-vendor software, and it is not intended to host production workloads.

## What is included

### Chimera Command Center UX

Open `http://localhost:8989/__chimera/` after starting the engine.

The embedded, zero-dependency dashboard is a full infrastructure command center:

- premium dark control-plane shell with provider-persona navigation, a foldable sidebar, and full-width stacked panels for readability
- six live KPIs: requests, error rate, active sessions, transferred bytes, exports and response latency
- simulator health ring, uptime and build/version status
- visual datacenter → cluster → host → datastore topology, with a full-screen zoomable view
- traffic-activity donut driven by real gateway request classes
- live request feed with method, path, HTTP status and duration
- searchable/paginated VM inventory with power-state filters, export selection, and a per-VM Fixture badge (real file vs. synthetic)
- VMDK Library: upload a `.vmdk` straight from the browser, or browse and pin a file already staged on the host — with an optional manual VM assignment that overrides the automatic name/round-robin matching. `fixture_vmdk_dir` scanning recurses into subdirectories and supports additional read-only watch directories
- deterministic scenario launcher for clean, slow, flaky and resume paths
- slide-over Fault Studio for latency, status codes, API/NFC failures, stream drops and bandwidth caps
- real username/password login (default `admin`/`admin`, changeable from the dashboard's Settings panel) gating every mutating control
- a Settings panel showing the listen address and letting an operator change the admin login live, no restart needed
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
- self-signed HTTPS option (`CHIMERA_TLS=true`)
- admin health/state/scenario APIs, gated by a real login (default `admin`/`admin`, changeable)

### Nutanix + Hyper-V personas — available

Select with `"persona"` in config or `CHIMERA_PERSONA`. These serve protocol endpoints only (no Command Center yet):

| Persona | Endpoint | Coverage |
|---|---|---|
| Nutanix Prism | `/api/nutanix/v3` | Basic auth, cluster identity, VM list/detail, power-state tasks, deterministic disk export |
| Microsoft Hyper-V | `/wsman` | WS-Man Identify, Enumerate/Pull of `Msvm_ComputerSystem`, `RequestStateChange` |

```bash
CHIMERA_PERSONA=nutanix CHIMERA_USERNAME=admin CHIMERA_PASSWORD=secret go run ./cmd/chimera serve
CHIMERA_PERSONA=hyperv CHIMERA_USERNAME=Administrator CHIMERA_PASSWORD=secret go run ./cmd/chimera serve
```

### Persona roadmap

| Persona | Status | Target compatibility |
|---|---|---|
| VMware vSphere | **Implemented** | SOAP / VIM / OVF / HTTP NFC |
| Nutanix Prism | **Available** | Prism v3 auth / inventory / power / disk export |
| Microsoft Hyper-V | **Available** | WS-Man Identify / Enumerate / Pull / power |
| Proxmox VE | Planned | PVE REST API / tasks / storage |
| OpenStack | Planned | Keystone / Nova / Glance / Cinder |
| Cloud APIs | Planned | AWS/Azure-style discovery and export test surfaces |

The provider contract and layout are documented in [`docs/PROVIDER_ARCHITECTURE.md`](docs/PROVIDER_ARCHITECTURE.md).

## Why this matches Transiva

Transiva's vSphere provider uses govmomi for normal login and inventory. Its export flow creates an OVF descriptor, calls `vm.Export()`, waits for the lease, and downloads each lease item. When a partial local file exists, Transiva sends `Range: bytes=N-`. Chimera implements those exact boundaries so Transiva can exercise its production code path without requiring a physical vCenter.

## Install

**Prebuilt packages** (Linux amd64/arm64): grab the `.deb` or `.rpm` from the [latest release](https://github.com/zyvorai/chimera/releases/latest), then:

```bash
sudo apt install ./chimera_*.deb   # Debian/Ubuntu
sudo dnf install ./chimera-*.rpm   # Fedora/RHEL/Alma
sudo systemctl enable --now chimera
```

This installs the binary to `/usr/bin/chimera`, a systemd unit (`systemd/chimera.service`), and an env file at `/etc/chimera/chimera.env`. Logs go to the journal (`journalctl -u chimera` or `-t chimera`).

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
  login:    admin / admin
  token:    chimera-admin
  sample VM path: /DC0/vm/DC0_C0_RP0_VM0
  ⚠ Using default admin credentials (admin/admin), reachable on all interfaces by default —
    set CHIMERA_ADMIN_USERNAME/CHIMERA_ADMIN_PASSWORD (or change it in the dashboard's Settings) to change them.
```

Open the Command Center — `/` also redirects here:

```text
http://localhost:8989/__chimera/
```

The browser dashboard itself is a full-page login gate — nothing renders until you log in with `admin`/`admin` (printed at startup, changeable from the dashboard's Settings panel or `CHIMERA_ADMIN_USERNAME`/`CHIMERA_ADMIN_PASSWORD`). That's a frontend-only gate, though: the underlying read APIs (health, bootstrap, inventory, telemetry, VMDK list) stay public in the disposable lab regardless, so scripts/CI can still poll them directly without a token. Chimera listens on `0.0.0.0` by default, so change the admin login before exposing an instance beyond your own machine.

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

Matching is a three-pass process: any file explicitly pinned to a VM (see "Manual assignment" below) wins first; then a file named after a VM (e.g. `DC0_C0_RP0_VM0.vmdk`, matched on its own basename even when nested in a subdirectory) is assigned to that VM; any leftover files and VMs are then paired off in sorted order. VMs that still get nothing keep the default generated fixture. Mutually exclusive with `fixture_vmdk`. The directory is scanned **recursively** and re-scanned automatically every 5 seconds — drop in (or remove) a file anywhere under it, at any depth, and it's picked up without restarting the server. See the VMDK Library panel in the Command Center, or `GET /__chimera/api/vmdks`, to see the resulting assignment.

Additional read-only directories can be scanned the same way via `fixture_vmdk_dirs` / `CHIMERA_FIXTURE_VMDK_DIRS` (comma-separated) — useful when fixtures live in more than one place on the host. Uploads through the dashboard always land in the primary `fixture_vmdk_dir`; the extra directories only ever contribute files an operator staged there directly.

**Manual assignment** — from the dashboard's VMDK Library card, either upload a `.vmdk` directly from the browser, or browse and pick a file already staged on the host, and optionally assign it to a specific VM. That assignment overrides the automatic name/round-robin matching (reported as `manual` in `GET /__chimera/api/vmdks`) until cleared. The host browser is hard-scoped to the configured fixture directories — it can never reveal an arbitrary path on the host.

Don't have a real VMDK handy? `make fixtures` fetches a small, official, checksummed Alpine Linux cloud image and converts it to VMDK automatically (see `scripts/fetch-sample-fixtures.sh`):

```bash
make fixtures
CHIMERA_FIXTURE_VMDK_DIR=./fixtures ./bin/chimera serve -listen 0.0.0.0:8989
```

`scripts/verify-real-fixture.sh` proves the whole path end-to-end: it exports the assigned VM's disk through the real NFC download (not just `selftest`'s 4KB probe) and confirms `qemu-img info` recognizes the result as a valid disk image.

## Configuration

| JSON | Environment | Default |
|---|---|---:|
| `persona` | `CHIMERA_PERSONA` | `vsphere` (`nutanix`, `hyperv`) |
| `listen` | `CHIMERA_LISTEN` | `0.0.0.0:8989` |
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
| `admin_username` | `CHIMERA_ADMIN_USERNAME` | `admin` |
| `admin_password` | `CHIMERA_ADMIN_PASSWORD` | `admin` |
| `fixture_vmdk` | `CHIMERA_FIXTURE_VMDK` | generated fixture |
| `fixture_vmdk_dir` | `CHIMERA_FIXTURE_VMDK_DIR` | generated fixture |
| `fixture_vmdk_dirs` | `CHIMERA_FIXTURE_VMDK_DIRS` (comma-separated) | none |
| `fixture_size_mb` | `CHIMERA_FIXTURE_SIZE_MB` | `16` |

`username`/`password` are the simulated vCenter login used by API clients (govc, Transiva, etc.) — unrelated to `admin_username`/`admin_password`, which gate the Chimera dashboard's own login.

## Admin APIs

Public read endpoints (plus the login exchange below):

```text
GET  /__chimera/health
GET  /__chimera/api/bootstrap
GET  /__chimera/api/inventory
GET  /__chimera/api/vmdks
GET  /__chimera/api/telemetry
POST /__chimera/login
```

`GET /__chimera/api/telemetry` returns gateway-derived live metrics and a bounded recent-request feed. Dashboard polling under `/__chimera` is intentionally excluded from those counters so the UI does not create its own traffic metrics.

`POST /__chimera/login` exchanges an `{"username","password"}` body (default `admin`/`admin`) for the admin bearer token — this is what the dashboard's login form calls; scripts can call it directly too, or just use the token printed at startup.

Every other endpoint is protected and requires:

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
POST /__chimera/admin/credentials       # change the dashboard login: {"username","password"}
POST /__chimera/api/vmdks/upload        # multipart "file" (+optional "vm_name") into fixture_vmdk_dir
GET  /__chimera/api/vmdks/browse        # ?root=<index>&path=<relative> — list files under a configured fixture root
POST /__chimera/api/vmdks/assign        # {"root","file_name","vm_name"} — pin a file already on disk to a VM
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
              | vSphere persona     |             | HTTP personas      |
              | govmomi simulator   |             | Nutanix Prism v3   |
              | SOAP/VIM inventory  |             | Hyper-V WS-Man     |
              +----------+----------+             +--------------------+
                         |                                   ^
                         v                                   |
              +---------------------+             +--------------------+
              | export compatibility|             | planned: PVE /     |
              | OVF · ExportVm · NFC|             | OpenStack / cloud  |
              +----------+----------+             +--------------------+
                         |
                         v
                    fixture store
```

## Project layout

```text
cmd/chimera/             CLI: serve, selftest, print-config
internal/config/         config + environment overrides (incl. persona)
internal/lab/            simulator lifecycle + TLS + HTTP persona start
internal/exportshim/     current vSphere OVF/ExportVm/NFC compatibility
internal/fixture/        generated/real VMDK fixture registry
internal/gateway/        public proxy, APIs, Range, faults and embedded UX
internal/faults/         deterministic scenario state
internal/personas/       nutanix (Prism v3) + hyperv (WS-Man) + shared store
internal/selftest/       govmomi end-to-end probe
integration/             vSphere + Nutanix/Hyper-V compatibility tests
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
