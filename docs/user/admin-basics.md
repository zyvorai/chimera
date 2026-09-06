# Admin basics

| Topic | Detail |
|-------|--------|
| Persona | `persona` / `CHIMERA_PERSONA` — `vsphere` (default), `nutanix`, `hyperv`, `aws`, or `azure` |
| Listen | `listen` in JSON config — default `0.0.0.0:8989` |
| Public host | `public_host` for NFC URLs (use real hostname for remote clients) |
| TLS | `tls: true` or `CHIMERA_TLS=true` — self-signed HTTPS |
| Simulator auth | `username` / `password` — default `administrator@vsphere.local` / `vmware` (also used by Nutanix/Hyper-V basic auth; AWS access key / secret; Azure subscription ID / bearer token) |
| Admin UI login | Default `admin` / `admin`; `CHIMERA_ADMIN_USERNAME` / `CHIMERA_ADMIN_PASSWORD`; changeable in Settings drawer (vSphere only) |
| Admin API token | `admin_token` — default `chimera-admin` (Bearer for curl/scenario scripts) |
| Fixtures | `fixture_vmdk`, `fixture_vmdk_dir`, `fixture_vmdk_dirs`, `fixture_size_mb` (vSphere NFC path) |
| Estate shape | `datacenters`, `clusters`, `hosts_per_cluster`, `datastores`, `vms_per_pool` |
| Packages | `make package` → `.deb`/`.rpm`; systemd unit `chimera.service` |
| Remote deploy | `./scripts/deploy-remote.sh <host> [user]` |
| Logs | `journalctl -u chimera` or `-t chimera` |

## Endpoints

| Path | Purpose | Persona |
|------|---------|---------|
| `/__chimera/` | Command Center UX | `vsphere` |
| `/sdk` | vSphere SOAP/VIM for govmomi / Transiva | `vsphere` |
| `/` | Redirects to Command Center | `vsphere` |
| `/api/nutanix/v3` | Prism-compatible REST (auth, VMs, power, disk export) | `nutanix` |
| `/wsman` | Hyper-V WS-Man SOAP (Identify, Enumerate/Pull, power) | `hyperv` |
| `/` and `/snapshots/...` | AWS EC2 Query + EBS snapshot blocks (SigV4) | `aws` |
| `/subscriptions/...` | Azure ARM Compute + managed disk SAS (Bearer) | `azure` |

Use `http(s)://<host>:8989/...` in docs and client configs — substitute your lab hostname or DNS name for `<host>`.

## Operate from the console (UX)

1. **Listen address:** Settings drawer (gear) → read-only bind shown at top; restart required to change `CHIMERA_LISTEN` / config `listen`.
2. **Rotate UI login:** Settings → new username + password → **Save login** → log out from user chip → sign in again.
3. **Verify estate:** Overview topology subtitle + Inventory VM count vs config fields above.
4. **Fixture dirs:** VMDK Library subtitle must show configured path; empty subtitle → set `fixture_vmdk_dir` and restart.
5. **API automation:** Use `Authorization: Bearer <admin_token>` for `/__chimera/scenario/*`, `/__chimera/faults`, `/__chimera/reset` — same privilege as an logged-in browser tab.
6. **Remote clients:** Footer persona strip shows public SDK URL Transiva should mirror; align `public_host` when NFC URLs must not say `localhost`.

**Empty / fail:** Settings save while logged out → login gate. VMDK upload 401 → sign in or pass Bearer token. Simulator login fails with correct UI admin → simulator creds are separate (`administrator@vsphere.local` / `vmware`).

**Success:** Listen matches curl/browser port; rotated admin login works; remote Transiva reaches `/sdk` on `<host>`.

## Security notes

- Change default admin credentials before exposing beyond a lab VLAN.
- Simulator credentials are strict (wrong password fails) — not an “accept anything” mock.
- Chimera is for **test and compatibility**, not production VM hosting.

See also product-repo `docs/UX.md`, `docs/TRANSIVA.md`, `docs/PROVIDER_ARCHITECTURE.md`.

## Related pages

- [Settings](pages/command-center/settings.md)
- [Login](pages/command-center/login.md)
- [Getting Started](getting-started.md)
