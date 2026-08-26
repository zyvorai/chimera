package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
	g := New(u, "http://public.invalid", "", faults.New(), nil)
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
	s, err := fixture.New("", 1)
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
	g := New(u, "http://public.invalid", "", faults.New(), s)
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
	g := New(u, "http://public.invalid", "", faults.New(), nil)
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
	g := New(u, "http://chimera.test", "secret", faults.New(), nil, Meta{
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
