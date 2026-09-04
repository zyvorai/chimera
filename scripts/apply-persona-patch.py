#!/usr/bin/env python3
from pathlib import Path
p=Path("internal/config/config.go")
s=p.read_text()
s=s.replace('type Config struct {\n', 'type Config struct {\n\tPersona      string        `json:"persona"`\n', 1)
s=s.replace('return Config{\n', 'return Config{\n\t\tPersona:        "vsphere",\n', 1)
s=s.replace('func (c Config) Validate() error {\n', 'func (c Config) Validate() error {\n\tswitch strings.ToLower(strings.TrimSpace(c.Persona)) {\n\tcase "vsphere", "nutanix", "hyperv":\n\tdefault:\n\t\treturn fmt.Errorf("persona must be one of vsphere, nutanix, hyperv; got %q", c.Persona)\n\t}\n', 1)
s=s.replace('\tstr("CHIMERA_LISTEN", &c.Listen)\n', '\tstr("CHIMERA_PERSONA", &c.Persona)\n\tstr("CHIMERA_LISTEN", &c.Listen)\n', 1)
p.write_text(s)

p=Path("internal/lab/lab.go")
s=p.read_text()
needle='func Start(ctx context.Context, cfg config.Config) (*Lab, error) {\n'
repl=needle+'\tswitch strings.ToLower(strings.TrimSpace(cfg.Persona)) {\n\tcase "nutanix", "hyperv":\n\t\treturn StartHTTPPersona(ctx, cfg)\n\t}\n'
if needle not in s: raise SystemExit("cannot find lab.Start")
s=s.replace(needle,repl,1)
p.write_text(s)

p=Path("config.example.json")
s=p.read_text()
if '"persona"' not in s:
    s=s.replace('{\n','{\n  "persona": "vsphere",\n',1)
p.write_text(s)
