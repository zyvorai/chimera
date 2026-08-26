package fixture

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultAlias = "disk-0.vmdk"

// fixtureDirPollInterval is how often fixture_vmdk_dir is re-scanned for
// added/removed/renamed files, so an operator can drop in a new VMDK without
// restarting the server.
const fixtureDirPollInterval = 5 * time.Second

// Fixture source methods, reported via Assignments/SourceFor.
const (
	MethodNameMatch  = "name-match"  // fixture_vmdk_dir file whose sanitized basename matched a VM's sanitized name
	MethodRoundRobin = "round-robin" // fixture_vmdk_dir leftover file paired with a leftover VM in sorted order
	MethodSharedFile = "shared-file" // single fixture_vmdk shared by every VM
	MethodGenerated  = "generated"   // per-VM deterministic synthetic fixture
	MethodUnassigned = "unassigned"  // fixture_vmdk_dir file that matched no VM
)

// Assignment describes one file found under fixture_vmdk_dir and, if any, the
// VM it was matched to.
type Assignment struct {
	FileName  string `json:"file_name"`
	Path      string `json:"-"`
	SizeBytes int64  `json:"size_bytes"`
	VMName    string `json:"vm_name"`
	Method    string `json:"method"`
}

type Store struct {
	mu          sync.RWMutex
	dir         string
	fixturePath string
	fixtureDir  string
	vmNames     []string
	sizeMB      int
	files       map[string]string

	// byVM and assignments are re-built by Rescan (initially, then on a
	// timer if fixtureDir is set) — reads/writes go through mu.
	byVM        map[string]Assignment
	assignments []Assignment

	stopPoll chan struct{}
}

// New constructs a fixture Store. fixturePath (a single shared VMDK) and
// fixtureDir (a directory of VMDKs, one matched per VM) are mutually
// exclusive. vmNames is the real, live list of simulated VM names, used to
// match against fixtureDir's contents.
func New(fixturePath, fixtureDir string, sizeMB int, vmNames []string) (*Store, error) {
	if sizeMB < 1 {
		sizeMB = 16
	}
	if fixturePath != "" && fixtureDir != "" {
		return nil, fmt.Errorf("fixture_vmdk and fixture_vmdk_dir are mutually exclusive")
	}
	if fixturePath != "" {
		abs, err := filepath.Abs(fixturePath)
		if err != nil {
			return nil, err
		}
		st, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("fixture_vmdk: %w", err)
		}
		if !st.Mode().IsRegular() {
			return nil, fmt.Errorf("fixture_vmdk is not a regular file: %s", abs)
		}
		fixturePath = abs
	}

	if fixtureDir != "" {
		abs, err := filepath.Abs(fixtureDir)
		if err != nil {
			return nil, err
		}
		st, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("fixture_vmdk_dir: %w", err)
		}
		if !st.IsDir() {
			return nil, fmt.Errorf("fixture_vmdk_dir is not a directory: %s", abs)
		}
		fixtureDir = abs
	}

	dir, err := os.MkdirTemp("", "chimera-fixtures-")
	if err != nil {
		return nil, err
	}
	s := &Store{
		dir: dir, fixturePath: fixturePath, fixtureDir: fixtureDir,
		vmNames: append([]string(nil), vmNames...), sizeMB: sizeMB,
		files: make(map[string]string), stopPoll: make(chan struct{}),
	}
	if fixtureDir != "" {
		if err := s.Rescan(); err != nil {
			return nil, err
		}
		go s.pollFixtureDir()
	}
	return s, nil
}

// Rescan re-scans fixture_vmdk_dir and updates the VM/file assignment. It is
// called once synchronously in New(), then periodically by pollFixtureDir so
// an operator can drop in a new VMDK without restarting the server. It is a
// no-op error if fixture_vmdk_dir isn't configured.
func (s *Store) Rescan() error {
	if s.fixtureDir == "" {
		return fmt.Errorf("fixture_vmdk_dir is not configured")
	}
	byVM, assignments, err := scanFixtureDir(s.fixtureDir, s.vmNames)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.byVM = byVM
	s.assignments = assignments
	s.mu.Unlock()
	return nil
}

func (s *Store) pollFixtureDir() {
	ticker := time.NewTicker(fixtureDirPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = s.Rescan() // best-effort: a transient read error just keeps the prior assignment
		case <-s.stopPoll:
			return
		}
	}
}

// scanFixtureDir lists dir's *.vmdk files and assigns each to a VM: first by
// matching the file's sanitized basename to a VM's sanitized name, then by
// pairing any leftover files with any leftover VMs in sorted order. Files
// left unassigned (more files than unmatched VMs) stay MethodUnassigned;
// VMs left unassigned (more VMs than unmatched files) are simply absent
// from byVM, so Prepare falls back to the generated-fixture path for them.
func scanFixtureDir(dir string, vmNames []string) (map[string]Assignment, []Assignment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("fixture_vmdk_dir: %w", err)
	}
	type diskFile struct {
		name string
		path string
		size int64
	}
	var files []diskFile
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".vmdk") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, diskFile{name: e.Name(), path: filepath.Join(dir, e.Name()), size: info.Size()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

	vmSorted := append([]string(nil), vmNames...)
	sort.Strings(vmSorted)
	vmByKey := make(map[string]string, len(vmSorted))
	for _, vm := range vmSorted {
		key := sanitize(vm)
		if _, exists := vmByKey[key]; !exists {
			vmByKey[key] = vm
		}
	}

	result := make([]Assignment, len(files))
	for i, f := range files {
		result[i] = Assignment{FileName: f.name, Path: f.path, SizeBytes: f.size, Method: MethodUnassigned}
	}
	byVM := make(map[string]Assignment, len(vmSorted))
	assignedVM := make(map[string]bool, len(vmSorted))

	// Pass 1: filename match.
	for i, f := range files {
		base := strings.TrimSuffix(f.name, filepath.Ext(f.name))
		key := sanitize(base)
		if vm, ok := vmByKey[key]; ok && !assignedVM[key] {
			a := Assignment{FileName: f.name, Path: f.path, SizeBytes: f.size, VMName: vm, Method: MethodNameMatch}
			result[i] = a
			byVM[key] = a
			assignedVM[key] = true
		}
	}

	// Pass 2: round robin over leftovers.
	var leftoverFileIdx []int
	for i := range files {
		if result[i].Method == MethodUnassigned {
			leftoverFileIdx = append(leftoverFileIdx, i)
		}
	}
	var leftoverVMs []string
	for _, vm := range vmSorted {
		if !assignedVM[sanitize(vm)] {
			leftoverVMs = append(leftoverVMs, vm)
		}
	}
	n := min(len(leftoverVMs), len(leftoverFileIdx))
	for i := range n {
		idx := leftoverFileIdx[i]
		vm := leftoverVMs[i]
		key := sanitize(vm)
		a := Assignment{FileName: files[idx].name, Path: files[idx].path, SizeBytes: files[idx].size, VMName: vm, Method: MethodRoundRobin}
		result[idx] = a
		byVM[key] = a
		assignedVM[key] = true
	}

	return byVM, result, nil
}

func (s *Store) Close() error {
	if s.fixtureDir != "" {
		close(s.stopPoll)
	}
	if s.dir == "" {
		return nil
	}
	return os.RemoveAll(s.dir)
}

func (s *Store) Alias() string { return defaultAlias }

// Directory returns the configured fixture_vmdk_dir, or "" if not configured.
func (s *Store) Directory() string { return s.fixtureDir }

// Assignments returns the file-centric view of the fixture_vmdk_dir scan,
// sorted by file name. Empty/nil when directory mode isn't configured.
func (s *Store) Assignments() []Assignment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.assignments == nil {
		return nil
	}
	out := make([]Assignment, len(s.assignments))
	copy(out, s.assignments)
	return out
}

// SourceFor reports how vmName's exported disk fixture is sourced: MethodSharedFile
// (single fixture_vmdk for every VM), MethodNameMatch/MethodRoundRobin (a
// fixture_vmdk_dir assignment), or MethodGenerated (per-VM synthetic fixture,
// fileName "").
func (s *Store) SourceFor(vmName string) (method, fileName string) {
	if s.fixturePath != "" {
		return MethodSharedFile, filepath.Base(s.fixturePath)
	}
	if s.fixtureDir != "" {
		s.mu.RLock()
		a, ok := s.byVM[sanitize(vmName)]
		s.mu.RUnlock()
		if ok {
			return a.Method, a.FileName
		}
	}
	return MethodGenerated, ""
}

func (s *Store) Prepare(vmName string) (string, int64, error) {
	if s.fixturePath != "" {
		st, err := os.Stat(s.fixturePath)
		if err != nil {
			return "", 0, err
		}
		return s.fixturePath, st.Size(), nil
	}
	if s.fixtureDir != "" {
		s.mu.RLock()
		a, ok := s.byVM[sanitize(vmName)]
		s.mu.RUnlock()
		if ok {
			st, err := os.Stat(a.Path)
			if err != nil {
				return "", 0, err
			}
			return a.Path, st.Size(), nil
		}
		// No directory-mode assignment for this VM: fall through to the
		// generated synthetic fixture below.
	}

	safe := sanitize(vmName)
	p := filepath.Join(s.dir, safe+".vmdk")
	if st, err := os.Stat(p); err == nil {
		return p, st.Size(), nil
	}

	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			st, statErr := os.Stat(p)
			return p, st.Size(), statErr
		}
		return "", 0, err
	}
	defer f.Close()

	// Deterministic transport fixture. It is intentionally not a real VMDK.
	// Supply fixture_vmdk to test hyper2kvm/qemu conversion as well.
	marker := []byte("CHIMERA-TRANSPORT-FIXTURE\n")
	if _, err := f.Write(marker); err != nil {
		return "", 0, err
	}
	target := int64(s.sizeMB) * 1024 * 1024
	remaining := target - int64(len(marker))
	block := make([]byte, 1024*1024)
	for i := range block {
		block[i] = byte((i * 31) % 251)
	}
	for remaining > 0 {
		n := int64(len(block))
		if n > remaining {
			n = remaining
		}
		if _, err := f.Write(block[:n]); err != nil {
			return "", 0, err
		}
		remaining -= n
	}
	if err := f.Sync(); err != nil {
		return "", 0, err
	}
	return p, target, nil
}

func (s *Store) Register(leaseID, alias, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[key(leaseID, alias)] = path
}

func (s *Store) UnregisterLease(leaseID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := leaseID + "/"
	for k := range s.files {
		if strings.HasPrefix(k, prefix) {
			delete(s.files, k)
		}
	}
}

func (s *Store) Lookup(leaseID, alias string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.files[key(leaseID, alias)]
	return p, ok
}

func key(leaseID, alias string) string { return leaseID + "/" + alias }

func sanitize(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "vm"
	}
	r := strings.NewReplacer("/", "-", "\\", "-", "..", "-", ":", "-", "\x00", "")
	v = r.Replace(v)
	if len(v) > 120 {
		v = v[:120]
	}
	return v
}

// CopyRange copies [start, end] inclusive. end < 0 means EOF.
func CopyRange(dst io.Writer, f *os.File, start, end int64) (int64, error) {
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return 0, err
	}
	if end >= start {
		return io.CopyN(dst, f, end-start+1)
	}
	return io.Copy(dst, f)
}
