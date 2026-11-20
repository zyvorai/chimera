# Third-party software

## Go dependencies (compiled into the `chimera` binary)

- `github.com/vmware/govmomi` v0.56.0 (Apache-2.0) — vSphere API types, client behavior, and simulator functionality.
- `github.com/google/uuid` v1.6.0 (BSD-3-Clause) — pulled in transitively by govmomi's simulator package.
- `gopkg.in/yaml.v3` v3.0.1 (Apache-2.0 / MIT) — pulled in transitively by govmomi's simulator package.

See `go.mod`/`go.sum` for the authoritative, complete dependency graph.

## Build and packaging tools (not linked into the binary)

- [`nfpm`](https://nfpm.goreleaser.com) (MIT) — used by `scripts/package.sh`/`make package` to build `.deb`/`.rpm` packages. Not required to build or run Chimera itself.
- [`qemu-img`](https://www.qemu.org/) (GPL-2.0) — used by `scripts/fetch-sample-fixtures.sh`/`scripts/verify-real-fixture.sh` to convert/inspect disk images. Invoked as an external command; not linked into Chimera.

## Sample fixture images (fetched on demand, never bundled)

`scripts/fetch-sample-fixtures.sh` downloads Alpine Linux's official "tiny" cloud image from `dl-cdn.alpinelinux.org` when an operator explicitly runs it (`make fixtures`). This image is not bundled, redistributed, or committed by Chimera — it's fetched directly from Alpine's own infrastructure, checksum-verified against Alpine's published `.sha512`, and used purely as sample disk content for local testing. See [alpinelinux.org](https://www.alpinelinux.org) for Alpine's own licensing.

## Trademarks

VMware, vSphere, vCenter, and related marks are trademarks of their respective owners. Chimera is an independent compatibility-testing project and is not VMware vCenter Server. Alpine Linux is a trademark of its respective owners; Chimera's use of Alpine's cloud images is limited to fetching official, unmodified images for local test fixtures.
