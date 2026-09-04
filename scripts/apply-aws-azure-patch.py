#!/usr/bin/env python3
"""Apply AWS + Azure persona wiring to a Chimera checkout that already has persona support."""
from pathlib import Path
import shutil
import subprocess


def replace_once(text: str, old: str, new: str, label: str) -> str:
    if old in text:
        return text.replace(old, new, 1)
    if new in text:
        return text
    raise SystemExit(f"cannot patch {label}: expected source text not found")

# Allow cloud persona names in config validation.
p = Path("internal/config/config.go")
s = p.read_text()
s = replace_once(
    s,
    'case "vsphere", "nutanix", "hyperv":',
    'case "vsphere", "nutanix", "hyperv", "aws", "azure":',
    "config persona validation",
)
s = s.replace(
    'persona must be one of vsphere, nutanix, hyperv; got %q',
    'persona must be one of vsphere, nutanix, hyperv, aws, azure; got %q',
)
p.write_text(s)

# Route AWS/Azure away from the govmomi vSphere backend.
p = Path("internal/lab/lab.go")
s = p.read_text()
s = replace_once(
    s,
    'case "nutanix", "hyperv":',
    'case "nutanix", "hyperv", "aws", "azure":',
    "lab persona dispatch",
)
p.write_text(s)

# Add cloud handlers to the HTTP persona server.
p = Path("internal/lab/persona.go")
s = p.read_text()
if 'internal/personas/aws' not in s:
    s = replace_once(
        s,
        '\t"github.com/zyvorai/chimera/internal/config"\n',
        '\t"github.com/zyvorai/chimera/internal/config"\n\tawspersona "github.com/zyvorai/chimera/internal/personas/aws"\n\tazurepersona "github.com/zyvorai/chimera/internal/personas/azure"\n',
        "persona imports",
    )
s = s.replace(
    "// StartHTTPPersona starts the Nutanix Prism or Hyper-V WS-Man persona.\n// vSphere continues to use the existing govmomi-backed Start path.",
    "// StartHTTPPersona starts the Nutanix, Hyper-V, AWS, or Azure persona.\n// vSphere continues to use the existing govmomi-backed Start path.",
)
if 'case "aws":' not in s:
    anchor = '''\tcase "hyperv":\n\t\th = hyperv.New(cfg.Username, cfg.Password, cfg.VMsPerPool)\n\t\tendpoint = "/wsman"\n'''
    addition = anchor + '''\tcase "aws":\n\t\th = awspersona.New(cfg.Username, cfg.Password, cfg.VMsPerPool)\n\t\tendpoint = "/"\n\tcase "azure":\n\t\th = azurepersona.New(cfg.Username, cfg.Password, cfg.VMsPerPool)\n\t\tendpoint = "/subscriptions/" + cfg.Username + "/providers/Microsoft.Compute/virtualMachines"\n'''
    s = replace_once(s, anchor, addition, "HTTP persona cases")
p.write_text(s)

# Add deterministic list ordering and generic async tasks used by Azure.
p = Path("internal/personas/common/model.go")
s = p.read_text()
if '"sort"' not in s:
    s = replace_once(s, '\t"fmt"\n', '\t"fmt"\n\t"sort"\n', "common imports")
if 'sort.Slice(out' not in s:
    s = replace_once(
        s,
        '''\tfor _, vm := range s.VMs {\n\t\tout = append(out, *vm)\n\t}\n\treturn out\n''',
        '''\tfor _, vm := range s.VMs {\n\t\tout = append(out, *vm)\n\t}\n\tsort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })\n\treturn out\n''',
        "deterministic VM list",
    )
if 'func (s *Store) NewTask' not in s:
    new_task = '''func (s *Store) NewTask(operation, targetID string) Task {\n\ts.mu.Lock()\n\tdefer s.mu.Unlock()\n\tseed := sha256.Sum256([]byte(operation + targetID + time.Now().UTC().Format(time.RFC3339Nano)))\n\ttask := Task{ID: "task-" + hex.EncodeToString(seed[:6]), State: "SUCCEEDED", Operation: operation, VMID: targetID, CreatedAt: time.Now().UTC()}\n\ts.Tasks[task.ID] = &task\n\treturn task\n}\n\n'''
    s = replace_once(s, 'func (s *Store) Task(id string) (Task, bool) {\n', new_task + 'func (s *Store) Task(id string) (Task, bool) {\n', "generic task support")
p.write_text(s)

# Format changed Go files if gofmt is available.
gofmt = shutil.which("gofmt")
if gofmt:
    subprocess.run([
        gofmt, "-w",
        "internal/config/config.go",
        "internal/lab/lab.go",
        "internal/lab/persona.go",
        "internal/personas/common/model.go",
        "internal/personas/aws/server.go",
        "internal/personas/azure/server.go",
        "integration/cloud_personas_e2e_test.go",
    ], check=True)

print("AWS and Azure persona wiring applied.")
