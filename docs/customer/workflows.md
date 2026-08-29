# Common workflows

## A. Prove the vSphere contract (self-test)

```bash
./bin/chimera selftest \
  -url http://localhost:8989/sdk \
  -user administrator@vsphere.local \
  -pass vmware
```

Expect: login → inventory → OVF descriptor → ExportVm → NFC download success.

## B. Wire Transiva export

```bash
./scripts/make-transiva-config.sh > /tmp/transiva-chimera.yaml
# Point Transiva / hyperexport at that config (see docs/TRANSIVA.md)
```

Confirm `vcenter_url` uses Chimera's host (not laptop `127.0.0.1` if Transiva runs elsewhere) and `insecure: true` for lab TLS.

## C. Resume / Range download

```bash
./scripts/scenario.sh resume
```

Drops the next NFC stream after ~2 MiB so Transiva retries with `Range: bytes=N-` and Chimera returns HTTP 206.

Reset faults from Fault Studio or scenario reset afterwards.

## D. Fault injection for client hardening

From Command Center → Fault Studio:

1. Set latency or next-N API failures.
2. Enable NFC stream drop / bandwidth cap.
3. Re-run Transiva export or `selftest`.
4. Clear faults before clean runs.

## E. Real VMDK fixtures

1. Stage `.vmdk` files under `fixture_vmdk_dir` (or upload via **VMDK Library**).
2. Optionally pin a file to a VM (Manual fixture badge).
3. Export that VM through Transiva to exercise conversion, not only transfer.

Matching rules: see product-repo `docs/TRANSIVA.md`.
