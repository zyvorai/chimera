# Chimera — Customer Documentation

**Chimera** is a programmable infrastructure simulation engine for integration-testing migration, discovery, export, and automation software **without** provisioning real vSphere, Nutanix, or Hyper-V.

The **Command Center** lives at `/__chimera/` on the same listener as the fake vCenter SDK (vSphere persona). Nutanix Prism (`/api/nutanix/v3`) and Hyper-V WS-Man (`/wsman`) are available via `CHIMERA_PERSONA` as protocol surfaces without the dashboard.

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
  Command Center  →  http(s)://<host>:8989/__chimera/   (persona=vsphere)
  vSphere SDK     →  http(s)://<host>:8989/sdk
  Nutanix Prism   →  http(s)://<host>:8989/api/nutanix/v3   (persona=nutanix)
  Hyper-V WS-Man  →  http(s)://<host>:8989/wsman            (persona=hyperv)
  Simulator login →  administrator@vsphere.local / vmware   (vSphere)
  UI login        →  admin / admin  (change in Settings)
```

Chimera is a **test and compatibility appliance**. It is not VMware, Nutanix, or Microsoft software and is not for production workloads.

---

*Zyvor · [zyvor.dev](https://zyvor.dev) · Chimera · Apache-2.0*
