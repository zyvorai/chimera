// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/zyvorai/chimera/internal/faults"
	"github.com/zyvorai/chimera/internal/fixture"
)

// maxVMDKUploadBytes bounds a single /__chimera/api/vmdks/upload request body.
const maxVMDKUploadBytes = 8 << 30 // 8 GiB

type Meta struct {
	Version       string
	Persona       string
	Username      string
	Datacenters   int
	Clusters      int
	Hosts         int
	Datastores    int
	VMs           int
	FixtureSizeMB int
	TLS           bool
	// Listen is the configured bind address, shown read-only in the
	// dashboard's Settings drawer (changing it requires a restart).
	Listen string
	// AdminUsername/AdminPassword seed the Gateway's live, mutable login
	// credentials at construction — see Gateway.adminUser/adminPass.
	AdminUsername string
	AdminPassword string
}

type Gateway struct {
	backend    *url.URL
	publicBase string
	proxy      *httputil.ReverseProxy
	faults     *faults.State
	fixtures   *fixture.Store
	registry   *simulator.Registry
	adminToken string
	started    time.Time
	meta       Meta
	telemetry  *telemetry

	// credMu guards adminUser/adminPass — unlike the rest of Meta (mostly
	// static deployment info read without locking), these can change live
	// via the Settings drawer's "change admin login" action.
	credMu    sync.RWMutex
	adminUser string
	adminPass string
}

func New(backend *url.URL, publicBase, adminToken string, fs *faults.State, fixtures *fixture.Store, registry *simulator.Registry, meta ...Meta) *Gateway {
	target := &url.URL{Scheme: backend.Scheme, Host: backend.Host}
	p := httputil.NewSingleHostReverseProxy(target)
	original := p.ModifyResponse
	p.ModifyResponse = func(res *http.Response) error {
		if original != nil {
			if err := original(res); err != nil {
				return err
			}
		}
		ct := strings.ToLower(res.Header.Get("Content-Type"))
		if !strings.Contains(ct, "xml") && !strings.Contains(ct, "text") && !strings.Contains(ct, "json") {
			return nil
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		_ = res.Body.Close()
		backendBase := backend.Scheme + "://" + backend.Host
		b = bytes.ReplaceAll(b, []byte(backendBase), []byte(publicBase))
		res.Body = io.NopCloser(bytes.NewReader(b))
		res.ContentLength = int64(len(b))
		res.Header.Set("Content-Length", strconv.Itoa(len(b)))
		return nil
	}
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "chimera upstream error: "+err.Error(), http.StatusBadGateway)
	}
	m := Meta{Version: "dev", Persona: "vSphere", VMs: 1}
	if len(meta) > 0 {
		m = meta[0]
	}
	return &Gateway{
		backend: backend, publicBase: publicBase, proxy: p, faults: fs, fixtures: fixtures, registry: registry,
		adminToken: adminToken, started: time.Now(), meta: m, telemetry: newTelemetry(),
		adminUser: m.AdminUsername, adminPass: m.AdminPassword,
	}
}

// checkLogin reports whether username/password match the current admin
// login credentials, using a constant-time comparison. There is no
// lockout/rate-limiting — Chimera is a test-lab simulator, not a hardened
// production auth system.
func (g *Gateway) checkLogin(username, password string) bool {
	g.credMu.RLock()
	defer g.credMu.RUnlock()
	return subtle.ConstantTimeCompare([]byte(username), []byte(g.adminUser)) == 1 &&
		subtle.ConstantTimeCompare([]byte(password), []byte(g.adminPass)) == 1
}

// setCredentials replaces the live admin login username/password. It does
// not affect adminToken (the bearer secret every admin API call already
// uses) — only future /login attempts need the new credentials.
func (g *Gateway) setCredentials(username, password string) {
	g.credMu.Lock()
	defer g.credMu.Unlock()
	g.adminUser, g.adminPass = username, password
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" && r.Method == http.MethodGet {
		http.Redirect(w, r, "/__chimera/", http.StatusFound)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/__chimera") {
		g.serveAdmin(w, r)
		return
	}
	cw := &captureWriter{ResponseWriter: w}
	started := g.telemetry.begin(r)
	defer func() {
		g.telemetry.finish(r, cw.status, cw.bytes, started)
	}()

	isLocalNFC := strings.HasPrefix(r.URL.Path, "/chimera-nfc/")
	isNFC := isLocalNFC || strings.HasPrefix(r.URL.Path, "/nfc/")
	delay, fail, status, drop, dropAfter, bps := g.faults.Before(isNFC)
	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-r.Context().Done():
			return
		}
	}
	if fail {
		http.Error(cw, "chimera injected fault", status)
		return
	}
	if isLocalNFC {
		g.serveLocalNFC(cw, r, drop, dropAfter, bps)
		return
	}
	if isNFC && (r.Header.Get("Range") != "" || drop || bps > 0) {
		g.serveBackendNFC(cw, r, drop, dropAfter, bps)
		return
	}
	g.proxy.ServeHTTP(cw, r)
}

func (g *Gateway) serveLocalNFC(w http.ResponseWriter, r *http.Request, drop bool, dropAfter, bps int64) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/chimera-nfc/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || path.Base(parts[1]) != parts[1] {
		http.NotFound(w, r)
		return
	}
	if g.fixtures == nil {
		http.NotFound(w, r)
		return
	}
	p, ok := g.fixtures.Lookup(parts[0], parts[1])
	if !ok {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	size := st.Size()
	start, end, partial, err := parseRange(r.Header.Get("Range"), size)
	if err != nil {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.Error(w, "bad range", http.StatusRequestedRangeNotSatisfiable)
		return
	}
	length := end - start + 1
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	if partial {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if r.Method == http.MethodHead {
		return
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return
	}
	var src io.Reader = io.LimitReader(f, length)
	if bps > 0 {
		src = &rateReader{r: src, bps: bps}
	}
	if drop && dropAfter > 0 && dropAfter < length {
		_, _ = io.CopyN(w, src, dropAfter)
		if fl, ok := w.(http.Flusher); ok {
			fl.Flush()
		}
		panic(http.ErrAbortHandler)
	}
	_, _ = io.Copy(w, src)
}

func (g *Gateway) serveBackendNFC(w http.ResponseWriter, r *http.Request, drop bool, dropAfter, bps int64) {
	upstream := *r.URL
	upstream.Scheme = g.backend.Scheme
	upstream.Host = g.backend.Host

	req, err := http.NewRequestWithContext(r.Context(), r.Method, upstream.String(), nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for k, vv := range r.Header {
		if strings.EqualFold(k, "Range") || strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	defer res.Body.Close()

	for k, vv := range res.Header {
		if hopHeader(k) || strings.EqualFold(k, "Content-Length") {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Accept-Ranges", "bytes")

	var start int64
	if rh := r.Header.Get("Range"); rh != "" {
		start, err = parseRangeStart(rh)
		if err != nil {
			http.Error(w, "bad range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if start > 0 {
			if _, err := io.CopyN(io.Discard, res.Body, start); err != nil {
				http.Error(w, "range starts beyond object", http.StatusRequestedRangeNotSatisfiable)
				return
			}
		}
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(res.StatusCode)
	}

	var src io.Reader = res.Body
	if bps > 0 {
		src = &rateReader{r: src, bps: bps}
	}
	if drop && dropAfter > 0 {
		_, _ = io.CopyN(w, src, dropAfter)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		panic(http.ErrAbortHandler)
	}
	_, _ = io.Copy(w, src)
}

func parseRange(v string, size int64) (start, end int64, partial bool, err error) {
	if size < 0 {
		return 0, 0, false, fmt.Errorf("invalid size")
	}
	if size == 0 {
		if v != "" {
			return 0, 0, false, fmt.Errorf("range on empty object")
		}
		return 0, -1, false, nil
	}
	if v == "" {
		return 0, size - 1, false, nil
	}
	if !strings.HasPrefix(v, "bytes=") || strings.Contains(v, ",") {
		return 0, 0, false, fmt.Errorf("unsupported range")
	}
	spec := strings.TrimPrefix(v, "bytes=")
	pair := strings.SplitN(spec, "-", 2)
	if len(pair) != 2 || pair[0] == "" { // suffix ranges deliberately not supported
		return 0, 0, false, fmt.Errorf("unsupported range")
	}
	start, err = strconv.ParseInt(pair[0], 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false, fmt.Errorf("invalid range")
	}
	end = size - 1
	if pair[1] != "" {
		end, err = strconv.ParseInt(pair[1], 10, 64)
		if err != nil || end < start {
			return 0, 0, false, fmt.Errorf("invalid range")
		}
		if end >= size {
			end = size - 1
		}
	}
	return start, end, true, nil
}

func parseRangeStart(v string) (int64, error) {
	if !strings.HasPrefix(v, "bytes=") {
		return 0, fmt.Errorf("unsupported range")
	}
	s := strings.TrimPrefix(v, "bytes=")
	if i := strings.IndexByte(s, '-'); i >= 0 {
		s = s[:i]
	}
	if s == "" {
		return 0, fmt.Errorf("suffix ranges unsupported")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid range")
	}
	return n, nil
}

func hopHeader(k string) bool {
	switch strings.ToLower(k) {
	case "connection", "proxy-connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailers", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

type rateReader struct {
	r   io.Reader
	bps int64
}

func (r *rateReader) Read(p []byte) (int, error) {
	if r.bps <= 0 {
		return r.r.Read(p)
	}
	max := int(r.bps / 10)
	if max < 1024 {
		max = 1024
	}
	if len(p) > max {
		p = p[:max]
	}
	n, err := r.r.Read(p)
	if n > 0 {
		time.Sleep(time.Duration(float64(time.Second) * float64(n) / float64(r.bps)))
	}
	return n, err
}

func (g *Gateway) serveAdmin(w http.ResponseWriter, r *http.Request) {
	public := r.URL.Path == "/__chimera" || r.URL.Path == "/__chimera/" ||
		r.URL.Path == "/__chimera/health" || r.URL.Path == "/__chimera/api/bootstrap" ||
		r.URL.Path == "/__chimera/api/inventory" || r.URL.Path == "/__chimera/api/telemetry" ||
		r.URL.Path == "/__chimera/api/vmdks" || r.URL.Path == "/__chimera/login"
	if g.adminToken != "" && !public && r.Header.Get("Authorization") != "Bearer "+g.adminToken {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.URL.Path == "/__chimera" || r.URL.Path == "/__chimera/":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, dashboardHTML)
	case r.URL.Path == "/__chimera/health":
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true, "uptime": time.Since(g.started).String(), "uptime_seconds": int(time.Since(g.started).Seconds()),
			"backend": g.backend.Host, "public": g.publicBase, "persona": g.meta.Persona,
		})
	case r.URL.Path == "/__chimera/api/bootstrap":
		_ = json.NewEncoder(w).Encode(g.bootstrap())
	case r.URL.Path == "/__chimera/api/inventory":
		_ = json.NewEncoder(w).Encode(g.inventory())
	case r.URL.Path == "/__chimera/api/telemetry":
		_ = json.NewEncoder(w).Encode(g.telemetry.snapshot())
	case r.URL.Path == "/__chimera/api/vmdks":
		_ = json.NewEncoder(w).Encode(g.vmdks())
	case r.URL.Path == "/__chimera/api/vmdks/upload" && r.Method == http.MethodPost:
		g.uploadVMDK(w, r)
	case r.URL.Path == "/__chimera/login" && r.Method == http.MethodPost:
		g.login(w, r)
	case r.URL.Path == "/__chimera/admin/credentials" && r.Method == http.MethodPost:
		g.changeCredentials(w, r)
	case r.URL.Path == "/__chimera/api/vmdks/browse" && r.Method == http.MethodGet:
		g.browseVMDKs(w, r)
	case r.URL.Path == "/__chimera/api/vmdks/assign" && r.Method == http.MethodPost:
		g.assignVMDK(w, r)
	case r.URL.Path == "/__chimera/state" && r.Method == http.MethodGet:
		_ = json.NewEncoder(w).Encode(g.faults.Snapshot())
	case r.URL.Path == "/__chimera/faults" && r.Method == http.MethodPost:
		var s faults.Snapshot
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		g.faults.Apply(s)
		_ = json.NewEncoder(w).Encode(g.faults.Snapshot())
	case r.URL.Path == "/__chimera/reset" && r.Method == http.MethodPost:
		g.faults.Reset()
		_ = json.NewEncoder(w).Encode(g.faults.Snapshot())
	case r.URL.Path == "/__chimera/scenario/clean" && r.Method == http.MethodPost:
		g.faults.Reset()
		_ = json.NewEncoder(w).Encode(g.faults.Snapshot())
	case r.URL.Path == "/__chimera/scenario/slow" && r.Method == http.MethodPost:
		g.faults.Apply(faults.Snapshot{LatencyMS: 750, FailStatus: 503, BandwidthBPS: 2 * 1024 * 1024})
		_ = json.NewEncoder(w).Encode(g.faults.Snapshot())
	case r.URL.Path == "/__chimera/scenario/flaky" && r.Method == http.MethodPost:
		g.faults.Apply(faults.Snapshot{LatencyMS: 150, FailNext: 2, NFCFailNext: 1, FailStatus: 503})
		_ = json.NewEncoder(w).Encode(g.faults.Snapshot())
	case r.URL.Path == "/__chimera/scenario/resume" && r.Method == http.MethodPost:
		g.faults.Apply(faults.Snapshot{NFCDropNext: 1, NFCDropAfter: 2 * 1024 * 1024, FailStatus: 503})
		_ = json.NewEncoder(w).Encode(g.faults.Snapshot())
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (g *Gateway) bootstrap() map[string]any {
	providers := []map[string]any{
		{"id": "vsphere", "name": "VMware vSphere", "short": "vSphere", "status": "active", "protocol": "SOAP / VIM / NFC", "accent": "violet"},
		{"id": "nutanix", "name": "Nutanix Prism", "short": "Nutanix", "status": "available", "protocol": "Prism v3/v4 API", "accent": "cyan"},
		{"id": "proxmox", "name": "Proxmox VE", "short": "Proxmox", "status": "planned", "protocol": "REST API", "accent": "orange"},
		{"id": "openstack", "name": "OpenStack", "short": "OpenStack", "status": "planned", "protocol": "Nova / Glance", "accent": "rose"},
		{"id": "hyperv", "name": "Microsoft Hyper-V", "short": "Hyper-V", "status": "available", "protocol": "WinRM / WMI", "accent": "blue"},
		{"id": "aws", "name": "Amazon Web Services", "short": "AWS", "status": "available", "protocol": "EC2 Query / EBS blocks", "accent": "orange"},
		{"id": "azure", "name": "Microsoft Azure", "short": "Azure", "status": "available", "protocol": "ARM Compute / Disks", "accent": "green"},
	}
	return map[string]any{
		"product": "Chimera", "tagline": "One engine. Many infrastructure personalities.", "version": g.meta.Version,
		"persona": g.meta.Persona, "username": g.meta.Username, "endpoint": g.publicBase + "/sdk", "public_base": g.publicBase,
		"tls": g.meta.TLS, "datacenters": g.meta.Datacenters, "clusters": g.meta.Clusters, "hosts": g.meta.Hosts,
		"datastores": g.meta.Datastores, "vms": g.vmCount(), "fixture_size_mb": g.meta.FixtureSizeMB, "providers": providers,
		"listen": g.meta.Listen,
	}
}

// vmCount prefers the live simulator registry's real VM count, falling back
// to the static Meta.VMs count when no registry is wired (e.g. in tests).
func (g *Gateway) vmCount() int {
	if g.registry != nil {
		return len(g.registry.All("VirtualMachine"))
	}
	return g.meta.VMs
}

func (g *Gateway) inventory() map[string]any {
	vms := make([]map[string]any, 0)
	if g.registry != nil {
		entities := g.registry.All("VirtualMachine")
		sort.Slice(entities, func(i, j int) bool {
			return entities[i].Entity().Name < entities[j].Entity().Name
		})
		for _, e := range entities {
			vm, ok := e.(*simulator.VirtualMachine)
			if !ok || vm.Name == "" {
				continue
			}
			cpu, memGB, diskGB := vmHardware(vm)
			source, file := "", ""
			if g.fixtures != nil {
				source, file = g.fixtures.SourceFor(vm.Name)
			}
			vms = append(vms, map[string]any{
				"id": vm.Self.Value, "name": vm.Name, "state": string(vm.Runtime.PowerState),
				"cpu": cpu, "memory_gb": memGB, "disk_gb": diskGB,
				"datastore": g.datastoreName(vm), "network": "VM Network", "exportable": true,
				"fixture_source": source, "fixture_file": file,
			})
		}
	}
	return map[string]any{"persona": g.meta.Persona, "virtual_machines": vms, "total": len(vms)}
}

func vmHardware(vm *simulator.VirtualMachine) (cpu int, memGB float64, diskGB int) {
	if vm.Config == nil {
		return 0, 0, 0
	}
	hw := vm.Config.Hardware
	cpu = int(hw.NumCPU)
	memGB = float64(hw.MemoryMB) / 1024
	var kb int64
	for _, d := range hw.Device {
		if disk, ok := d.(*types.VirtualDisk); ok {
			kb += disk.CapacityInKB
		}
	}
	diskGB = int(kb / (1024 * 1024))
	return cpu, memGB, diskGB
}

func (g *Gateway) datastoreName(vm *simulator.VirtualMachine) string {
	if g.registry == nil || len(vm.Datastore) == 0 {
		return ""
	}
	if ds, ok := g.registry.Get(vm.Datastore[0]).(*simulator.Datastore); ok {
		return ds.Name
	}
	return ""
}

func (g *Gateway) vmdks() map[string]any {
	var list []fixture.Assignment
	dir := ""
	var roots []string
	if g.fixtures != nil {
		list = g.fixtures.Assignments()
		dir = g.fixtures.Directory()
		roots = g.fixtures.Roots()
	}
	matched, robin, manual, unassigned := 0, 0, 0, 0
	for _, a := range list {
		switch a.Method {
		case fixture.MethodNameMatch:
			matched++
		case fixture.MethodRoundRobin:
			robin++
		case fixture.MethodManual:
			manual++
		default:
			unassigned++
		}
	}
	if list == nil {
		list = []fixture.Assignment{}
	}
	return map[string]any{
		"directory": dir, "roots": roots, "files": list, "total": len(list),
		"matched": matched, "round_robin": robin, "manual": manual, "unassigned": unassigned,
	}
}

// uploadVMDK accepts a multipart "file" field (a .vmdk), writes it into the
// configured fixture_vmdk_dir, and — via an optional "vm_name" form field —
// optionally pins it to a specific VM (fixture.Store.SetOverride), otherwise
// falling back to the store's usual name-match/round-robin scan.
func (g *Gateway) uploadVMDK(w http.ResponseWriter, r *http.Request) {
	if g.fixtures == nil || g.fixtures.Directory() == "" {
		http.Error(w, "fixture_vmdk_dir is not configured on this deployment", http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxVMDKUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing \"file\" form field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	name := filepath.Base(header.Filename)
	if name == "" || name == "." || name == string(filepath.Separator) || !strings.EqualFold(filepath.Ext(name), ".vmdk") {
		http.Error(w, "only .vmdk files are accepted", http.StatusBadRequest)
		return
	}

	dest := filepath.Join(g.fixtures.Directory(), name)
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := out.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if vmName := strings.TrimSpace(r.FormValue("vm_name")); vmName != "" {
		if err := g.fixtures.SetOverride(0, name, vmName); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else if err := g.fixtures.Rescan(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(g.vmdks())
}

// browseEntry is one file or subdirectory returned by browseVMDKs.
type browseEntry struct {
	Name      string `json:"name"`
	IsDir     bool   `json:"is_dir"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

// resolveFixtureBrowsePath validates that root/path (root an index into
// g.fixtures.Roots(), path relative and possibly containing "..") resolves
// to somewhere inside that root, even accounting for symlinks. This is the
// hard boundary for /api/vmdks/browse: an operator can only ever see what's
// already inside a directory they explicitly configured as a fixture
// source — never an arbitrary path on the host.
func (g *Gateway) resolveFixtureBrowsePath(rootIdx int, relPath string) (string, error) {
	roots := g.fixtures.Roots()
	if rootIdx < 0 || rootIdx >= len(roots) {
		return "", fmt.Errorf("unknown fixture root: %d", rootIdx)
	}
	rootAbs := roots[rootIdx]
	target := filepath.Join(rootAbs, relPath)
	if rel, err := filepath.Rel(rootAbs, target); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes the fixture root")
	}
	realRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", err
	}
	realTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", err
	}
	if rel, err := filepath.Rel(realRoot, realTarget); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes the fixture root")
	}
	return realTarget, nil
}

// browseVMDKs lists .vmdk files and subdirectories under one configured
// fixture root, so the dashboard can offer picking a VMDK already staged on
// the host instead of uploading it through the browser.
func (g *Gateway) browseVMDKs(w http.ResponseWriter, r *http.Request) {
	if g.fixtures == nil || len(g.fixtures.Roots()) == 0 {
		http.Error(w, "no fixture directories are configured on this deployment", http.StatusBadRequest)
		return
	}
	rootIdx, err := strconv.Atoi(r.URL.Query().Get("root"))
	if err != nil {
		rootIdx = 0
	}
	target, err := g.resolveFixtureBrowsePath(rootIdx, r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dirEntries, err := os.ReadDir(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	entries := make([]browseEntry, 0, len(dirEntries))
	for _, e := range dirEntries {
		if e.IsDir() {
			entries = append(entries, browseEntry{Name: e.Name(), IsDir: true})
			continue
		}
		if !strings.EqualFold(filepath.Ext(e.Name()), ".vmdk") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		entries = append(entries, browseEntry{Name: e.Name(), SizeBytes: info.Size()})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})
	_ = json.NewEncoder(w).Encode(map[string]any{
		"root": rootIdx, "path": r.URL.Query().Get("path"), "entries": entries,
	})
}

// assignVMDK pins a file already staged under a configured fixture root
// (found via browseVMDKs) to a VM, without re-uploading it.
func (g *Gateway) assignVMDK(w http.ResponseWriter, r *http.Request) {
	if g.fixtures == nil {
		http.Error(w, "fixture directories are not configured on this deployment", http.StatusBadRequest)
		return
	}
	var body struct {
		Root     int    `json:"root"`
		FileName string `json:"file_name"`
		VMName   string `json:"vm_name"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.FileName) == "" || strings.TrimSpace(body.VMName) == "" {
		http.Error(w, "file_name and vm_name are required", http.StatusBadRequest)
		return
	}
	// Confirm the file is actually inside the claimed root before pinning it.
	if _, err := g.resolveFixtureBrowsePath(body.Root, body.FileName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := g.fixtures.SetOverride(body.Root, body.FileName, body.VMName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(g.vmdks())
}

// login checks a username/password against the dashboard's admin login
// credentials and, on success, hands back the same bearer token every other
// admin API call already expects — the wire-level auth scheme underneath is
// unchanged, this just adds a real front door in front of it.
func (g *Gateway) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<12)).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !g.checkLogin(body.Username, body.Password) {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"token": g.adminToken})
}

// changeCredentials replaces the live admin login username/password.
// Reachable only with a valid bearer token already, consistent with every
// other mutating admin action in this codebase — no separate re-entry of
// the current password is required.
func (g *Gateway) changeCredentials(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<12)).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Username) == "" || strings.TrimSpace(body.Password) == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}
	g.setCredentials(body.Username, body.Password)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func LogPanicServer(s *http.Server, lnErr <-chan error) {
	select {
	case err := <-lnErr:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("gateway: %v", err)
		}
	default:
	}
}
