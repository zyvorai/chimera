# Validation

Chimera targets **Go 1.25+** because the vSphere persona is pinned to `github.com/vmware/govmomi v0.56.0`. Nutanix, Hyper-V, AWS, and Azure personas use the Go standard library only (no extra module deps).

## Full verification

```bash
make verify
```

(equivalently `./scripts/verify.sh`). This performs:

1. `gofmt` cleanliness check
2. `go test ./...` (including `internal/fixture` tests, the `integration/` govmomi vSphere E2E, Nutanix/Hyper-V persona E2E, and AWS/Azure cloud persona E2E)
3. `go vet ./...`
4. `go build -trimpath -o bin/chimera ./cmd/chimera`
5. optional embedded-dashboard JavaScript syntax validation when Node.js is installed

GitHub Actions runs the same `go test`/`vet`/`build`/`gofmt` path on every push (`.github/workflows/ci.yml`, Go 1.25.x).

## End-to-end validation beyond `go test`

Unit and integration tests cover the API surface; two additional scripts validate things that need a running process and/or real disk images:

- **`scripts/verify-real-fixture.sh`** — fetches a real sample VMDK (`make fixtures`), starts a throwaway `chimera serve`, downloads the *complete* export of the VM it was assigned to (`chimera selftest -vm <name> -save <path>`, not just the default 4KB probe), and confirms `qemu-img info` recognizes the result as a valid disk image. This is the proof that exported disks are real, not filler bytes.
- **`scripts/deploy-remote.sh`** — beyond installing Chimera as a systemd service on a remote host, it also runs `chimera selftest` against the freshly deployed instance and points `scripts/transiva-smoke.sh` at it, so a deploy is only reported successful once login → inventory → `ExportVm` → NFC download have actually been exercised against the real network path.

This project has also been validated against a live deployment alongside [Transiva](https://github.com/ssahani/transiva) (a real vSphere migration tool): pointing Transiva's `vcenter_url` at a running Chimera instance and confirming it can log in, discover VMs, and export a real, `qemu-img`-valid disk end to end. See `docs/TRANSIVA.md`.

Protocol persona smoke (no external cloud accounts required):

```bash
go test ./integration/ -run 'Persona|AWS|Azure'
```

## Test matrix

See [`docs/TEST_MATRIX.md`](TEST_MATRIX.md) for the full acceptance matrix (vSphere, Nutanix Prism, Hyper-V WS-Man, AWS EC2/EBS, and Azure ARM checks).
