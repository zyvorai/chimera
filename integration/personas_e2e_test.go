// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zyvorai/chimera/internal/personas/hyperv"
	"github.com/zyvorai/chimera/internal/personas/nutanix"
)

func TestNutanixDiscoveryPowerAndDiskExportE2E(t *testing.T) {
	ts := httptest.NewServer(nutanix.New("admin", "secret", 2))
	defer ts.Close()
	c := ts.Client()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/nutanix/v3/vms/list", strings.NewReader(`{"kind":"vm"}`))
	req.SetBasicAuth("admin", "secret")
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("list status %d", res.StatusCode)
	}
	var list struct {
		Entities []struct {
			Metadata struct {
				UUID string `json:"uuid"`
			} `json:"metadata"`
			Spec struct {
				Name      string `json:"name"`
				Resources struct {
					DiskList []struct {
						UUID string `json:"uuid"`
					} `json:"disk_list"`
				} `json:"resources"`
			} `json:"spec"`
		} `json:"entities"`
	}
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Entities) != 2 {
		t.Fatalf("want 2 VMs, got %d", len(list.Entities))
	}
	vmID := list.Entities[0].Metadata.UUID
	diskID := list.Entities[0].Spec.Resources.DiskList[0].UUID
	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/api/nutanix/v3/vms/"+vmID+"/set_power_state", bytes.NewBufferString(`{"transition":"ON"}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "secret")
	res, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("power status %d", res.StatusCode)
	}
	res.Body.Close()
	req, _ = http.NewRequest(http.MethodGet, ts.URL+"/api/nutanix/v3/vms/"+vmID+"/disks/"+diskID+"/data", nil)
	req.SetBasicAuth("admin", "secret")
	res, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, 4096))
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 4096 {
		t.Fatalf("want 4096 exported bytes, got %d", len(b))
	}
}

func TestHyperVWSManIdentifyEnumeratePullAndPowerE2E(t *testing.T) {
	ts := httptest.NewServer(hyperv.New("Administrator", "secret", 2))
	defer ts.Close()
	c := ts.Client()
	call := func(body string) string {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/wsman", strings.NewReader(body))
		req.SetBasicAuth("Administrator", "secret")
		req.Header.Set("Content-Type", "application/soap+xml;charset=UTF-8")
		res, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		if res.StatusCode != 200 {
			t.Fatalf("status=%d body=%s", res.StatusCode, b)
		}
		return string(b)
	}
	identify := call(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Header><a:Action xmlns:a="http://www.w3.org/2005/08/addressing">http://schemas.dmtf.org/wbem/wsman/identity/1/wsmanidentity/Identify</a:Action></s:Header><s:Body/></s:Envelope>`)
	if !strings.Contains(identify, "Zyvor Chimera") {
		t.Fatalf("bad identify: %s", identify)
	}
	enum := call(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Header><a:Action xmlns:a="http://www.w3.org/2005/08/addressing">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate</a:Action></s:Header><s:Body/></s:Envelope>`)
	if !strings.Contains(enum, "chimera-vms") {
		t.Fatal("missing enumeration context")
	}
	pull := call(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Header><a:Action xmlns:a="http://www.w3.org/2005/08/addressing">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Pull</a:Action></s:Header><s:Body/></s:Envelope>`)
	if !strings.Contains(pull, "chimera-vm-01") || !strings.Contains(pull, "vm-0001") {
		t.Fatalf("missing VM inventory: %s", pull)
	}
	power := call(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Header><a:Action xmlns:a="http://www.w3.org/2005/08/addressing">RequestStateChange</a:Action><w:SelectorSet xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd"><w:Selector Name="Name">vm-0001</w:Selector></w:SelectorSet></s:Header><s:Body><p:RequestStateChange_INPUT xmlns:p="urn:test"><p:RequestedState>2</p:RequestedState></p:RequestStateChange_INPUT></s:Body></s:Envelope>`)
	if !strings.Contains(power, "ReturnValue>4096") {
		t.Fatalf("bad power response: %s", power)
	}
}
