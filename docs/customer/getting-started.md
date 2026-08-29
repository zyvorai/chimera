# Getting Started with Chimera

## What you need

- Go 1.25+ to build from source, **or** a `.deb`/`.rpm` from [GitHub Releases](https://github.com/zyvorai/chimera/releases/latest)
- Browser for Command Center
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
http://localhost:8989/__chimera/
```

Log in with **admin** / **admin** (change later in Settings). Without a session the UI shows only the login screen.

## 3. Confirm the vSphere persona

```bash
./bin/chimera selftest \
  -url http://localhost:8989/sdk \
  -user administrator@vsphere.local \
  -pass vmware
```

A passing self-test proves auth, inventory, OVF descriptor, `ExportVm`, and NFC lease download.

## 4. First useful checks in the UI

1. Overview KPIs populate (requests, sessions, exports).
2. Topology shows datacenter → cluster → host → datastore.
3. VM inventory lists simulator VMs (for example `DC0_C0_RP0_VM0`).
4. Scenario Launcher: try **Clean Environment**.

## Next

- [Using the Dashboard](using-the-dashboard.md)
- [Wire Transiva](workflows.md)
- [Admin basics](admin-basics.md)
