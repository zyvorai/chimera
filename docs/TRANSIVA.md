# Transiva integration

Chimera is aligned to Transiva's actual vSphere provider path:

1. parse `vcenter_url` with govmomi
2. authenticate using `username` + `password`
3. discover/set a datacenter with `find.Finder`
4. find the selected VM
5. call `OvfManager.CreateDescriptor`
6. call `VirtualMachine.Export` (`ExportVm` SOAP method)
7. wait for a ready `HttpNfcLease`
8. iterate lease `Items`
9. download each item
10. if a partial file exists, send `Range: bytes=N-`
11. complete the lease

## Start Chimera

```bash
make build
./bin/chimera serve -config config.example.json
```

Defaults:

- URL: `http://localhost:8989/sdk`
- username: `administrator@vsphere.local`
- password: `vmware`
- VM: `DC0_C0_RP0_VM0`
- full path: `/DC0/vm/DC0_C0_RP0_VM0`

## Confirm the vSphere contract first

```bash
./bin/chimera selftest \
  -url http://localhost:8989/sdk \
  -user administrator@vsphere.local \
  -pass vmware
```

A passing self-test proves the fake vCenter can be authenticated, inventoried, exported and read through an NFC lease.

## Generate Transiva config

```bash
./scripts/make-transiva-config.sh > /tmp/transiva-chimera.yaml
cat /tmp/transiva-chimera.yaml
```

Then:

```bash
cd ../transiva
./scripts/build.sh
./bin/hyperexport --config /tmp/transiva-chimera.yaml
```

## Authentication tests

Good credentials should work. Test a rejected login with:

```yaml
vcenter_url: "http://localhost:8989/sdk"
username: "administrator@vsphere.local"
password: "wrong-password"
insecure: true
```

The simulator listener is configured with an explicit user/password, so this is strict rather than an “accept any credentials” mock.

## Resume in one Transiva run

```bash
./scripts/scenario.sh resume
```

The next NFC request is aborted after 2 MiB. `nfc_drop_next=1` means only that selected transfer is dropped. Transiva's retry opens the partial local file, sends `Range: bytes=N-`, and Chimera returns HTTP 206 with the remainder.

Reset afterwards:

```bash
./scripts/scenario.sh clean
```

## Retry testing

```bash
./scripts/scenario.sh flaky
```

This injects early HTTP 503 failures, including an NFC failure, so retry/backoff can be verified.

## Slow-link testing

```bash
./scripts/scenario.sh slow
```

The scenario adds 750 ms request latency and caps NFC bandwidth to 2 MiB/s.

## Testing conversion, not just transfer

By default, the exported `disk-0.vmdk` is deterministic test bytes. That validates Transiva through its download boundary but is not a valid disk format.

For qemu-img/hyper2kvm/GuestKit tests, use a real disposable VMDK. Two modes are available:

**Single shared file** — every VM's export lease serves the same file, which keeps conversion tests small and deterministic:

```bash
export CHIMERA_FIXTURE_VMDK=/path/to/test.vmdk
./bin/chimera serve -listen 0.0.0.0:8989
```

**Directory of VMDKs, one per VM** — point at a directory instead, and each simulated VM gets a distinct file:

```bash
export CHIMERA_FIXTURE_VMDK_DIR=/path/to/vmdks
./bin/chimera serve -listen 0.0.0.0:8989
```

Matching is a three-pass process: a file explicitly pinned to a VM from the dashboard (upload or host browser) wins first; then a file whose name (minus extension) matches a VM's name — e.g. `DC0_C0_RP0_VM0.vmdk`, matched on its own basename even if nested in a subdirectory — is assigned to that VM; any leftover files and leftover VMs are then paired off in sorted order. VMs that still get nothing keep the default generated synthetic fixture. `CHIMERA_FIXTURE_VMDK` and `CHIMERA_FIXTURE_VMDK_DIR`/`CHIMERA_FIXTURE_VMDK_DIRS` are mutually exclusive. The directory (and any `CHIMERA_FIXTURE_VMDK_DIRS` additions) is scanned recursively and re-scanned automatically every 5 seconds, so dropping in (or removing) a VMDK anywhere under it is picked up without restarting the server. See the VMDK Library card and the Fixture column in the dashboard's Inventory table (`http://localhost:8989/__chimera/`) to see which VM got which file, and `GET /__chimera/api/vmdks` for the same data as JSON.

Don't need shell/SSH access to place the file at all — the VMDK Library card's "Upload VMDK" button can upload a `.vmdk` straight from the browser, or browse and pin one already staged on the host, with an optional explicit VM assignment.

Don't have a real VMDK handy? `make fixtures` (`scripts/fetch-sample-fixtures.sh`) fetches a small, official, checksummed Alpine Linux cloud image and converts it to VMDK automatically — so the export path serves a genuinely real, valid disk image out of the box instead of the generated filler bytes. `scripts/verify-real-fixture.sh` proves this end-to-end: it downloads the *complete* exported disk (via `chimera selftest -vm <name> -save <path>`, not just the default 4KB probe) and confirms `qemu-img info` recognizes it as a valid image.

## Connecting via Transiva's web dashboard (Providers page)

Transiva's daemon (`transivad`) also has its own "Providers" UI (`Migrate hub → Providers → Connect vSphere`) as an alternative to the CLI config above. Transiva's connect form assumes `https://` for a bare `host:port`, so the **vCenter Host** field must always have an explicit scheme matching how the target Chimera is actually running:

- `CHIMERA_TLS` unset/`false` (the default): `http://<chimera-host>:8989` — a bare `host:port` here fails with `server gave HTTP response to HTTPS client`.
- `CHIMERA_TLS=true`: `https://<chimera-host>:8989` — using `http://` here instead fails differently, with `POST "/sdk": 400 Bad Request` (Chimera's listener speaks TLS only, no plain-HTTP fallback on the same port). If you flip `CHIMERA_TLS` after already saving a credential, just edit the vCenter Host field to the corrected scheme and reconnect — no need to delete it first (Transiva commit `682765c` fixed a bug where the edited field could be silently overridden by the old saved value).

Either way:

- Username: `administrator@vsphere.local`
- Password: `vmware`
- Datacenter: `DC0`

Chimera's self-signed cert works with the CLI config path because the generated YAML sets `insecure: true`. Transiva's web connect form has a matching "Skip TLS certificate verification" checkbox next to the vCenter Host field — check it when pointing at a TLS-enabled Chimera with a self-signed cert. Two real bugs were found and fixed in Transiva while testing this against a live TLS-enabled Chimera instance (both outside this repo, in `transiva-`): the connect form originally had no way to skip certificate verification at all, and separately, reconnecting after editing a previously-saved vCenter Host could silently keep using the old saved URL because a stale internal `host` field shadowed the edited one. Both are fixed as of Transiva commits `b13ee6a` and `682765c`; a form that still fails after checking the box and confirming the URL is current is worth re-checking against those, not assumed to be a fresh Chimera-side issue.

Once connected, VM discovery for that provider goes through Transiva's `GET /api/providers/vms?provider=vsphere` (not the legacy `/vms/list`), so navigate to the **VMs** tab or **Migrate hub** to see Chimera's simulated inventory rather than expecting it on the connect dialog itself.

## Running Chimera with TLS

Set `CHIMERA_TLS=true` to have Chimera serve a self-signed HTTPS listener instead of plain HTTP — the same port switches entirely to TLS (no plain-HTTP fallback on that port). To point Transiva at it:

- CLI config: `vcenter_url: "https://<chimera-host>:8989/sdk"` with `insecure: true` (already the default in `scripts/make-transiva-config.sh`'s generated config, since the cert is self-signed) so Transiva doesn't reject it for an untrusted CA.
- Web dashboard connect form: `https://<chimera-host>:8989`.
- Any `curl`/`chimera selftest` call against it needs `-k`/`-insecure=true` for the same reason.

Confirmed working end-to-end this way: `chimera selftest -insecure=true` and a live Transiva instance both talking to a TLS-enabled Chimera on the same host.

## Remote/container clients

Set `CHIMERA_PUBLIC_HOST` to an address reachable by Transiva. For example, if both are in Compose:

```bash
CHIMERA_PUBLIC_HOST=chimera:8989
```

This matters because the public host is embedded in `HttpNfcLeaseInfo.DeviceUrl`.
