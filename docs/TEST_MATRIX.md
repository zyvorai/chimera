# Chimera test matrix

Use this matrix as the acceptance suite for Transiva's vSphere provider.

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

19. `OvfManager.CreateDescriptor` returns a descriptor.
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
