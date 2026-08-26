package integration

import (
	"context"
	"testing"
	"time"

	"github.com/zyvorai/chimera/internal/config"
	"github.com/zyvorai/chimera/internal/lab"
	"github.com/zyvorai/chimera/internal/selftest"
)

func TestLoginInventoryExportAndNFC(t *testing.T) {
	cfg := config.Default()
	cfg.Listen = "127.0.0.1:0"
	cfg.PublicHost = ""
	cfg.TLS = false
	cfg.VMsPerPool = 2
	cfg.FixtureSizeMB = 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l, err := lab.Start(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := l.Close(closeCtx); err != nil {
			t.Errorf("close: %v", err)
		}
	}()

	testCtx, testCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer testCancel()
	got, err := selftest.Run(testCtx, l.URL.String(), cfg.Username, cfg.Password, true, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Datacenter == "" || got.VM == "" || got.NFCFiles < 1 || got.BytesRead < 1 {
		t.Fatalf("unexpected selftest result: %+v", got)
	}
}

func TestBadPasswordRejected(t *testing.T) {
	cfg := config.Default()
	cfg.Listen = "127.0.0.1:0"
	cfg.FixtureSizeMB = 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l, err := lab.Start(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		_ = l.Close(closeCtx)
	}()

	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()
	if _, err := selftest.Run(testCtx, l.URL.String(), cfg.Username, "definitely-wrong", true, "", ""); err == nil {
		t.Fatal("expected invalid login")
	}
}
