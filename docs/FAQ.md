# FAQ

Questions people evaluating Chimera actually ask, before they've decided
to adopt it.

## Licensing & cost

**Is it really free?** Yes. Apache-2.0 — use, modify, and run it for
personal, lab, and commercial production use at no charge, subject to
preserving notices. See the README's [License](../README.md#license)
section.

**What does "Enterprise" mean here?** Production support, SLAs, and
Zyvor's other commercial products are licensed separately. Contact
sales@zyvor.dev. Nothing in this repository requires it.

## Support

**What if I find a bug?** Open a GitHub issue or PR — run `make verify`
first (build, vet, tests, `gofmt`, embedded dashboard JS syntax check).

**What if I find a security vulnerability?** This repo has no dedicated
`SECURITY.md` today — open an issue and flag it clearly as security-
sensitive rather than including exploit details in a public issue.

## Scope

**Is Chimera a replacement for real vSphere/Nutanix/AWS/Azure?** No —
explicitly not: "Chimera is a test and compatibility appliance. It is not
VMware, Nutanix, Microsoft, Amazon, Red Hat, Proxmox or cloud-vendor
software, and it is not intended to host production workloads." It exists
so you can integration-test migration/discovery/export/automation software
without provisioning the real platform.

**Which personas are actually deep vs. just protocol surfaces?** Only
**vSphere** is deep today — a real govmomi session, full inventory,
`OvfManager.CreateDescriptor`, `ExportVm`, `HttpNfcLease`
ready/complete/abort/progress, VMDK download with HTTP Range/206 resume,
and deterministic fault injection. **Nutanix Prism v3, Hyper-V WS-Man, AWS
EC2/EBS, and Azure ARM are "available" as protocol surfaces only — no
Command Center yet** (see `docs/PROVIDER_ARCHITECTURE.md`'s "Status
today"). Proxmox VE and OpenStack are roadmap items, not started.

**Why is vSphere so much deeper than the others?** It's the persona built
to match a real customer product (Transiva) end to end — see "Why this
matches Transiva" in the README and `docs/TRANSIVA.md`. The other personas
exist to broaden coverage but haven't had the same integration depth
invested yet.

## Production readiness

**What version is this?** Only `v0.1.0` is tagged. There's no
`CHANGELOG.md` in this repo yet — check git tags/releases directly for the
latest.

**Is the vSphere persona's fault injection/latency simulation
deterministic?** Yes — this is explicit in the README: "deterministic
faults" and TLS options are part of the vSphere persona's design, not
randomized behavior that would make test results flaky.

## Integration

**How do I know what's covered for my platform?** See
[`docs/TEST_MATRIX.md`](TEST_MATRIX.md) — sections 1–74 cover vSphere,
75–87 cover Nutanix/Hyper-V, 88–98 cover AWS/Azure.
