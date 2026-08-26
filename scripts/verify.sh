#!/usr/bin/env bash
set -euo pipefail

printf '==> Go version\n'
go version

printf '\n==> Format check\n'
unformatted="$(gofmt -l $(find . -name '*.go' -type f))"
if [[ -n "$unformatted" ]]; then
  echo "Unformatted Go files:" >&2
  echo "$unformatted" >&2
  exit 1
fi

printf '\n==> Unit + integration tests\n'
go test ./...

printf '\n==> Vet\n'
go vet ./...

printf '\n==> Build\n'
mkdir -p bin
go build -trimpath -o bin/chimera ./cmd/chimera

if command -v node >/dev/null 2>&1; then
  printf '\n==> Embedded dashboard JavaScript syntax\n'
  python3 - <<'PY'
from pathlib import Path
s=Path('internal/gateway/ui.go').read_text()
a=s.index('const dashboardHTML = `')+len('const dashboardHTML = `')
b=s.rindex('`')
h=s[a:b]
js=h.split('<script>',1)[1].split('</script>',1)[0]
Path('/tmp/chimera-dashboard.js').write_text(js)
PY
  node --check /tmp/chimera-dashboard.js
else
  printf '\n==> node not installed; skipped optional dashboard JS syntax check\n'
fi

printf '\nChimera verification passed.\n'
