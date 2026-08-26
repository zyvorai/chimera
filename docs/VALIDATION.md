# Validation

Chimera targets **Go 1.25+** because the vSphere persona is pinned to `github.com/vmware/govmomi v0.56.0`.

## Full verification on a Go 1.25+ machine

```bash
./scripts/verify.sh
```

The verification script performs:

1. `gofmt` cleanliness check
2. `go test ./...`
3. `go vet ./...`
4. `go build -trimpath -o bin/chimera ./cmd/chimera`
5. optional embedded-dashboard JavaScript syntax validation when Node.js is installed

GitHub Actions runs the Go test/vet/build path with Go 1.25.x.

## Validation performed while generating this repository

The build environment available to the assistant contains Go 1.23.2. The final module intentionally requires Go 1.25.0, and outbound toolchain/module downloads are blocked in that environment. Therefore a genuine final `go test ./...` against govmomi v0.56.0 cannot be executed there.

To still validate the code changed for the Command Center, a temporary verification copy was made with only its Go directive lowered to 1.23.0. No source files were altered in that copy. The packages that do not import govmomi were then compiled and tested:

```text
go test ./internal/gateway ./internal/config ./internal/faults ./internal/fixture
PASS

go vet ./internal/gateway ./internal/config ./internal/faults ./internal/fixture
PASS

go test -race ./internal/gateway ./internal/config ./internal/faults ./internal/fixture
PASS
```

The embedded dashboard was additionally checked with:

```text
node --check <extracted embedded JavaScript>
PASS

Python html.parser over embedded document
PASS

duplicate DOM id scan
PASS (0 duplicates)
```

The gateway tests cover authentication, public bootstrap/inventory APIs, the new live telemetry endpoint, backend NFC Range behavior, and local fixture Range/206 behavior.

The remaining Go 1.25/govmomi-backed integration test is `integration/chimera_test.go`; it exercises login, inventory, `ExportVm`, `HttpNfcLease`, NFC download, and bad-password rejection.
