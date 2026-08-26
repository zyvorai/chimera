package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vmware/govmomi/simulator"

	"github.com/zyvorai/chimera/internal/faults"
	"github.com/zyvorai/chimera/internal/fixture"
)

func TestParseRangeStart(t *testing.T) {
	tests := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"bytes=0-", 0, true}, {"bytes=99-", 99, true}, {"bytes=20-30", 20, true}, {"items=1-", 0, false}, {"bytes=-10", 0, false},
	}
	for _, tt := range tests {
		got, err := parseRangeStart(tt.in)
		if tt.ok && err != nil {
			t.Fatalf("%s: %v", tt.in, err)
		}
		if !tt.ok && err == nil {
			t.Fatalf("%s: expected error", tt.in)
		}
		if tt.ok && got != tt.want {
			t.Fatalf("%s: got %d want %d", tt.in, got, tt.want)
		}
	}
}

func TestNFCRangeShim(t *testing.T) {
	payload := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nfc/lease/disk.vmdk" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(payload)
	}))
	defer up.Close()
	u, _ := url.Parse(up.URL)
	g := New(u, "http://public.invalid", "", faults.New(), nil, nil)
	pub := httptest.NewServer(g)
	defer pub.Close()

	req, _ := http.NewRequest(http.MethodGet, pub.URL+"/nfc/lease/disk.vmdk", nil)
	req.Header.Set("Range", "bytes=10-")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusPartialContent {
		t.Fatalf("status=%d", res.StatusCode)
	}
	got, _ := io.ReadAll(res.Body)
	if string(got) != string(payload[10:]) {
		t.Fatalf("got %q want %q", got, payload[10:])
	}
}

func TestLocalNFCRange(t *testing.T) {
	s, err := fixture.New("", "", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p, size, err := s.Prepare("vm0")
	if err != nil {
		t.Fatal(err)
	}
	s.Register("lease-x", "disk-0.vmdk", p)

	up := httptest.NewServer(http.NotFoundHandler())
	defer up.Close()
	u, _ := url.Parse(up.URL)
	g := New(u, "http://public.invalid", "", faults.New(), s, nil)
	pub := httptest.NewServer(g)
	defer pub.Close()

	start := int64(4096)
	req, _ := http.NewRequest(http.MethodGet, pub.URL+"/chimera-nfc/lease-x/disk-0.vmdk", nil)
	req.Header.Set("Range", "bytes=4096-")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusPartialContent {
		t.Fatalf("status=%d", res.StatusCode)
	}
	if got := res.Header.Get("Content-Range"); got != "bytes 4096-1048575/1048576" {
		t.Fatalf("content-range=%q", got)
	}
	n, err := io.Copy(io.Discard, res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if n != size-start {
		t.Fatalf("read=%d want=%d", n, size-start)
	}
}

func TestTelemetryTracksInfrastructureTraffic(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer up.Close()
	u, _ := url.Parse(up.URL)
	g := New(u, "http://public.invalid", "", faults.New(), nil, nil)
	pub := httptest.NewServer(g)
	defer pub.Close()

	res, err := http.Get(pub.URL + "/sdk")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, res.Body)
	_ = res.Body.Close()

	res, err = http.Get(pub.URL + "/__chimera/api/telemetry")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("telemetry status=%d", res.StatusCode)
	}
	got := string(body)
	if !strings.Contains(got, `"requests":1`) || !strings.Contains(got, `"path":"/sdk"`) || !strings.Contains(got, `"category":"vSphere SDK"`) {
		t.Fatalf("telemetry body=%s", got)
	}
}

func TestCommandCenterBootstrapAndAuth(t *testing.T) {
	up := httptest.NewServer(http.NotFoundHandler())
	defer up.Close()
	u, _ := url.Parse(up.URL)
	g := New(u, "http://chimera.test", "secret", faults.New(), nil, nil, Meta{
		Version: "test", Persona: "vSphere", Username: "administrator@vsphere.local",
		Datacenters: 1, Clusters: 1, Hosts: 2, Datastores: 1, VMs: 4, FixtureSizeMB: 16,
	})
	pub := httptest.NewServer(g)
	defer pub.Close()

	res, err := http.Get(pub.URL + "/__chimera/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), "Infrastructure Simulation Engine") {
		t.Fatalf("dashboard status=%d body=%q", res.StatusCode, string(body))
	}

	res, err = http.Get(pub.URL + "/__chimera/api/bootstrap")
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `"vms":4`) || !strings.Contains(string(body), `"id":"nutanix"`) {
		t.Fatalf("bootstrap status=%d body=%q", res.StatusCode, string(body))
	}

	res, err = http.Get(pub.URL + "/__chimera/state")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated state status=%d", res.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodGet, pub.URL+"/__chimera/state", nil)
	req.Header.Set("Authorization", "Bearer secret")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("authenticated state status=%d", res.StatusCode)
	}
}

// realVMModel builds a real govmomi simulator model (not the fake counts
// this package used to fabricate inventory from) and returns its registry
// plus the real VM names it created.
func realVMModel(t *testing.T) (*simulator.Registry, []string) {
	t.Helper()
	model := simulator.VPX()
	model.Machine = 2
	if err := model.Create(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(model.Remove)
	registry := model.Service.Context.Map
	var names []string
	for _, e := range registry.All("VirtualMachine") {
		if vm, ok := e.(*simulator.VirtualMachine); ok && vm.Name != "" {
			names = append(names, vm.Name)
		}
	}
	if len(names) == 0 {
		t.Fatal("model produced no VMs")
	}
	return registry, names
}

func TestInventoryReflectsRealSimulatorVMs(t *testing.T) {
	registry, vmNames := realVMModel(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, vmNames[0]+".vmdk"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtures, err := fixture.New("", dir, 1, vmNames)
	if err != nil {
		t.Fatal(err)
	}
	defer fixtures.Close()

	up := httptest.NewServer(http.NotFoundHandler())
	defer up.Close()
	u, _ := url.Parse(up.URL)
	g := New(u, "http://public.invalid", "", faults.New(), fixtures, registry)
	pub := httptest.NewServer(g)
	defer pub.Close()

	res, err := http.Get(pub.URL + "/__chimera/api/inventory")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var got struct {
		VirtualMachines []map[string]any `json:"virtual_machines"`
		Total           int              `json:"total"`
	}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Total != len(vmNames) {
		t.Fatalf("total=%d want=%d", got.Total, len(vmNames))
	}
	real := make(map[string]bool, len(vmNames))
	for _, n := range vmNames {
		real[n] = true
	}
	seenReal, matched := false, false
	for _, vm := range got.VirtualMachines {
		name, _ := vm["name"].(string)
		if !real[name] {
			t.Fatalf("inventory returned a name not in the real simulator model: %v", vm)
		}
		if name == vmNames[0] {
			seenReal = true
			if vm["fixture_source"] == "name-match" && vm["fixture_file"] == vmNames[0]+".vmdk" {
				matched = true
			}
		}
		switch vm["state"] {
		case "poweredOn", "poweredOff", "suspended":
		default:
			t.Fatalf("unexpected state: %v", vm["state"])
		}
	}
	if !seenReal {
		t.Fatalf("expected VM %q in inventory, got %v", vmNames[0], got.VirtualMachines)
	}
	if !matched {
		t.Fatalf("expected VM %q to have fixture_source=name-match", vmNames[0])
	}
}

func TestVMDKsEndpoint(t *testing.T) {
	registry, vmNames := realVMModel(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, vmNames[0]+".vmdk"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "unmatched-disk.vmdk"), []byte("y"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtures, err := fixture.New("", dir, 1, vmNames)
	if err != nil {
		t.Fatal(err)
	}
	defer fixtures.Close()

	up := httptest.NewServer(http.NotFoundHandler())
	defer up.Close()
	u, _ := url.Parse(up.URL)
	g := New(u, "http://public.invalid", "", faults.New(), fixtures, registry)
	pub := httptest.NewServer(g)
	defer pub.Close()

	res, err := http.Get(pub.URL + "/__chimera/api/vmdks")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var got struct {
		Total      int                  `json:"total"`
		Matched    int                  `json:"matched"`
		RoundRobin int                  `json:"round_robin"`
		Unassigned int                  `json:"unassigned"`
		Files      []fixture.Assignment `json:"files"`
	}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	// 2 VMs, 2 files: one matches by name, the leftover file/VM pair off via round-robin.
	if got.Total != 2 || got.Matched != 1 || got.RoundRobin != 1 || got.Unassigned != 0 {
		t.Fatalf("got total=%d matched=%d round_robin=%d unassigned=%d, want 2/1/1/0", got.Total, got.Matched, got.RoundRobin, got.Unassigned)
	}
}
