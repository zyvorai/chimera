# Getting Started with Chimera

## What you need

- Go 1.25+ to build from source, **or** a `.deb`/`.rpm` from [GitHub Releases](https://github.com/zyvorai/chimera/releases/latest)
- Browser for Command Center at `/__chimera/`
- Optional: Transiva checkout for end-to-end export smoke

## 1. Start the engine

```bash
git clone https://github.com/zyvorai/chimera.git
cd chimera
go mod tidy
go build -o bin/chimera ./cmd/chimera
./bin/chimera serve -config config.example.json
```

Or remote install:

```bash
./scripts/deploy-remote.sh <host> [user]
```

## 2. Open Command Center

```text
http://<host>:8989/__chimera/
```

Log in with **admin** / **admin** (change later in Settings). Without a session the UI shows only the login screen — no inventory flash.

## 3. Confirm the vSphere persona

```bash
./bin/chimera selftest \
  -url http://<host>:8989/sdk \
  -user administrator@vsphere.local \
  -pass vmware
```

A passing self-test proves auth, inventory, OVF descriptor, `ExportVm`, and NFC lease download.

### Optional: Nutanix or Hyper-V protocol personas

Command Center stays on the vSphere listener. For Prism or WS-Man smoke tests, start a second process (or replace the persona) with env overrides:

```bash
CHIMERA_PERSONA=nutanix CHIMERA_USERNAME=admin CHIMERA_PASSWORD=secret \
  ./bin/chimera serve -listen 0.0.0.0:8990
# → http://<host>:8990/api/nutanix/v3

CHIMERA_PERSONA=hyperv CHIMERA_USERNAME=Administrator CHIMERA_PASSWORD=secret \
  ./bin/chimera serve -listen 0.0.0.0:8991
# → http://<host>:8991/wsman
```

Coverage today: Nutanix basic auth, cluster identity, VM list/detail, power tasks, disk export bytes; Hyper-V Identify, Enumerate/Pull, `RequestStateChange`. See product-repo `README.md` and `docs/PROVIDER_ARCHITECTURE.md`.

## 4. First useful checks in the UI

1. Overview KPIs populate after self-test (requests, sessions, exports).
2. Topology shows datacenter → cluster → host → datastore.
3. VM inventory lists simulator VMs (for example `DC0_C0_RP0_VM0`).
4. Scenario Launcher: try **Clean Environment**, then **Slow Fabric** and watch latency in Live requests.

## Operate from the console (UX)

1. Browse to `http(s)://<host>:8989/__chimera/` and sign in (**admin** / **admin** unless rotated).
2. Confirm sidebar **System Health** ring and footer **vSphere Persona** strip show uptime and SDK endpoint.
3. Click **↻ Refresh** on Overview if KPIs are flat, then run `selftest` from the host.
4. Open **Inventory** — verify VM count badge matches config; copy a VM name with the row **⇩** action.
5. Optional: **Settings** (gear) → change admin password before sharing the lab URL.

**Empty / fail:** Login loop → defaults or `CHIMERA_ADMIN_*`; KPIs zero after selftest → client must use `/sdk` URL. Inventory empty → restart after config estate changes.

**Success:** KPIs and Live requests move during selftest; topology subtitle shows your datacenter/cluster counts.

## Next

- [Using the Dashboard](using-the-dashboard.md)
- [Wire Transiva](workflows.md)
- [Admin basics](admin-basics.md)
- [Page-by-page guides](pages/README.md)
