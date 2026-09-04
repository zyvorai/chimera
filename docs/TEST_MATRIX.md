# Chimera test matrix

Use this matrix as the acceptance suite for Chimera personas. Sections 1–74 focus on Transiva's vSphere provider; sections 75–87 cover Nutanix and Hyper-V protocol personas.

## Authentication and session

1. Valid username/password logs in.
2. Wrong password returns `InvalidLogin`.
3. Empty username is rejected.
4. Empty password is rejected.
5. Session cookie is reused across property calls.
6. Logout invalidates the session.
7. A request without a session to a protected VIM method returns `NotAuthenticated`.
8. Multiple concurrent Transiva clients can log in independently.

## Discovery and inventory

9. Default datacenter discovery succeeds.
10. Datacenter wildcard list returns all seeded datacenters.
11. VM wildcard list returns all seeded VMs.
12. Exact VM name lookup succeeds.
13. Full inventory path lookup succeeds.
14. Missing VM lookup returns a non-retryable not-found condition.
15. 100-VM inventory remains responsive.
16. 500-VM inventory can be used as a scale/latency test.
17. Powered-on and powered-off VM metadata can be queried.
18. Datastore and host references resolve through the property collector.

## OVF and export lease

19. `OvfManager.CreateDescriptor` returns a descriptor, including when called on a real authenticated session before `VirtualMachine.Export` (Transiva's actual sequence) — `integration/chimera_test.go`'s `TestCreateDescriptorThenExportVm`.
20. `VirtualMachine.Export` returns an HTTP NFC lease.
21. Lease is observable in ready state through the property collector.
22. Lease contains at least one downloadable item.
23. Lease download URL is rewritten to the public Chimera endpoint.
24. Lease completion succeeds after download.
25. Lease abort succeeds when the client cancels.
26. Two VM exports can run concurrently.

## NFC transfer

27. Full GET returns the complete fixture stream.
28. Parallel GETs work.
29. `Range: bytes=0-` returns HTTP 206.
30. `Range: bytes=N-` returns data beginning at offset N.
31. Invalid ranges return HTTP 416.
32. Bandwidth limiting slows a download deterministically.
33. A one-shot configured connection drop interrupts exactly the selected stream.
34. Retry after an injected NFC 503 succeeds.
35. A partial local file can be resumed with a correct Content-Range and without duplicated bytes.

## Retry/failure behavior

36. Inject one SOAP 503 and verify retry.
37. Inject two SOAP 503 responses and verify backoff.
38. Inject more failures than `retry_attempts` and verify terminal failure.
39. Add 750 ms latency and verify timeout handling.
40. Combine latency and 503 faults.
41. Clear the scenario and verify the next run is clean.

## Transiva-specific regression

42. `NewVSphereClient` connects using the generated Transiva config.
43. `find.Finder.DefaultDatacenter` works.
44. `finder.VirtualMachine` resolves `DC0_C0_RP0_VM0`.
45. `ExportOVF` writes the OVF descriptor.
46. `downloadFilesParallel` downloads all lease items.
47. Existing partial file causes a Range request.
48. Checkpoint save/restore resumes the file.
49. HTTP 403/404 remains non-retryable.
50. Successful export calls lease Complete and removes its checkpoint.

## Commands

```bash
make run
./bin/chimera selftest -url http://127.0.0.1:8989/sdk -user administrator@vsphere.local -pass vmware
go test ./integration/ -run Persona
./scripts/scenario.sh flaky
./scripts/scenario.sh resume
./scripts/scenario.sh clean
```

## Fixture modes

51. Generated transport fixture is deterministic and has the configured byte size.
52. Supplied real VMDK (`fixture_vmdk`, single-file mode) is served read-only through the lease alias.
53. Multiple leases can reference the same single-file real fixture independently (`fixture_vmdk` mode only).
54. Completing/aborting a lease removes only that lease mapping.
55. Conversion pipeline is run only with a valid VMDK fixture.

## Directory-matched fixtures (`fixture_vmdk_dir`)

56. A directory file whose sanitized basename matches a VM's sanitized name is assigned to that VM (`name-match`).
57. Leftover files and leftover VMs are paired off in sorted order (`round-robin`).
58. A VM with no directory-mode match still gets the existing generated synthetic fixture, unchanged.
59. Setting `fixture_vmdk` and `fixture_vmdk_dir` together is rejected at config validation.
60. Non-`.vmdk` files in the configured directory are ignored.
61. `/__chimera/api/inventory` VM entries reflect the real simulator model's names/power states — the same names `ExportVm` operates on — not a fabricated count.
62. `/__chimera/api/vmdks` lists every `.vmdk` file in the configured directory with its assignment method, including files that matched no VM.
63. A real fixture image fetched via `make fixtures` round-trips through `qemu-img info` as a valid disk image, confirmed via `scripts/verify-real-fixture.sh` (downloads the complete export via `chimera selftest -vm <name> -save <path>`, not just the default 4KB probe).
64. `fixture_vmdk_dir` is re-scanned automatically (every 5s) — a VMDK added after startup is picked up and assigned without restarting the server.
65. Scanning recurses into subdirectories; name-match keys off the file's own basename, not its full relative path, so nesting doesn't break matching.
66. `fixture_vmdk_dirs` (additional read-only roots) are scanned and matched the same way as `fixture_vmdk_dir`; browser uploads always land only in the primary directory.
67. A file with the same relative name in two different fixture roots doesn't collide — override keys are unique per root.
68. `SetOverride`/`ClearOverride` (manual assignment) take priority over name-match/round-robin, are reported as method `manual`, and reject an unknown VM name.

## Dashboard upload/browse/assign and login (`internal/gateway`)

69. Unauthenticated `POST /__chimera/api/vmdks/upload` is rejected (401); a valid upload with an explicit `vm_name` lands on disk and is reported `manual`; a non-`.vmdk` filename is rejected (400).
70. `GET /__chimera/api/vmdks/browse` lists files/subdirectories under a configured fixture root; a `..`-path-traversal attempt is rejected (400) — the browser can never reveal a path outside the configured roots, even via a symlink.
71. `POST /__chimera/api/vmdks/assign` pins a file already on disk to a VM without re-uploading it, and is rejected unauthenticated.
72. `POST /__chimera/login` accepts the configured username/password (default `admin`/`admin`) and returns the same bearer token every other protected endpoint expects; wrong credentials return 401.
73. `POST /__chimera/admin/credentials` (authenticated) changes the live admin login; the old credentials then fail `/login` and the new ones succeed. Unauthenticated calls are rejected.
74. `GET /` redirects (302) to `/__chimera/`.

## Nutanix Prism persona (`CHIMERA_PERSONA=nutanix`)

75. Basic auth rejects missing/wrong credentials with 401 + `WWW-Authenticate`.
76. `GET /api/nutanix/v3/cluster` returns cluster identity.
77. `POST /api/nutanix/v3/vms/list` returns seeded VM entities (`vms_per_pool`).
78. `GET /api/nutanix/v3/vms/{id}` returns VM detail with disk list.
79. `POST /api/nutanix/v3/vms/{id}/set_power_state` creates a task and updates power state.
80. `GET /api/nutanix/v3/tasks/{id}` returns task status.
81. `GET /api/nutanix/v3/vms/{id}/disks/{disk}/data` returns deterministic export bytes.
82. Covered by `integration/personas_e2e_test.go` (`TestNutanixDiscoveryPowerAndDiskExportE2E`).

## Hyper-V WS-Man persona (`CHIMERA_PERSONA=hyperv`)

83. Basic auth rejects missing/wrong credentials with 401.
84. WS-Man Identify returns Chimera product identity.
85. Enumerate returns an enumeration context; Pull returns `Msvm_ComputerSystem` inventory.
86. `RequestStateChange` returns async `ReturnValue` 4096 and updates VM power.
87. Covered by `integration/personas_e2e_test.go` (`TestHyperVWSManIdentifyEnumeratePullAndPowerE2E`).
