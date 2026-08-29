# Admin basics

| Topic | Detail |
|-------|--------|
| Listen | `listen` in JSON config — default `0.0.0.0:8989` |
| Public host | `public_host` for NFC URLs (use real hostname for remote clients) |
| TLS | `tls: true` or `CHIMERA_TLS=true` — self-signed HTTPS |
| Simulator auth | `username` / `password` — default `administrator@vsphere.local` / `vmware` |
| Admin UI login | Default `admin` / `admin`; `CHIMERA_ADMIN_USERNAME` / `CHIMERA_ADMIN_PASSWORD`; changeable in Settings |
| Admin API token | `admin_token` — default `chimera-admin` |
| Fixtures | `fixture_vmdk`, `fixture_vmdk_dir`, `fixture_vmdk_dirs`, `fixture_size_mb` |
| Estate shape | `datacenters`, `clusters`, `hosts_per_cluster`, `datastores`, `vms_per_pool` |
| Packages | `make package` → `.deb`/`.rpm`; systemd unit `chimera.service` |
| Remote deploy | `./scripts/deploy-remote.sh <host> [user]` |
| Logs | `journalctl -u chimera` or `-t chimera` |

## Endpoints

| Path | Purpose |
|------|---------|
| `/__chimera/` | Command Center UX |
| `/sdk` | vSphere SOAP/VIM for govmomi / Transiva |
| `/` | Redirects to Command Center |

## Security notes

- Change default admin credentials before exposing beyond a lab VLAN.
- Simulator credentials are strict (wrong password fails) — not an “accept anything” mock.
- Chimera is for **test and compatibility**, not production VM hosting.

See also product-repo `docs/UX.md`, `docs/TRANSIVA.md`, `docs/PROVIDER_ARCHITECTURE.md`.
