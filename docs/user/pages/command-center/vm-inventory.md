# VM inventory

## Purpose

Searchable, paginated table of simulator VMs with power state, CPU/memory/disk, datastore, and Fixture badge (real VMDK vs synthetic).

## When to use it

- Pick an export target for Transiva or manual `ExportVm` tests
- Confirm fixture matching (`Matched`, `Round-robin`, `Shared`, `Manual`, `Synthetic`)
- Filter powered-on VMs before copying a name to the clipboard

## How to get there

- Screen id: `#inventory`
- Nav: **Control Center → Inventory** (sidebar ▤) — badge shows total VM count
- Global search (**⌘K** / Ctrl+K) filters this table and scrolls to it

## Operate from the console (UX)

1. Sign in and scroll to **Virtual Machines (*n*)** (or use sidebar **Inventory**).
2. Use **Search VM by name…** or the top-bar global search — both filter the same table.
3. Set **All Power States** to **Powered On** or **Powered Off** when narrowing export candidates.
4. Scan columns: vCPU, memory, disk, datastore, **Fixture** badge (hover for underlying filename).
5. Click the **⇩** icon on a row to copy that VM name to the clipboard (toast confirms selection).
6. Click **⬡ Export VM** in the card header to copy the first row matching current filters — useful after power-state filter.
7. Paginate with numbered page buttons (5 VMs per page); sidebar count should match config `vms_per_pool` totals.

**Empty / fail:** **No matching virtual machines.** with empty search → check simulator counts in config and restart Chimera after estate changes. All **Synthetic** when you expected real disks → stage VMDKs in [VMDK Library](vmdk-library.md) or set `fixture_vmdk_dir`. Copy actions toast **No VM selected** → relax filters.

**Success:** Expected VM names (for example `DC0_C0_RP0_VM0`) listed; fixture badges reflect matching rules; copied name pastes into Transiva config or shell scripts.

## Related pages

- [VMDK Library](vmdk-library.md)
- [Topology](topology.md)
- [Workflows](../../workflows.md)
- [Getting Started](../../getting-started.md)
- [Page index](../../PAGE_INDEX.md)
