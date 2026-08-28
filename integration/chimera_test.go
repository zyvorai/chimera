// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/types"

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

// TestCreateDescriptorThenExportVm exercises Transiva's actual real client
// path (OvfManager.CreateDescriptor, then VirtualMachine.Export), unlike
// selftest.Run which only calls Export directly. This is the exact sequence
// that exposed a session-override bug in exportshim: the Handler hook's
// override was silently discarded by Session.Get falling back to the real
// registered OvfManager, identical to the earlier ExportVm bug.
func TestCreateDescriptorThenExportVm(t *testing.T) {
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

	u := *l.URL
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	client, err := govmomi.NewClient(testCtx, &u, true)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer client.Logout(context.Background())

	f := find.NewFinder(client.Client, true)
	dcs, err := f.DatacenterList(testCtx, "*")
	if err != nil || len(dcs) == 0 {
		t.Fatalf("datacenter list: %v", err)
	}
	f.SetDatacenter(dcs[0])
	vms, err := f.VirtualMachineList(testCtx, "*")
	if err != nil || len(vms) == 0 {
		t.Fatalf("vm list: %v", err)
	}

	mgr := ovf.NewManager(client.Client)
	res, err := mgr.CreateDescriptor(testCtx, vms[0], types.OvfCreateDescriptorParams{})
	if err != nil {
		t.Fatalf("CreateDescriptor: %v", err)
	}
	if res.OvfDescriptor == "" {
		t.Fatal("CreateDescriptor returned an empty descriptor")
	}

	lease, err := vms[0].Export(testCtx)
	if err != nil {
		t.Fatalf("Export after CreateDescriptor: %v", err)
	}
	_ = lease.Abort(context.Background(), nil)
}
