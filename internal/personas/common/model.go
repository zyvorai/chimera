// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type Disk struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SizeBytes int64  `json:"size_bytes"`
	Format    string `json:"format"`
}

type VM struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	CPUs     int               `json:"cpus"`
	MemoryMB int               `json:"memory_mb"`
	Power    string            `json:"power"`
	Disks    []Disk            `json:"disks"`
	NICs     []string          `json:"nics"`
	Labels   map[string]string `json:"labels,omitempty"`
}

type Task struct {
	ID        string    `json:"id"`
	State     string    `json:"state"`
	Operation string    `json:"operation"`
	VMID      string    `json:"vm_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	mu    sync.RWMutex
	VMs   map[string]*VM
	Tasks map[string]*Task
}

func NewStore(count int) *Store {
	if count < 1 {
		count = 3
	}
	s := &Store{VMs: map[string]*VM{}, Tasks: map[string]*Task{}}
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("vm-%04d", i+1)
		diskID := fmt.Sprintf("disk-%04d", i+1)
		s.VMs[id] = &VM{
			ID: id, Name: fmt.Sprintf("chimera-vm-%02d", i+1), CPUs: 2, MemoryMB: 4096,
			Power: "OFF", Disks: []Disk{{ID: diskID, Name: "os", SizeBytes: 16 << 20, Format: "raw"}},
			NICs: []string{"default"}, Labels: map[string]string{"chimera": "true"},
		}
	}
	return s
}

func (s *Store) List() []VM {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]VM, 0, len(s.VMs))
	for _, vm := range s.VMs {
		out = append(out, *vm)
	}
	return out
}

func (s *Store) Get(id string) (VM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vm, ok := s.VMs[id]
	if !ok {
		return VM{}, false
	}
	return *vm, true
}

func (s *Store) SetPower(id, state string) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vm, ok := s.VMs[id]
	if !ok {
		return Task{}, false
	}
	vm.Power = state
	seed := sha256.Sum256([]byte(id + state + time.Now().UTC().Format(time.RFC3339Nano)))
	task := Task{ID: "task-" + hex.EncodeToString(seed[:6]), State: "SUCCEEDED", Operation: "set_power", VMID: id, CreatedAt: time.Now().UTC()}
	s.Tasks[task.ID] = &task
	return task, true
}

func (s *Store) Task(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.Tasks[id]
	if !ok {
		return Task{}, false
	}
	return *t, true
}

func DiskBytes(vmID, diskID string, size int64) []byte {
	if size < 4096 {
		size = 4096
	}
	if size > 16<<20 {
		size = 16 << 20
	}
	seed := sha256.Sum256([]byte(vmID + ":" + diskID))
	out := make([]byte, size)
	for i := range out {
		out[i] = seed[i%len(seed)]
	}
	return out
}
