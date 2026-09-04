// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package nutanix

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/zyvorai/chimera/internal/personas/common"
)

type Server struct {
	Username string
	Password string
	Store    *common.Store
}

func New(username, password string, vmCount int) *Server {
	return &Server{Username: username, Password: password, Store: common.NewStore(vmCount)}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.auth(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Prism"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	p := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case r.Method == http.MethodGet && p == "/api/nutanix/v3/cluster":
		writeJSON(w, map[string]any{"metadata": map[string]any{"kind": "cluster"}, "status": map[string]any{"name": "chimera-prism", "resources": map[string]any{"config": map[string]any{"service_list": []string{"AOS", "AHV"}}}}})
	case r.Method == http.MethodPost && p == "/api/nutanix/v3/vms/list":
		vms := s.Store.List()
		entities := make([]any, 0, len(vms))
		for _, vm := range vms {
			entities = append(entities, prismVM(vm))
		}
		writeJSON(w, map[string]any{"api_version": "3.1", "metadata": map[string]any{"total_matches": len(entities)}, "entities": entities})
	case strings.HasPrefix(p, "/api/nutanix/v3/vms/"):
		id := strings.TrimPrefix(p, "/api/nutanix/v3/vms/")
		if strings.Contains(id, "/") {
			s.vmSubresource(w, r, id)
			return
		}
		vm, ok := s.Store.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, prismVM(vm))
	case r.Method == http.MethodGet && strings.HasPrefix(p, "/api/nutanix/v3/tasks/"):
		id := strings.TrimPrefix(p, "/api/nutanix/v3/tasks/")
		t, ok := s.Store.Task(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]any{"api_version": "3.1", "metadata": map[string]any{"uuid": t.ID}, "status": map[string]any{"state": t.State, "operation": t.Operation}})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) vmSubresource(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.Split(rest, "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	id, action := parts[0], strings.Join(parts[1:], "/")
	switch {
	case r.Method == http.MethodPost && action == "set_power_state":
		var body struct {
			Transition string `json:"transition"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		state := "ON"
		if strings.EqualFold(body.Transition, "OFF") || strings.EqualFold(body.Transition, "ACPI_SHUTDOWN") {
			state = "OFF"
		}
		t, ok := s.Store.SetPower(id, state)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]any{"status": map[string]any{"execution_context": map[string]any{"task_uuid": t.ID}, "state": "PENDING"}})
	case r.Method == http.MethodGet && strings.HasPrefix(action, "disks/") && strings.HasSuffix(action, "/data"):
		diskID := strings.TrimSuffix(strings.TrimPrefix(action, "disks/"), "/data")
		vm, ok := s.Store.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		found := false
		var size int64
		for _, d := range vm.Disks {
			if d.ID == diskID {
				found = true
				size = d.SizeBytes
				break
			}
		}
		if !found {
			http.NotFound(w, r)
			return
		}
		b := common.DiskBytes(id, diskID, size)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprint(len(b)))
		w.Header().Set("Accept-Ranges", "bytes")
		_, _ = w.Write(b)
	default:
		http.NotFound(w, r)
	}
}

func prismVM(vm common.VM) map[string]any {
	disks := make([]any, 0, len(vm.Disks))
	for _, d := range vm.Disks {
		disks = append(disks, map[string]any{"uuid": d.ID, "disk_size_bytes": d.SizeBytes, "device_properties": map[string]any{"device_type": "DISK"}})
	}
	return map[string]any{"api_version": "3.1", "metadata": map[string]any{"kind": "vm", "uuid": vm.ID}, "spec": map[string]any{"name": vm.Name, "resources": map[string]any{"num_sockets": 1, "num_vcpus_per_socket": vm.CPUs, "memory_size_mib": vm.MemoryMB, "disk_list": disks}}, "status": map[string]any{"state": "COMPLETE", "resources": map[string]any{"power_state": vm.Power}}}
}

func (s *Server) auth(r *http.Request) bool {
	u, p, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(u), []byte(s.Username)) == 1 && subtle.ConstantTimeCompare([]byte(p), []byte(s.Password)) == 1
}
func writeJSON(w http.ResponseWriter, v any) { _ = json.NewEncoder(w).Encode(v) }
