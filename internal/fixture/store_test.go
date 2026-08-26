package fixture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedFixtureAndRegistry(t *testing.T) {
	s, err := New("", "", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p, n, err := s.Prepare("demo/vm")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1024*1024 {
		t.Fatalf("size=%d", n)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() != n {
		t.Fatalf("stat=%d want=%d", st.Size(), n)
	}
	s.Register("lease-1", "disk-0.vmdk", p)
	if got, ok := s.Lookup("lease-1", "disk-0.vmdk"); !ok || got != p {
		t.Fatalf("lookup=%q ok=%v", got, ok)
	}
	s.UnregisterLease("lease-1")
	if _, ok := s.Lookup("lease-1", "disk-0.vmdk"); ok {
		t.Fatal("lease should be removed")
	}
}

func TestDirectoryFixtureNameMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "DC0_H0_VM0.vmdk"), 111)
	writeFile(t, filepath.Join(dir, "stray.vmdk"), 222)

	s, err := New("", dir, 1, []string{"DC0_H0_VM0", "DC0_H0_VM1"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, n, err := s.Prepare("DC0_H0_VM0")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "DC0_H0_VM0.vmdk" || n != 111 {
		t.Fatalf("p=%q n=%d, want DC0_H0_VM0.vmdk/111", p, n)
	}
	if method, file := s.SourceFor("DC0_H0_VM0"); method != MethodNameMatch || file != "DC0_H0_VM0.vmdk" {
		t.Fatalf("SourceFor(matched)=%q/%q", method, file)
	}

	p, n, err = s.Prepare("DC0_H0_VM1")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "stray.vmdk" || n != 222 {
		t.Fatalf("p=%q n=%d, want stray.vmdk/222 (round-robin leftover)", p, n)
	}
	if method, _ := s.SourceFor("DC0_H0_VM1"); method != MethodRoundRobin {
		t.Fatalf("SourceFor(roundrobin)=%q", method)
	}

	assignments := s.Assignments()
	if len(assignments) != 2 {
		t.Fatalf("assignments=%d, want 2", len(assignments))
	}
}

func TestDirectoryFixtureRoundRobinAndLeftoverGenerated(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "onlyfile.vmdk"), 333)

	s, err := New("", dir, 1, []string{"vmA", "vmB", "vmC"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	robin := 0
	generated := 0
	for _, vm := range []string{"vmA", "vmB", "vmC"} {
		switch method, _ := s.SourceFor(vm); method {
		case MethodRoundRobin:
			robin++
		case MethodGenerated:
			generated++
		default:
			t.Fatalf("vm=%q unexpected method=%q", vm, method)
		}
	}
	if robin != 1 || generated != 2 {
		t.Fatalf("robin=%d generated=%d, want 1/2", robin, generated)
	}
}

func TestDirectoryFixtureAndSingleFileMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "one.vmdk")
	writeFile(t, f, 10)

	if _, err := New(f, dir, 1, nil); err == nil {
		t.Fatal("expected error when fixture_vmdk and fixture_vmdk_dir are both set")
	}
}

func TestDirectoryFixtureIgnoresNonVMDKFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "readme.txt"), 5)
	writeFile(t, filepath.Join(dir, "disk.vmdk"), 5)

	s, err := New("", dir, 1, []string{"vmA"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	assignments := s.Assignments()
	if len(assignments) != 1 || assignments[0].FileName != "disk.vmdk" {
		t.Fatalf("assignments=%+v, want only disk.vmdk", assignments)
	}
}

func writeFile(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.WriteFile(path, make([]byte, size), 0600); err != nil {
		t.Fatal(err)
	}
}
