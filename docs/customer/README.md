# Chimera — Customer Documentation

**Chimera** is a programmable infrastructure simulation engine for integration-testing migration, discovery, export, and automation software **without** provisioning real vSphere (or other planned personas).

The **Command Center** lives at `/__chimera/` on the same listener as the fake vCenter SDK.

| You want to… | Open |
|--------------|------|
| Install and open Command Center | [Getting Started](getting-started.md) |
| Learn the dashboard shell | [Using the Dashboard](using-the-dashboard.md) |
| Screen-by-screen UX | [Page-by-page guides](pages/README.md) |
| Look up surfaces | [Complete page index](PAGE_INDEX.md) |
| Ports, env, TLS, deploy | [Admin basics](admin-basics.md) |
| Transiva export / fault scenarios | [Common workflows](workflows.md) |

**→ [Docs one-pager](https://zyvor.dev/docs/chimera)** · **[GitHub](https://github.com/zyvorai/chimera)** · **[Transiva](https://zyvor.dev/docs/transiva)**

## Printable PDFs

```bash
node scripts/customer-docs/build-customer-pdfs.mjs
```

Output lands in [`pdf/`](pdf/):

- `Chimera-Customer-README.pdf`
- `Chimera-Getting-Started.pdf`
- `Chimera-Page-by-Page.pdf`
- `Chimera-Admin-Basics.pdf`

## Product at a glance

```text
  Command Center  →  http(s)://<host>:8989/__chimera/
  vSphere SDK     →  http(s)://<host>:8989/sdk
  Simulator login →  administrator@vsphere.local / vmware
  UI login        →  admin / admin  (change in Settings)
```

Chimera is a **test and compatibility appliance**. It is not VMware software and is not for production workloads.

---

*Zyvor · [zyvor.dev](https://zyvor.dev) · Chimera · Apache-2.0*
