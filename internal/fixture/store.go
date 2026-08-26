package fixture

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const defaultAlias = "disk-0.vmdk"

type Store struct {
	mu          sync.RWMutex
	dir         string
	fixturePath string
	sizeMB      int
	files       map[string]string
}

func New(fixturePath string, sizeMB int) (*Store, error) {
	if sizeMB < 1 {
		sizeMB = 16
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
	dir, err := os.MkdirTemp("", "chimera-fixtures-")
	if err != nil {
		return nil, err
	}
	return &Store{dir: dir, fixturePath: fixturePath, sizeMB: sizeMB, files: make(map[string]string)}, nil
}

func (s *Store) Close() error {
	if s.dir == "" {
		return nil
	}
	return os.RemoveAll(s.dir)
}

func (s *Store) Alias() string { return defaultAlias }

func (s *Store) Prepare(vmName string) (string, int64, error) {
	if s.fixturePath != "" {
		st, err := os.Stat(s.fixturePath)
		if err != nil {
			return "", 0, err
		}
		return s.fixturePath, st.Size(), nil
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
