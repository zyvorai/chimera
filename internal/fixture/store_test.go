// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package fixture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedFixtureAndRegistry(t *testing.T) {
	s, err := New("", "", nil, 1, nil)
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

	s, err := New("", dir, nil, 1, []string{"DC0_H0_VM0", "DC0_H0_VM1"})
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

	s, err := New("", dir, nil, 1, []string{"vmA", "vmB", "vmC"})
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

	if _, err := New(f, dir, nil, 1, nil); err == nil {
		t.Fatal("expected error when fixture_vmdk and fixture_vmdk_dir are both set")
	}
}

func TestDirectoryFixtureIgnoresNonVMDKFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "readme.txt"), 5)
	writeFile(t, filepath.Join(dir, "disk.vmdk"), 5)

	s, err := New("", dir, nil, 1, []string{"vmA"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	assignments := s.Assignments()
	if len(assignments) != 1 || assignments[0].FileName != "disk.vmdk" {
		t.Fatalf("assignments=%+v, want only disk.vmdk", assignments)
	}
}

func TestDirectoryFixturePicksUpNewFileOnRescan(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "DC0_H0_VM0.vmdk"), 111)

	s, err := New("", dir, nil, 1, []string{"DC0_H0_VM0", "DC0_H0_VM1"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if method, _ := s.SourceFor("DC0_H0_VM1"); method != MethodGenerated {
		t.Fatalf("before adding a file, SourceFor(VM1)=%q, want generated", method)
	}

	// Drop in a new file after construction, matching the second VM by name.
	writeFile(t, filepath.Join(dir, "DC0_H0_VM1.vmdk"), 222)

	// Without an explicit Rescan, the old assignment should still hold.
	if method, _ := s.SourceFor("DC0_H0_VM1"); method != MethodGenerated {
		t.Fatalf("immediately after adding a file (no rescan yet), SourceFor(VM1)=%q, want still generated", method)
	}

	if err := s.Rescan(); err != nil {
		t.Fatal(err)
	}
	if method, file := s.SourceFor("DC0_H0_VM1"); method != MethodNameMatch || file != "DC0_H0_VM1.vmdk" {
		t.Fatalf("after Rescan, SourceFor(VM1)=%q/%q, want name-match/DC0_H0_VM1.vmdk", method, file)
	}
	p, n, err := s.Prepare("DC0_H0_VM1")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "DC0_H0_VM1.vmdk" || n != 222 {
		t.Fatalf("p=%q n=%d, want DC0_H0_VM1.vmdk/222", p, n)
	}
}

func TestRescanErrorsWithoutFixtureDir(t *testing.T) {
	s, err := New("", "", nil, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Rescan(); err == nil {
		t.Fatal("expected Rescan to error when fixture_vmdk_dir isn't configured")
	}
}

func TestSetOverridePinsFileToVM(t *testing.T) {
	dir := t.TempDir()
	// "stray.vmdk" would otherwise round-robin to whichever VM is first in
	// sorted leftover order; pin it to VM1 instead of VM0.
	writeFile(t, filepath.Join(dir, "stray.vmdk"), 111)

	s, err := New("", dir, nil, 1, []string{"VM0", "VM1"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if method, _ := s.SourceFor("VM0"); method != MethodRoundRobin {
		t.Fatalf("before override: SourceFor(VM0)=%q, want round-robin", method)
	}

	if err := s.SetOverride(0, "stray.vmdk", "VM1"); err != nil {
		t.Fatal(err)
	}
	if method, file := s.SourceFor("VM1"); method != MethodManual || file != "stray.vmdk" {
		t.Fatalf("SourceFor(VM1)=%q/%q, want manual/stray.vmdk", method, file)
	}
	if method, _ := s.SourceFor("VM0"); method != MethodGenerated {
		t.Fatalf("SourceFor(VM0)=%q, want generated (no files left for it)", method)
	}

	if err := s.SetOverride(0, "stray.vmdk", "does-not-exist"); err == nil {
		t.Fatal("expected SetOverride to reject an unknown VM")
	}
}

func TestClearOverrideFallsBackToAutoMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "stray.vmdk"), 111)

	s, err := New("", dir, nil, 1, []string{"VM0", "VM1"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.SetOverride(0, "stray.vmdk", "VM1"); err != nil {
		t.Fatal(err)
	}
	if method, _ := s.SourceFor("VM1"); method != MethodManual {
		t.Fatalf("SourceFor(VM1)=%q, want manual", method)
	}

	if err := s.ClearOverride(0, "stray.vmdk"); err != nil {
		t.Fatal(err)
	}
	if method, _ := s.SourceFor("VM0"); method != MethodRoundRobin {
		t.Fatalf("after clear: SourceFor(VM0)=%q, want round-robin", method)
	}
}

func TestDirectoryFixtureRecursesSubdirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "web-tier"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "web-tier", "DC0_H0_VM0.vmdk"), 111)

	s, err := New("", dir, nil, 1, []string{"DC0_H0_VM0"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Name-match works on the file's own basename even though it's nested.
	if method, file := s.SourceFor("DC0_H0_VM0"); method != MethodNameMatch || file != "web-tier/DC0_H0_VM0.vmdk" {
		t.Fatalf("SourceFor=%q/%q, want name-match/web-tier/DC0_H0_VM0.vmdk", method, file)
	}
	assignments := s.Assignments()
	if len(assignments) != 1 || assignments[0].Root != 0 {
		t.Fatalf("assignments=%+v, want 1 entry at root 0", assignments)
	}
}

func TestMultipleWatchDirectoriesAreScanned(t *testing.T) {
	primary := t.TempDir()
	extra := t.TempDir()
	writeFile(t, filepath.Join(primary, "DC0_H0_VM0.vmdk"), 111)
	writeFile(t, filepath.Join(extra, "DC0_H0_VM1.vmdk"), 222)

	s, err := New("", primary, []string{extra}, 1, []string{"DC0_H0_VM0", "DC0_H0_VM1"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if got := s.Roots(); len(got) != 2 || got[0] != primary || got[1] != extra {
		t.Fatalf("Roots()=%v, want [%s %s]", got, primary, extra)
	}
	if method, _ := s.SourceFor("DC0_H0_VM0"); method != MethodNameMatch {
		t.Fatalf("SourceFor(VM0)=%q, want name-match (from primary root)", method)
	}
	if method, _ := s.SourceFor("DC0_H0_VM1"); method != MethodNameMatch {
		t.Fatalf("SourceFor(VM1)=%q, want name-match (from extra root)", method)
	}
}

func TestOverrideKeysAreUniqueAcrossRoots(t *testing.T) {
	primary := t.TempDir()
	extra := t.TempDir()
	// Same relative file name in both roots — must not collide.
	writeFile(t, filepath.Join(primary, "stray.vmdk"), 111)
	writeFile(t, filepath.Join(extra, "stray.vmdk"), 222)

	s, err := New("", primary, []string{extra}, 1, []string{"VM0", "VM1"})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.SetOverride(1, "stray.vmdk", "VM1"); err != nil {
		t.Fatal(err)
	}
	if method, file := s.SourceFor("VM1"); method != MethodManual || file != "stray.vmdk" {
		t.Fatalf("SourceFor(VM1)=%q/%q, want manual/stray.vmdk", method, file)
	}
	// The primary root's same-named file must still be independently
	// available for VM0 via round-robin, not shadowed by the override.
	if method, _ := s.SourceFor("VM0"); method != MethodRoundRobin {
		t.Fatalf("SourceFor(VM0)=%q, want round-robin (unaffected by root-1 override)", method)
	}
	for _, a := range s.Assignments() {
		if a.FileName == "stray.vmdk" && a.Root == 0 && a.VMName != "VM0" {
			t.Fatalf("primary root's stray.vmdk unexpectedly assigned to %q", a.VMName)
		}
	}
}

func writeFile(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.WriteFile(path, make([]byte, size), 0600); err != nil {
		t.Fatal(err)
	}
}
