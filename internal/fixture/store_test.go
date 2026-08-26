package fixture

import (
	"os"
	"testing"
)

func TestGeneratedFixtureAndRegistry(t *testing.T) {
	s, err := New("", 1)
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
