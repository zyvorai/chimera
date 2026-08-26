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

Matching is two-pass: a file whose name (minus extension) matches a VM's name — e.g. `DC0_C0_RP0_VM0.vmdk` — is assigned to that VM; any leftover files and leftover VMs are then paired off in sorted order. VMs that still get nothing keep the default generated synthetic fixture. `CHIMERA_FIXTURE_VMDK` and `CHIMERA_FIXTURE_VMDK_DIR` are mutually exclusive. The directory is scanned once at startup — restart the server to pick up added/removed files. See the VMDK Library card and the Fixture column in the dashboard's Inventory table (`http://localhost:8989/__chimera/`) to see which VM got which file, and `GET /__chimera/api/vmdks` for the same data as JSON.

Don't have a real VMDK handy? `make fixtures` (`scripts/fetch-sample-fixtures.sh`) fetches a small, official, checksummed Alpine Linux cloud image and converts it to VMDK automatically — so the export path serves a genuinely real, valid disk image out of the box instead of the generated filler bytes. `scripts/verify-real-fixture.sh` proves this end-to-end: it downloads the *complete* exported disk (via `chimera selftest -vm <name> -save <path>`, not just the default 4KB probe) and confirms `qemu-img info` recognizes it as a valid image.

## Remote/container clients

Set `CHIMERA_PUBLIC_HOST` to an address reachable by Transiva. For example, if both are in Compose:

```bash
CHIMERA_PUBLIC_HOST=chimera:8989
```

This matters because the public host is embedded in `HttpNfcLeaseInfo.DeviceUrl`.
