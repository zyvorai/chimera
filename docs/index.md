# Chimera

**Chimera** is a programmable infrastructure simulation engine for
integration-testing migration, discovery, export, and automation software
without provisioning the real infrastructure platform. The architecture is
provider-persona based: **vSphere is the deepest persona** (Command Center +
govmomi), with **Nutanix Prism v3**, **Hyper-V WS-Man**, **AWS EC2/EBS**, and
**Azure ARM** available as protocol surfaces — one engine, many
infrastructure personalities.

For the full project overview, installation instructions, and roadmap, see
the **[README on GitHub](https://github.com/zyvorai/chimera)**.

## Start here

- **[FAQ](FAQ.md)** — licensing, support, and scope questions
- **[Troubleshooting](TROUBLESHOOTING.md)** — real issues with their actual fix
- **[Provider Architecture](PROVIDER_ARCHITECTURE.md)** — the persona model and what's implemented vs. protocol-only
- **[Test Matrix](TEST_MATRIX.md)** — the acceptance suite across all personas
- **[User Documentation](user/README.md)** — Command Center getting-started, dashboard, and admin guides

Chimera is Apache-2.0 licensed. It is a test and compatibility appliance —
not VMware, Nutanix, Microsoft, Amazon, Red Hat, Proxmox, or cloud-vendor
software — and it is not intended to host production workloads.
