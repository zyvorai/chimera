# VMDK Library

## Purpose

List and manage `.vmdk` fixtures under configured fixture directories; upload from the browser or pick staged host files; optionally pin to a VM.

## When to use it

- Exercise Transiva **conversion** with real disks, not only synthetic NFC transfer
- Inspect which VM (if any) each on-disk fixture is assigned to
- Add fixtures without shell access to the host (upload or browse within allowed roots only)

## How to get there

- Screen id: `#vmdks`
- Nav: **Control Center → VMDK Library** (sidebar ⛁) — badge shows file count

## Operate from the console (UX)

1. Sign in and scroll to **VMDK Library (*n*)** — subtitle shows the active directory path or **No fixture_vmdk_dir configured**.
2. Review the table: file name, size, assigned VM, assignment method (`Matched`, `Round-robin`, `Manual`, etc.).
3. Click **⇪ Upload VMDK** (requires login):
   - **Upload a file:** choose a local `.vmdk`, optionally pick **Assign to VM** (or leave **Auto** for name-match / round-robin).
   - Click **Upload** — toast confirms; library rescans within ~5s even without restart.
4. Or **pick a file already on the host:** select fixture root from dropdown, navigate folders (↑ goes up one level), click a `.vmdk` row — you must pick a VM first when assigning from browse.
5. After manual pin, confirm the VM row in [VM inventory](vm-inventory.md) shows **Manual** fixture badge.
6. Files with no VM assignment still appear here — they will not show in the VM-keyed inventory table until matched.

**Empty / fail:** **No VMDK directory configured.** → set `fixture_vmdk_dir` / `fixture_vmdk_dirs` in config and restart. **Upload failed** → check disk space and 8 GiB upload cap. Browse list empty → only configured fixture roots are visible (never arbitrary host paths). **Assign failed** → select target VM before clicking a file in browse mode.

**Success:** New files appear with size and method; assigned VM column populated; Transiva export of that VM uses the real disk.

## Related pages

- [VM inventory](vm-inventory.md)
- [Workflows](../../workflows.md)
- Product repo `docs/TRANSIVA.md` (matching rules)
- [Admin basics](../../admin-basics.md)
- [Page index](../../PAGE_INDEX.md)
