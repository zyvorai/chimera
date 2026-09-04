// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package azure

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zyvorai/chimera/internal/personas/common"
)

const defaultResourceGroup = "chimera-rg"

type operation struct {
	Task      common.Task
	AccessSAS string
}

type Server struct {
	SubscriptionID string
	BearerToken    string
	Region         string
	Store          *common.Store

	mu         sync.RWMutex
	operations map[string]operation
}

func New(subscriptionID, bearerToken string, vmCount int) *Server {
	if strings.TrimSpace(subscriptionID) == "" {
		subscriptionID = "00000000-0000-0000-0000-000000000000"
	}
	return &Server{
		SubscriptionID: subscriptionID,
		BearerToken:    bearerToken,
		Region:         "eastus",
		Store:          common.NewStore(vmCount),
		operations:     map[string]operation{},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/__chimera/azure/disks/") {
		s.serveDiskData(w, r)
		return
	}
	if err := s.auth(r); err != nil {
		writeARMError(w, http.StatusUnauthorized, "AuthenticationFailed", err.Error())
		return
	}
	if strings.HasPrefix(r.URL.Path, "/__chimera/azure/operations/") {
		s.serveOperation(w, r)
		return
	}

	parts := splitPath(r.URL.Path)
	if len(parts) < 2 || !strings.EqualFold(parts[0], "subscriptions") || parts[1] != s.SubscriptionID {
		writeARMError(w, http.StatusNotFound, "SubscriptionNotFound", "unknown subscription")
		return
	}

	if isSubscriptionVMList(parts) && r.Method == http.MethodGet {
		s.listVMs(w, "")
		return
	}
	if rg, ok := resourceGroup(parts); ok {
		if isResourceGroupVMList(parts) && r.Method == http.MethodGet {
			s.listVMs(w, rg)
			return
		}
		if vmName, suffix, ok := vmPath(parts); ok {
			s.serveVM(w, r, rg, vmName, suffix)
			return
		}
		if diskName, suffix, ok := diskPath(parts); ok {
			s.serveDisk(w, r, rg, diskName, suffix)
			return
		}
	}
	writeARMError(w, http.StatusNotFound, "ResourceNotFound", "resource not found")
}

func (s *Server) listVMs(w http.ResponseWriter, rg string) {
	vms := s.Store.List()
	values := make([]any, 0, len(vms))
	for _, vm := range vms {
		values = append(values, s.vmResource(vm, chooseRG(rg)))
	}
	writeJSON(w, http.StatusOK, map[string]any{"value": values})
}

func (s *Server) serveVM(w http.ResponseWriter, r *http.Request, rg, vmName, suffix string) {
	vm, ok := s.findVMByName(vmName)
	if !ok {
		writeARMError(w, http.StatusNotFound, "ResourceNotFound", "virtual machine not found")
		return
	}
	switch {
	case suffix == "" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, s.vmResource(vm, rg))
	case suffix == "instanceView" && r.Method == http.MethodGet:
		power := "PowerState/deallocated"
		if strings.EqualFold(vm.Power, "ON") {
			power = "PowerState/running"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"computerName": vm.Name,
			"statuses": []map[string]any{
				{"code": "ProvisioningState/succeeded", "level": "Info", "displayStatus": "Provisioning succeeded"},
				{"code": power, "level": "Info", "displayStatus": strings.TrimPrefix(power, "PowerState/")},
			},
		})
	case (suffix == "start" || suffix == "powerOff" || suffix == "deallocate" || suffix == "restart") && r.Method == http.MethodPost:
		state := "ON"
		if suffix == "powerOff" || suffix == "deallocate" {
			state = "OFF"
		}
		t, _ := s.Store.SetPower(vm.ID, state)
		if suffix == "restart" {
			t.Operation = "restart"
		}
		s.saveOperation(t, "")
		opURL := "/__chimera/azure/operations/" + t.ID
		w.Header().Set("Azure-AsyncOperation", opURL)
		w.Header().Set("Location", opURL)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusAccepted)
	default:
		writeARMError(w, http.StatusNotFound, "ResourceNotFound", "unsupported VM operation")
	}
}

func (s *Server) serveDisk(w http.ResponseWriter, r *http.Request, rg, diskName, suffix string) {
	vm, disk, ok := s.findDiskByName(diskName)
	if !ok {
		writeARMError(w, http.StatusNotFound, "ResourceNotFound", "managed disk not found")
		return
	}
	if suffix == "" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, s.diskResource(vm, disk, rg))
		return
	}
	if suffix == "beginGetAccess" && r.Method == http.MethodPost {
		var req struct {
			Access            string `json:"access"`
			DurationInSeconds int    `json:"durationInSeconds"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Access != "" && !strings.EqualFold(req.Access, "Read") {
			writeARMError(w, http.StatusBadRequest, "InvalidParameter", "only Read access is supported")
			return
		}
		if req.DurationInSeconds <= 0 {
			req.DurationInSeconds = 3600
		}
		t := s.Store.NewTask("begin_get_access", disk.ID)
		expiry := time.Now().UTC().Add(time.Duration(req.DurationInSeconds) * time.Second).Unix()
		sig := sasSignature(disk.ID, expiry)
		access := fmt.Sprintf("/__chimera/azure/disks/%s/download?sp=r&se=%d&sig=%s", disk.ID, expiry, sig)
		s.saveOperation(t, access)
		opURL := "/__chimera/azure/operations/" + t.ID
		w.Header().Set("Azure-AsyncOperation", opURL)
		w.Header().Set("Location", opURL)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusAccepted)
		return
	}
	writeARMError(w, http.StatusNotFound, "ResourceNotFound", "unsupported disk operation")
}

func (s *Server) serveOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := path.Base(r.URL.Path)
	s.mu.RLock()
	op, ok := s.operations[id]
	s.mu.RUnlock()
	if !ok {
		writeARMError(w, http.StatusNotFound, "OperationNotFound", "operation not found")
		return
	}
	out := map[string]any{"status": "Succeeded", "name": op.Task.ID}
	if op.AccessSAS != "" {
		out["properties"] = map[string]any{"output": map[string]any{"accessSAS": absoluteURL(r, op.AccessSAS)}}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) serveDiskData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parts := splitPath(r.URL.Path)
	if len(parts) != 5 || parts[0] != "__chimera" || parts[1] != "azure" || parts[2] != "disks" || parts[4] != "download" {
		http.NotFound(w, r)
		return
	}
	diskID := parts[3]
	if r.URL.Query().Get("sp") != "r" {
		http.Error(w, "invalid SAS permission", http.StatusForbidden)
		return
	}
	expiry, err := strconv.ParseInt(r.URL.Query().Get("se"), 10, 64)
	if err != nil || time.Now().UTC().Unix() > expiry {
		http.Error(w, "expired SAS", http.StatusForbidden)
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("sig")), []byte(sasSignature(diskID, expiry))) != 1 {
		http.Error(w, "invalid SAS", http.StatusForbidden)
		return
	}
	vm, disk, ok := s.findDiskByID(diskID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	b := common.DiskBytes(vm.ID, disk.ID, disk.SizeBytes)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("x-ms-blob-type", "PageBlob")
	w.Header().Set("x-ms-version", "2023-11-03")
	http.ServeContent(w, r, disk.Name+".vhd", time.Unix(0, 0), bytes.NewReader(b))
}

func (s *Server) vmResource(vm common.VM, rg string) map[string]any {
	disks := make([]any, 0, len(vm.Disks))
	for i, d := range vm.Disks {
		entry := map[string]any{
			"name":         diskName(vm, d, i),
			"createOption": chooseCreateOption(i),
			"managedDisk":  map[string]any{"id": diskResourceID(s.SubscriptionID, rg, diskName(vm, d, i)), "storageAccountType": "Premium_LRS"},
			"diskSizeGB":   gib(d.SizeBytes),
		}
		if i == 0 {
			entry["osType"] = "Linux"
		}
		disks = append(disks, entry)
	}
	storage := map[string]any{}
	if len(disks) > 0 {
		storage["osDisk"] = disks[0]
		if len(disks) > 1 {
			storage["dataDisks"] = disks[1:]
		} else {
			storage["dataDisks"] = []any{}
		}
	}
	return map[string]any{
		"id":       vmResourceID(s.SubscriptionID, rg, vm.Name),
		"name":     vm.Name,
		"type":     "Microsoft.Compute/virtualMachines",
		"location": s.Region,
		"tags":     vm.Labels,
		"properties": map[string]any{
			"hardwareProfile": map[string]any{"vmSize": "Standard_D2s_v5"},
			"storageProfile":  storage,
			"osProfile":       map[string]any{"computerName": vm.Name, "adminUsername": "chimera"},
			"networkProfile": map[string]any{"networkInterfaces": []any{
				map[string]any{"id": fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/networkInterfaces/%s-nic0", s.SubscriptionID, rg, vm.Name), "properties": map[string]any{"primary": true}},
			}},
			"provisioningState": "Succeeded",
		},
	}
}

func (s *Server) diskResource(vm common.VM, disk common.Disk, rg string) map[string]any {
	name := ""
	for i, d := range vm.Disks {
		if d.ID == disk.ID {
			name = diskName(vm, d, i)
			break
		}
	}
	return map[string]any{
		"id":       diskResourceID(s.SubscriptionID, rg, name),
		"name":     name,
		"type":     "Microsoft.Compute/disks",
		"location": s.Region,
		"sku":      map[string]any{"name": "Premium_LRS", "tier": "Premium"},
		"properties": map[string]any{
			"diskSizeGB":        gib(disk.SizeBytes),
			"diskState":         "Attached",
			"managedBy":         vmResourceID(s.SubscriptionID, rg, vm.Name),
			"provisioningState": "Succeeded",
			"creationData":      map[string]any{"createOption": "FromImage"},
		},
	}
}

func (s *Server) findVMByName(name string) (common.VM, bool) {
	for _, vm := range s.Store.List() {
		if vm.Name == name {
			return vm, true
		}
	}
	return common.VM{}, false
}

func (s *Server) findDiskByName(name string) (common.VM, common.Disk, bool) {
	for _, vm := range s.Store.List() {
		for i, d := range vm.Disks {
			if diskName(vm, d, i) == name {
				return vm, d, true
			}
		}
	}
	return common.VM{}, common.Disk{}, false
}

func (s *Server) findDiskByID(id string) (common.VM, common.Disk, bool) {
	for _, vm := range s.Store.List() {
		for _, d := range vm.Disks {
			if d.ID == id {
				return vm, d, true
			}
		}
	}
	return common.VM{}, common.Disk{}, false
}

func (s *Server) saveOperation(t common.Task, access string) {
	s.mu.Lock()
	s.operations[t.ID] = operation{Task: t, AccessSAS: access}
	s.mu.Unlock()
}

func (s *Server) auth(r *http.Request) error {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return fmt.Errorf("missing bearer token")
	}
	got := strings.TrimPrefix(h, "Bearer ")
	if subtle.ConstantTimeCompare([]byte(got), []byte(s.BearerToken)) != 1 {
		return fmt.Errorf("invalid bearer token")
	}
	return nil
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func resourceGroup(parts []string) (string, bool) {
	for i := 2; i+1 < len(parts); i++ {
		if strings.EqualFold(parts[i], "resourceGroups") {
			return parts[i+1], true
		}
	}
	return "", false
}

func isSubscriptionVMList(parts []string) bool {
	return len(parts) == 5 && strings.EqualFold(parts[2], "providers") && strings.EqualFold(parts[3], "Microsoft.Compute") && strings.EqualFold(parts[4], "virtualMachines")
}

func isResourceGroupVMList(parts []string) bool {
	return len(parts) == 7 && strings.EqualFold(parts[2], "resourceGroups") && strings.EqualFold(parts[4], "providers") && strings.EqualFold(parts[5], "Microsoft.Compute") && strings.EqualFold(parts[6], "virtualMachines")
}

func vmPath(parts []string) (name, suffix string, ok bool) {
	if len(parts) < 8 || !strings.EqualFold(parts[4], "providers") || !strings.EqualFold(parts[5], "Microsoft.Compute") || !strings.EqualFold(parts[6], "virtualMachines") {
		return "", "", false
	}
	name = parts[7]
	if len(parts) > 8 {
		suffix = strings.Join(parts[8:], "/")
	}
	return name, suffix, true
}

func diskPath(parts []string) (name, suffix string, ok bool) {
	if len(parts) < 8 || !strings.EqualFold(parts[4], "providers") || !strings.EqualFold(parts[5], "Microsoft.Compute") || !strings.EqualFold(parts[6], "disks") {
		return "", "", false
	}
	name = parts[7]
	if len(parts) > 8 {
		suffix = strings.Join(parts[8:], "/")
	}
	return name, suffix, true
}

func vmResourceID(sub, rg, name string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", sub, rg, name)
}

func diskResourceID(sub, rg, name string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/disks/%s", sub, rg, name)
}

func diskName(vm common.VM, disk common.Disk, idx int) string {
	if idx == 0 {
		return vm.Name + "-osdisk"
	}
	return fmt.Sprintf("%s-datadisk-%d", vm.Name, idx)
}

func chooseCreateOption(i int) string {
	if i == 0 {
		return "FromImage"
	}
	return "Empty"
}

func chooseRG(rg string) string {
	if rg == "" {
		return defaultResourceGroup
	}
	return rg
}

func gib(bytes int64) int {
	g := int((bytes + (1 << 30) - 1) >> 30)
	if g < 1 {
		return 1
	}
	return g
}

func sasSignature(diskID string, expiry int64) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:chimera-azure-sas", diskID, expiry)))
	return hex.EncodeToString(h[:16])
}

func absoluteURL(r *http.Request, rel string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + rel
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeARMError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": code, "message": message}})
}

// Keep deterministic ordering in any future ARM collection extensions.
func sortedKeys[M ~map[string]V, V any](m M) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
