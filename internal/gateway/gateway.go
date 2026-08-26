package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/zyvorai/chimera/internal/faults"
	"github.com/zyvorai/chimera/internal/fixture"
)

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
}

type Gateway struct {
	backend    *url.URL
	publicBase string
	proxy      *httputil.ReverseProxy
	faults     *faults.State
	fixtures   *fixture.Store
	adminToken string
	started    time.Time
	meta       Meta
	telemetry  *telemetry
}

func New(backend *url.URL, publicBase, adminToken string, fs *faults.State, fixtures *fixture.Store, meta ...Meta) *Gateway {
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
	return &Gateway{backend: backend, publicBase: publicBase, proxy: p, faults: fs, fixtures: fixtures, adminToken: adminToken, started: time.Now(), meta: m, telemetry: newTelemetry()}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		r.URL.Path == "/__chimera/api/inventory" || r.URL.Path == "/__chimera/api/telemetry"
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
		{"id": "nutanix", "name": "Nutanix Prism", "short": "Nutanix", "status": "planned", "protocol": "Prism v3/v4 API", "accent": "cyan"},
		{"id": "proxmox", "name": "Proxmox VE", "short": "Proxmox", "status": "planned", "protocol": "REST API", "accent": "orange"},
		{"id": "openstack", "name": "OpenStack", "short": "OpenStack", "status": "planned", "protocol": "Nova / Glance", "accent": "rose"},
		{"id": "hyperv", "name": "Microsoft Hyper-V", "short": "Hyper-V", "status": "planned", "protocol": "WinRM / WMI", "accent": "blue"},
		{"id": "cloud", "name": "Cloud APIs", "short": "Cloud", "status": "planned", "protocol": "AWS / Azure style", "accent": "green"},
	}
	return map[string]any{
		"product": "Chimera", "tagline": "One engine. Many infrastructure personalities.", "version": g.meta.Version,
		"persona": g.meta.Persona, "username": g.meta.Username, "endpoint": g.publicBase + "/sdk", "public_base": g.publicBase,
		"tls": g.meta.TLS, "datacenters": g.meta.Datacenters, "clusters": g.meta.Clusters, "hosts": g.meta.Hosts,
		"datastores": g.meta.Datastores, "vms": g.meta.VMs, "fixture_size_mb": g.meta.FixtureSizeMB, "providers": providers,
	}
}

func (g *Gateway) inventory() map[string]any {
	vms := make([]map[string]any, 0, g.meta.VMs)
	for i := 0; i < g.meta.VMs; i++ {
		state := "poweredOff"
		if i%3 == 0 {
			state = "poweredOn"
		}
		vms = append(vms, map[string]any{
			"id": fmt.Sprintf("vm-%03d", i+1), "name": fmt.Sprintf("DC0_C0_RP0_VM%d", i), "state": state,
			"cpu": 2 + (i%4)*2, "memory_gb": 4 + (i%4)*4, "disk_gb": 20 + i*10,
			"datastore": fmt.Sprintf("LocalDS_%d", i%maxInt(1, g.meta.Datastores)), "network": "VM Network",
			"exportable": true,
		})
	}
	return map[string]any{"persona": g.meta.Persona, "virtual_machines": vms, "total": len(vms)}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
