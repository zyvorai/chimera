package exportshim

import (
	"fmt"
	"html"
	"path"
	"strings"

	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/zyvorai/chimera/internal/fixture"
)

type Shim struct {
	Store      *fixture.Store
	PublicBase string
}

func New(store *fixture.Store, publicBase string) *Shim {
	return &Shim{Store: store, PublicBase: strings.TrimRight(publicBase, "/")}
}

// Handler plugs the two export operations Transiva needs into govmomi's
// simulator inventory: OvfManager.CreateDescriptor and VirtualMachine.ExportVm.
func (s *Shim) Handler(ctx *simulator.Context, m *simulator.Method) (mo.Reference, types.BaseMethodFault) {
	switch m.Name {
	case "CreateDescriptor":
		if m.This.Type == "OvfManager" {
			return &ovfHandler{shim: s, ref: m.This}, nil
		}
	case "ExportVm":
		if m.This.Type == "VirtualMachine" {
			return &vmHandler{shim: s, ref: m.This}, nil
		}
	}
	return nil, nil
}

type ovfHandler struct {
	shim *Shim
	ref  types.ManagedObjectReference
}

func (h *ovfHandler) Reference() types.ManagedObjectReference { return h.ref }

func (h *ovfHandler) CreateDescriptor(ctx *simulator.Context, req *types.CreateDescriptor) soap.HasFault {
	name := entityName(ctx, req.Obj)
	_, size, err := h.shim.Store.Prepare(name)
	if err != nil {
		return &methods.CreateDescriptorBody{Fault_: simulator.Fault(err.Error(), &types.FileFault{})}
	}
	disk := h.shim.Store.Alias()
	descriptor := descriptorXML(name, disk, size)
	return &methods.CreateDescriptorBody{Res: &types.CreateDescriptorResponse{
		Returnval: types.OvfCreateDescriptorResult{OvfDescriptor: descriptor},
	}}
}

type vmHandler struct {
	shim *Shim
	ref  types.ManagedObjectReference
}

func (h *vmHandler) Reference() types.ManagedObjectReference { return h.ref }

func (h *vmHandler) ExportVm(ctx *simulator.Context, _ *types.ExportVm) soap.HasFault {
	name := entityName(ctx, h.ref)
	filePath, size, err := h.shim.Store.Prepare(name)
	if err != nil {
		return &methods.ExportVmBody{Fault_: simulator.Fault(err.Error(), &types.FileFault{})}
	}

	lease := &leaseHandler{shim: h.shim}
	lease.State = types.HttpNfcLeaseStateReady
	ctx.Session.Put(lease)
	ref := lease.Reference()
	alias := h.shim.Store.Alias()
	h.shim.Store.Register(ref.Value, alias, filePath)

	u := h.shim.PublicBase + "/chimera-nfc/" + path.Join(ref.Value, alias)
	lease.Info = &types.HttpNfcLeaseInfo{
		Lease:  ref,
		Entity: h.ref,
		DeviceUrl: []types.HttpNfcLeaseDeviceUrl{{
			Key:           fmt.Sprintf("/%s/disk/0", h.ref.Value),
			Url:           u,
			SslThumbprint: "",
			Disk:          types.NewBool(true),
			TargetId:      alias,
			DatastoreKey:  "",
			FileSize:      size,
		}},
		LeaseTimeout:          300,
		TotalDiskCapacityInKB: size / 1024,
	}

	return &methods.ExportVmBody{Res: &types.ExportVmResponse{Returnval: ref}}
}

type leaseHandler struct {
	mo.HttpNfcLease
	shim *Shim
}

func (l *leaseHandler) HttpNfcLeaseComplete(ctx *simulator.Context, req *types.HttpNfcLeaseComplete) soap.HasFault {
	l.shim.Store.UnregisterLease(req.This.Value)
	ctx.Session.Remove(ctx, req.This)
	return &methods.HttpNfcLeaseCompleteBody{Res: new(types.HttpNfcLeaseCompleteResponse)}
}

func (l *leaseHandler) HttpNfcLeaseAbort(ctx *simulator.Context, req *types.HttpNfcLeaseAbort) soap.HasFault {
	l.shim.Store.UnregisterLease(req.This.Value)
	ctx.Session.Remove(ctx, req.This)
	return &methods.HttpNfcLeaseAbortBody{Res: new(types.HttpNfcLeaseAbortResponse)}
}

func (l *leaseHandler) HttpNfcLeaseProgress(_ *simulator.Context, req *types.HttpNfcLeaseProgress) soap.HasFault {
	l.TransferProgress = req.Percent
	return &methods.HttpNfcLeaseProgressBody{Res: new(types.HttpNfcLeaseProgressResponse)}
}

func (l *leaseHandler) HttpNfcLeaseGetManifest(_ *simulator.Context, _ *types.HttpNfcLeaseGetManifest) soap.HasFault {
	return &methods.HttpNfcLeaseGetManifestBody{Res: &types.HttpNfcLeaseGetManifestResponse{}}
}

func entityName(ctx *simulator.Context, ref types.ManagedObjectReference) string {
	if obj := ctx.Map.Get(ref); obj != nil {
		if e, ok := obj.(mo.Entity); ok && e.Entity().Name != "" {
			return e.Entity().Name
		}
	}
	return ref.Value
}

func descriptorXML(vmName, diskName string, diskBytes int64) string {
	capacity := diskBytes
	if capacity < 1 {
		capacity = 1
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://schemas.dmtf.org/ovf/envelope/1" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData">
  <References><File ovf:id="file1" ovf:href="%s" ovf:size="%d"/></References>
  <DiskSection><Info>Virtual disk information</Info><Disk ovf:diskId="vmdisk1" ovf:fileRef="file1" ovf:capacity="%d" ovf:capacityAllocationUnits="byte"/></DiskSection>
  <VirtualSystem ovf:id="%s"><Info>Chimera VM</Info><Name>%s</Name><VirtualHardwareSection><Info>Virtual hardware</Info></VirtualHardwareSection></VirtualSystem>
</Envelope>`, html.EscapeString(diskName), diskBytes, capacity, html.EscapeString(vmName), html.EscapeString(vmName))
}
