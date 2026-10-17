// Copyright 2026 Zyvor AI Labs · https://zyvor.dev
// SPDX-License-Identifier: Apache-2.0

package nutanix

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/zyvorai/chimera/internal/personas/common"
)

// Prism Central v4 surface (vmm + clustermgmt) as used by the official
// ntnx-api-golang-clients SDK. Any minor version (v4.0, v4.1, v4.3, ...) is
// accepted so SDK upgrades do not break the persona.
const (
	ClusterExtID   = "00000000-0000-4000-8000-00000000c1a5"
	ContainerExtID = "00000000-0000-4000-8000-0000000c0001"
	ContainerName  = "default-container"
)

var (
	v4VMs        = regexp.MustCompile(`^/api/vmm/v4\.\d+/ahv/config/vms(?:/([^/]+))?$`)
	v4Info       = regexp.MustCompile(`^/api/(?:vmm|clustermgmt)/unversioned/info$`)
	v4Containers = regexp.MustCompile(`^/api/clustermgmt/v4\.\d+/config/storage-containers$`)
)

// SetNFSRoot makes the persona materialise each VM disk as a file below
// <root>/<container>/.acropolis/vmdisk/<diskExtId>/, the layout a Nutanix
// storage container has when mounted over NFS (what Transiva's pickup reads).
func (s *Server) SetNFSRoot(root string) error {
	s.NFSRoot = root
	if root == "" {
		return nil
	}
	for _, vm := range s.Store.List() {
		for _, d := range vm.Disks {
			dir := filepath.Join(root, ContainerName, ".acropolis", "vmdisk", v4DiskID(d.ID))
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, "disk.raw"), common.DiskBytes(vm.ID, d.ID, d.SizeBytes), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) serveV4(w http.ResponseWriter, r *http.Request, p string) bool {
	// The SDK negotiates its API version with an OPTIONS call before anything else.
	if v4Info.MatchString(p) {
		writeJSON(w, map[string]any{"data": "v4.3"})
		return true
	}
	if m := v4VMs.FindStringSubmatch(p); m != nil {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return true
		}
		if m[1] == "" {
			s.v4ListVMs(w, r)
		} else {
			s.v4GetVM(w, r, m[1])
		}
		return true
	}
	if v4Containers.MatchString(p) && r.Method == http.MethodGet {
		writeJSON(w, map[string]any{
			"$objectType":            "clustermgmt.v4.config.ListStorageContainersApiResponse",
			"$dataItemDiscriminator": "clustermgmt.v4.config.StorageContainer",
			"data": []any{map[string]any{
				"$objectType":    "clustermgmt.v4.config.StorageContainer",
				"extId":          ContainerExtID,
				"containerExtId": ContainerExtID,
				"name":           ContainerName,
				"clusterExtId":   ClusterExtID,
			}},
			"metadata": map[string]any{"totalAvailableResults": 1},
		})
		return true
	}
	return false
}

func (s *Server) v4ListVMs(w http.ResponseWriter, r *http.Request) {
	vms := s.Store.List()
	page, _ := strconv.Atoi(r.URL.Query().Get("$page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("$limit"))
	if limit <= 0 {
		limit = 50
	}
	total := len(vms)
	lo := min(page*limit, total)
	hi := min(lo+limit, total)
	data := make([]any, 0, hi-lo)
	for _, vm := range vms[lo:hi] {
		data = append(data, v4VM(vm))
	}
	writeJSON(w, map[string]any{
		"$objectType":            "vmm.v4.ahv.config.ListVmsApiResponse",
		"$dataItemDiscriminator": "vmm.v4.ahv.config.Vm",
		"data":                   data,
		"metadata":               map[string]any{"totalAvailableResults": total},
	})
}

func (s *Server) v4GetVM(w http.ResponseWriter, r *http.Request, id string) {
	for _, vm := range s.Store.List() {
		if v4VMID(vm.ID) == id || vm.ID == id {
			writeJSON(w, map[string]any{
				"$objectType":            "vmm.v4.ahv.config.GetVmApiResponse",
				"$dataItemDiscriminator": "vmm.v4.ahv.config.Vm",
				"data":                   v4VM(vm),
				"metadata":               map[string]any{},
			})
			return
		}
	}
	http.NotFound(w, r)
}

// v4VMID / v4DiskID derive stable UUID-shaped extIds from the store ids.
func v4VMID(id string) string   { return uuidFrom("1", id) }
func v4DiskID(id string) string { return uuidFrom("2", id) }

func uuidFrom(kind, id string) string {
	h := make([]byte, 0, 12)
	for _, c := range []byte(id) {
		if c >= '0' && c <= '9' {
			h = append(h, c)
		}
	}
	for len(h) < 12 {
		h = append([]byte("0"), h...)
	}
	return "00000000-0000-4000-800" + kind + "-" + string(h[len(h)-12:])
}

func v4VM(vm common.VM) map[string]any {
	disks := make([]any, 0, len(vm.Disks))
	for i, d := range vm.Disks {
		disks = append(disks, map[string]any{
			"$objectType": "vmm.v4.ahv.config.Disk",
			"extId":       v4DiskID(d.ID),
			"diskAddress": map[string]any{"$objectType": "vmm.v4.ahv.config.DiskAddress", "busType": "SCSI", "index": i},
			"backingInfo": map[string]any{
				"$objectType":      "vmm.v4.ahv.config.VmDisk",
				"diskExtId":        v4DiskID(d.ID),
				"diskSizeBytes":    d.SizeBytes,
				"storageContainer": map[string]any{"$objectType": "vmm.v4.ahv.config.VmDiskContainerReference", "extId": ContainerExtID},
			},
		})
	}
	nics := make([]any, 0, len(vm.NICs))
	for range vm.NICs {
		nics = append(nics, map[string]any{"$objectType": "vmm.v4.ahv.config.Nic"})
	}
	return map[string]any{
		"$objectType":       "vmm.v4.ahv.config.Vm",
		"extId":             v4VMID(vm.ID),
		"name":              vm.Name,
		"powerState":        strings.ToUpper(vm.Power),
		"numSockets":        1,
		"numCoresPerSocket": vm.CPUs,
		"numThreadsPerCore": 1,
		"memorySizeBytes":   int64(vm.MemoryMB) << 20,
		"cluster":           map[string]any{"$objectType": "vmm.v4.ahv.config.ClusterReference", "extId": ClusterExtID},
		"disks":             disks,
		"nics":              nics,
	}
}
