package selftest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

type Result struct {
	Datacenter string
	VM         string
	NFCFiles   int
	BytesRead  int64
}

func Run(ctx context.Context, endpoint, username, password string, insecure bool) (*Result, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	u.User = url.UserPassword(username, password)
	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	defer c.Logout(context.Background())

	f := find.NewFinder(c.Client, true)
	dcs, err := f.DatacenterList(ctx, "*")
	if err != nil || len(dcs) == 0 {
		return nil, fmt.Errorf("datacenter list: %w", err)
	}
	f.SetDatacenter(dcs[0])
	vms, err := f.VirtualMachineList(ctx, "*")
	if err != nil || len(vms) == 0 {
		return nil, fmt.Errorf("vm list: %w", err)
	}
	lease, err := vms[0].Export(ctx)
	if err != nil {
		return nil, fmt.Errorf("export lease: %w", err)
	}
	defer lease.Abort(context.Background(), nil)
	lctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	info, err := lease.Wait(lctx, nil)
	if err != nil {
		return nil, fmt.Errorf("lease wait: %w", err)
	}
	var read int64
	if len(info.Items) > 0 {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, info.Items[0].URL.String(), nil)
		err = c.Client.Do(ctx, req, func(res *http.Response) error {
			if res.StatusCode != 200 && res.StatusCode != 206 {
				return fmt.Errorf("nfc http %s", res.Status)
			}
			n, e := io.CopyN(io.Discard, res.Body, 4096)
			if e != nil && e != io.EOF {
				return e
			}
			read = n
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("nfc read: %w", err)
		}
	}
	_ = lease.Complete(ctx)
	return &Result{Datacenter: dcs[0].Name(), VM: vms[0].Name(), NFCFiles: len(info.Items), BytesRead: read}, nil
}
