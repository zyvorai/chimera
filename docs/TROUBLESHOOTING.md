# Troubleshooting

Real issues, with the actual fix. If your symptom isn't here,
[open an issue](https://github.com/zyvorai/chimera/issues).

## My migration tool works against real vSphere but fails against Chimera

Check exactly which vSphere operations your tool calls against
`docs/PROVIDER_ARCHITECTURE.md`'s coverage notes and
`docs/TEST_MATRIX.md`'s acceptance matrix (sections 1–74) — the vSphere
persona is deep but not a byte-for-byte vCenter reimplementation. If the
operation you need isn't listed as covered, that's a real gap to file as
an issue, not a misconfiguration on your end.

## A Nutanix/Hyper-V/AWS/Azure integration only partially works

Expected — these four personas are documented as "available" **protocol
surfaces only, no Command Center yet**
(`docs/PROVIDER_ARCHITECTURE.md`'s "Status today"). If your tool needs
Command-Center-level features (the vSphere persona's dashboard/admin
APIs), that depth doesn't exist yet for the other four personas.

## VMDK download fails partway through / doesn't resume

Confirm your client actually issues an HTTP Range request (`206 Partial
Content`) rather than restarting the download from zero — the vSphere
persona's NFC download path specifically supports Range/206 resume
semantics, matching real vCenter behavior, but only if your client speaks
that protocol correctly.

## `HttpNfcLease` never reaches `ready`

Check whether you're polling the lease's progress correctly — real vCenter
(and Chimera's simulation of it) expects a client to poll
ready/complete/abort states rather than assuming synchronous completion.
See the README's description of the `ExportVm`/`HttpNfcLease` flow.

## Deterministic fault injection isn't behaving deterministically

If a "deterministic fault" scenario produces different results between
runs, check your test setup for anything non-deterministic on your own
side first (timing-dependent assertions, concurrent test execution against
the same simulated session) — the persona's fault injection is designed to
be reproducible, so an inconsistency is more often a test-harness issue
than a Chimera one. If you've ruled that out, file an issue with exact
reproduction steps.

## `make verify` fails locally but CI is green (or vice versa)

`make verify` runs "the same checks as CI (build, vet, tests, `gofmt`, and
a syntax check on the embedded dashboard JS)" per the README's
Contributing section — if it disagrees with CI, check your Go toolchain
version against `go.mod`'s pinned version first.

## Nothing here matches

Check `docs/PROVIDER_ARCHITECTURE.md` and `docs/TEST_MATRIX.md` for the
full per-persona coverage picture, then
[open an issue](https://github.com/zyvorai/chimera/issues) with the
specific operation/API call that isn't behaving as expected.
