---
hero:
  eyebrow: INFRASTRUCTURE SIMULATION ENGINE
  title: Chimera
  lead: >-
    One engine, many infrastructure personalities: integration-test
    migration, discovery, export, and automation software without
    provisioning the real vSphere, Nutanix, Hyper-V, AWS, or Azure platform.
  swatches:
    - {label: "vSphere"}
    - {label: "Nutanix"}
    - {label: "Hyper-V"}
    - {label: "AWS"}
    - {label: "Azure"}
  highlights:
    - {value: "5", label: "Infrastructure personas simulated by one engine", footnote: "1"}
    - {value: "1", label: "Persona with a full Command Center dashboard today — vSphere", footnote: "2"}
    - {value: "98", label: "Checks in the cross-persona acceptance matrix", footnote: "3"}
    - {value: "6", label: "Live KPIs on the Command Center dashboard", footnote: "4"}
    - {value: "Apache-2.0", label: "Open source license"}
  hub_bands:
    - {icon: "❓", title: "FAQ", description: "Questions people evaluating Chimera actually ask, before they've decided to adopt it.", href: "FAQ.md"}
    - {icon: "🛠️", title: "Troubleshooting", description: "Real issues, with the actual fix.", href: "TROUBLESHOOTING.md"}
    - {icon: "🧩", title: "Provider Architecture", description: "The persona model and what's implemented vs. protocol-only.", href: "PROVIDER_ARCHITECTURE.md"}
    - {icon: "✅", title: "Test Matrix", description: "The acceptance suite across all personas.", href: "TEST_MATRIX.md"}
    - {icon: "📘", title: "User Documentation", description: "Command Center getting-started, dashboard, and admin guides.", href: "user/README.md"}
footnotes:
  - {marker: "1", text: "One engine, five infrastructure personas: vSphere (deepest), Nutanix Prism v3, Hyper-V WS-Man, AWS EC2/EBS, and Azure ARM.", href: "PROVIDER_ARCHITECTURE.md", href_label: "See Provider Architecture."}
  - {marker: "2", text: "vSphere is the only persona with a full Command Center dashboard today; Nutanix, Hyper-V, AWS, and Azure are protocol surfaces only, per the Provider Architecture status table.", href: "PROVIDER_ARCHITECTURE.md", href_label: "See Provider Architecture — Status today."}
  - {marker: "3", text: "The acceptance matrix runs 98 checks: 1-74 vSphere, 75-87 Nutanix/Hyper-V, 88-98 AWS/Azure.", href: "TEST_MATRIX.md", href_label: "See the Test Matrix."}
  - {marker: "4", text: "The Command Center overview header tracks six live KPIs: total requests, error rate, active vSphere sessions, transferred response bytes, NFC/export starts, and average response latency.", href: "UX.md", href_label: "See Command Center UX."}
---

For the full project overview, installation instructions, and roadmap, see
the **[README on GitHub](https://github.com/zyvorai/chimera)**.

## How Chimera compares

<div class="compare-cards" markdown="1">

- **Chimera**
  One engine, five personas — vSphere deepest (full govmomi session, `ExportVm`/`HttpNfcLease`, HTTP Range/206 resume), plus Nutanix, Hyper-V, AWS, and Azure as protocol surfaces.
- **vcsim (govmomi's own vCenter simulator)**
  Similar vSphere depth, from the same govmomi project — but vSphere-only.
- **LocalStack**
  AWS service emulation only — no vSphere, Nutanix, Hyper-V, or Azure coverage.
- **WireMock / Testcontainers**
  Generic HTTP mocking or real service containers — you hand-write the persona yourself.

</div>

*(General characterizations as of writing — verify current features against each project's own docs.)*

Chimera is Apache-2.0 licensed. It is a test and compatibility appliance —
not VMware, Nutanix, Microsoft, Amazon, Red Hat, Proxmox, or cloud-vendor
software — and it is not intended to host production workloads.
