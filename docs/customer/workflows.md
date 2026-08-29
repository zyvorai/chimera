# Common workflows

End-to-end jobs mixing CLI and Command Center. Use `<host>` for any lab address — never hard-code deployment IPs in runbooks.

## A. Prove the vSphere contract (self-test)

```bash
./bin/chimera selftest \
  -url http://<host>:8989/sdk \
  -user administrator@vsphere.local \
  -pass vmware
```

Expect: login → inventory → OVF descriptor → ExportVm → NFC download success.

### Operate from the console (UX)

1. Sign in at `http(s)://<host>:8989/__chimera/`.
2. Run selftest from the host (command above).
3. Watch **Requests**, **Exports**, and **Avg Response** KPIs tick up.
4. Open **Live requests** — confirm SOAP/SDK/NFC paths with 2xx statuses.
5. **Top Activity** donut should gain segments after the run.

**Fail:** KPIs flat → selftest URL wrong (must be `/sdk`). Red statuses → [Fault Studio](pages/command-center/fault-studio.md) → **Reset** or **Clean Environment** scenario.

## B. Wire Transiva export

```bash
./scripts/make-transiva-config.sh > /tmp/transiva-chimera.yaml
# Point Transiva / hyperexport at that config (see docs/TRANSIVA.md)
```

Confirm `vcenter_url` uses Chimera's `<host>` (not laptop `127.0.0.1` if Transiva runs elsewhere) and `insecure: true` for lab TLS.

### Operate from the console (UX)

1. Copy a target VM from **Inventory** (row **⇩** or **Export VM** after filtering).
2. Start Transiva export; keep Command Center open on **Live requests**.
3. Monitor **Exports** KPI and NFC rows; check **Fixture** badge (`Synthetic` vs real VMDK).
4. Use footer SDK endpoint string when verifying Transiva points at the same `<host>`.

**Success:** Steady NFC traffic, rising byte counter, green statuses unless faults armed.

## C. Resume / Range download

```bash
./scripts/scenario.sh resume
# or click Resume Export in Scenario Launcher
```

Drops the next NFC stream after ~2 MiB so Transiva retries with `Range: bytes=N-` and Chimera returns HTTP 206.

### Operate from the console (UX)

1. Sign in → **Scenario Launcher** → **Resume Export** (or CLI above).
2. Run export; watch **Live requests** for initial failure then **206** on retry.
3. **Clean Environment** or Fault Studio **Reset** before the next baseline run.

**Fail:** No 206 → client may not send Range; see product `docs/TRANSIVA.md`.

## D. Fault injection for client hardening

### Operate from the console (UX)

1. Open **Fault Studio** (bolt icon or sidebar).
2. Set latency, **Fail next requests**, NFC drop fields, or bandwidth cap → **Apply policy**.
3. Re-run Transiva export or `selftest`; correlate with red rows in **Live requests**.
4. **Reset** or **Clean Environment** before benchmark runs.

**Success:** Injected errors visible in feed and error-rate KPI; clean reset returns nominal metrics.

## E. Real VMDK fixtures

1. Stage `.vmdk` files under `fixture_vmdk_dir` or use **VMDK Library → Upload VMDK**.
2. Optionally pin a file to a VM (**Manual** fixture badge in Inventory).
3. Export that VM through Transiva to exercise conversion, not only transfer.

### Operate from the console (UX)

1. **VMDK Library** → confirm file listed with assignment method.
2. **Inventory** → hover fixture badge for filename; expect **Matched** / **Manual** instead of **Synthetic**.
3. Run export; verify larger disk payload in telemetry bytes transferred.

Matching rules: see product-repo `docs/TRANSIVA.md`.

## Related pages

- [Using the Dashboard](using-the-dashboard.md)
- [Admin basics](admin-basics.md)
- [Page-by-page guides](pages/README.md)
