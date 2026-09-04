// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zyvorai/chimera/internal/personas/common"
)

const (
	ec2Namespace = "http://ec2.amazonaws.com/doc/2016-11-15/"
	blockSize    = 512 * 1024
)

type Snapshot struct {
	ID        string
	VolumeID  string
	VMID      string
	DiskID    string
	State     string
	StartedAt time.Time
}

type Server struct {
	AccessKey string
	SecretKey string
	Region    string
	Store     *common.Store

	mu        sync.RWMutex
	snapshots map[string]Snapshot
	nextSnap  atomic.Uint64
}

func New(accessKey, secretKey string, vmCount int) *Server {
	return &Server{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    "us-east-1",
		Store:     common.NewStore(vmCount),
		snapshots: map[string]Snapshot{},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	service := "ec2"
	if strings.HasPrefix(r.URL.Path, "/snapshots/") {
		service = "ebs"
	}
	if err := s.verifySigV4(r, body, service); err != nil {
		writeEC2Error(w, http.StatusUnauthorized, "AuthFailure", err.Error())
		return
	}

	if strings.HasPrefix(r.URL.Path, "/snapshots/") {
		s.serveEBS(w, r)
		return
	}
	if r.URL.Path != "/" && r.URL.Path != "/ec2" {
		http.NotFound(w, r)
		return
	}

	params := r.URL.Query()
	if len(body) > 0 {
		if form, err := url.ParseQuery(string(body)); err == nil {
			for k, vv := range form {
				for _, v := range vv {
					params.Add(k, v)
				}
			}
		}
	}
	action := params.Get("Action")
	switch action {
	case "DescribeInstances":
		s.describeInstances(w)
	case "DescribeVolumes":
		s.describeVolumes(w)
	case "StartInstances":
		s.changePower(w, params, "ON", "StartInstancesResponse")
	case "StopInstances":
		s.changePower(w, params, "OFF", "StopInstancesResponse")
	case "CreateSnapshot":
		s.createSnapshot(w, params)
	case "DescribeSnapshots":
		s.describeSnapshots(w, params)
	default:
		writeEC2Error(w, http.StatusBadRequest, "InvalidAction", "unsupported Action "+action)
	}
}

func (s *Server) describeInstances(w http.ResponseWriter) {
	type ebs struct {
		VolumeID string `xml:"volumeId"`
		Status   string `xml:"status"`
	}
	type mapping struct {
		DeviceName string `xml:"deviceName"`
		EBS        ebs    `xml:"ebs"`
	}
	type instance struct {
		InstanceID         string    `xml:"instanceId"`
		ImageID            string    `xml:"imageId"`
		InstanceState      ec2State  `xml:"instanceState"`
		PrivateDNSName     string    `xml:"privateDnsName"`
		PrivateIPAddress   string    `xml:"privateIpAddress"`
		InstanceType       string    `xml:"instanceType"`
		Architecture       string    `xml:"architecture"`
		RootDeviceType     string    `xml:"rootDeviceType"`
		RootDeviceName     string    `xml:"rootDeviceName"`
		BlockDeviceMapping []mapping `xml:"blockDeviceMapping>item"`
	}
	type reservation struct {
		ReservationID string     `xml:"reservationId"`
		OwnerID       string     `xml:"ownerId"`
		Instances     []instance `xml:"instancesSet>item"`
	}
	resp := struct {
		XMLName      xml.Name      `xml:"DescribeInstancesResponse"`
		Xmlns        string        `xml:"xmlns,attr"`
		RequestID    string        `xml:"requestId"`
		Reservations []reservation `xml:"reservationSet>item"`
	}{Xmlns: ec2Namespace, RequestID: requestID()}

	for i, vm := range s.Store.List() {
		maps := make([]mapping, 0, len(vm.Disks))
		for j, d := range vm.Disks {
			maps = append(maps, mapping{DeviceName: deviceName(j), EBS: ebs{VolumeID: volumeID(d.ID), Status: "attached"}})
		}
		resp.Reservations = append(resp.Reservations, reservation{
			ReservationID: fmt.Sprintf("r-chimera%04d", i+1), OwnerID: "000000000000",
			Instances: []instance{{
				InstanceID: instanceID(vm.ID), ImageID: "ami-chimera00000001", InstanceState: state(vm.Power),
				PrivateDNSName: vm.Name + ".chimera.internal", PrivateIPAddress: fmt.Sprintf("10.0.0.%d", i+10),
				InstanceType: "m6i.large", Architecture: "x86_64", RootDeviceType: "ebs", RootDeviceName: "/dev/sda1",
				BlockDeviceMapping: maps,
			}},
		})
	}
	writeXML(w, resp)
}

func (s *Server) describeVolumes(w http.ResponseWriter) {
	type attachment struct {
		VolumeID   string `xml:"volumeId"`
		InstanceID string `xml:"instanceId"`
		Device     string `xml:"device"`
		Status     string `xml:"status"`
	}
	type volume struct {
		VolumeID         string       `xml:"volumeId"`
		Size             int          `xml:"size"`
		SnapshotID       string       `xml:"snapshotId"`
		AvailabilityZone string       `xml:"availabilityZone"`
		Status           string       `xml:"status"`
		VolumeType       string       `xml:"volumeType"`
		Attachments      []attachment `xml:"attachmentSet>item"`
	}
	resp := struct {
		XMLName   xml.Name `xml:"DescribeVolumesResponse"`
		Xmlns     string   `xml:"xmlns,attr"`
		RequestID string   `xml:"requestId"`
		Volumes   []volume `xml:"volumeSet>item"`
	}{Xmlns: ec2Namespace, RequestID: requestID()}
	for _, vm := range s.Store.List() {
		for j, d := range vm.Disks {
			resp.Volumes = append(resp.Volumes, volume{
				VolumeID: volumeID(d.ID), Size: gib(d.SizeBytes), AvailabilityZone: s.Region + "a", Status: "in-use", VolumeType: "gp3",
				Attachments: []attachment{{VolumeID: volumeID(d.ID), InstanceID: instanceID(vm.ID), Device: deviceName(j), Status: "attached"}},
			})
		}
	}
	writeXML(w, resp)
}

func (s *Server) changePower(w http.ResponseWriter, params url.Values, target, root string) {
	ids := indexed(params, "InstanceId")
	if len(ids) == 0 {
		writeEC2Error(w, http.StatusBadRequest, "MissingParameter", "InstanceId.1 is required")
		return
	}
	changes := make([]powerChange, 0, len(ids))
	for _, iid := range ids {
		vmID := vmID(iid)
		vm, ok := s.Store.Get(vmID)
		if !ok {
			writeEC2Error(w, http.StatusBadRequest, "InvalidInstanceID.NotFound", iid)
			return
		}
		prev := state(vm.Power)
		if _, ok := s.Store.SetPower(vmID, target); !ok {
			writeEC2Error(w, http.StatusInternalServerError, "InternalError", iid)
			return
		}
		changes = append(changes, powerChange{InstanceID: iid, CurrentState: state(target), PreviousState: prev})
	}
	if root == "StartInstancesResponse" {
		writeXML(w, startInstancesResponse{Xmlns: ec2Namespace, RequestID: requestID(), Instances: changes})
		return
	}
	writeXML(w, stopInstancesResponse{Xmlns: ec2Namespace, RequestID: requestID(), Instances: changes})
}

func (s *Server) createSnapshot(w http.ResponseWriter, params url.Values) {
	vol := params.Get("VolumeId")
	vm, disk, ok := s.lookupVolume(vol)
	if !ok {
		writeEC2Error(w, http.StatusBadRequest, "InvalidVolume.NotFound", vol)
		return
	}
	id := fmt.Sprintf("snap-chimera%08x", s.nextSnap.Add(1))
	snap := Snapshot{ID: id, VolumeID: vol, VMID: vm.ID, DiskID: disk.ID, State: "completed", StartedAt: time.Now().UTC()}
	s.mu.Lock()
	s.snapshots[id] = snap
	s.mu.Unlock()
	resp := struct {
		XMLName    xml.Name `xml:"CreateSnapshotResponse"`
		Xmlns      string   `xml:"xmlns,attr"`
		RequestID  string   `xml:"requestId"`
		SnapshotID string   `xml:"snapshotId"`
		VolumeID   string   `xml:"volumeId"`
		Status     string   `xml:"status"`
		StartTime  string   `xml:"startTime"`
		Progress   string   `xml:"progress"`
	}{Xmlns: ec2Namespace, RequestID: requestID(), SnapshotID: id, VolumeID: vol, Status: "completed", StartTime: snap.StartedAt.Format(time.RFC3339), Progress: "100%"}
	writeXML(w, resp)
}

func (s *Server) describeSnapshots(w http.ResponseWriter, params url.Values) {
	wanted := map[string]bool{}
	for _, id := range indexed(params, "SnapshotId") {
		wanted[id] = true
	}
	type snapshot struct {
		SnapshotID string `xml:"snapshotId"`
		VolumeID   string `xml:"volumeId"`
		Status     string `xml:"status"`
		StartTime  string `xml:"startTime"`
		Progress   string `xml:"progress"`
	}
	resp := struct {
		XMLName   xml.Name   `xml:"DescribeSnapshotsResponse"`
		Xmlns     string     `xml:"xmlns,attr"`
		RequestID string     `xml:"requestId"`
		Snapshots []snapshot `xml:"snapshotSet>item"`
	}{Xmlns: ec2Namespace, RequestID: requestID()}
	s.mu.RLock()
	ids := make([]string, 0, len(s.snapshots))
	for id := range s.snapshots {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		snap := s.snapshots[id]
		if len(wanted) > 0 && !wanted[id] {
			continue
		}
		resp.Snapshots = append(resp.Snapshots, snapshot{SnapshotID: id, VolumeID: snap.VolumeID, Status: snap.State, StartTime: snap.StartedAt.Format(time.RFC3339), Progress: "100%"})
	}
	s.mu.RUnlock()
	writeXML(w, resp)
}

func (s *Server) serveEBS(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[0] != "snapshots" || parts[2] != "blocks" {
		http.NotFound(w, r)
		return
	}
	snapID := parts[1]
	s.mu.RLock()
	snap, ok := s.snapshots[snapID]
	s.mu.RUnlock()
	if !ok {
		writeJSONError(w, http.StatusNotFound, "ResourceNotFoundException", "snapshot not found")
		return
	}
	if len(parts) == 3 && r.Method == http.MethodGet {
		blocks := []map[string]any{{"BlockIndex": 0, "BlockToken": blockToken(snapID, 0)}}
		writeJSON(w, map[string]any{
			"Blocks": blocks, "BlockSize": blockSize, "ExpiryTime": time.Now().Add(1 * time.Hour).Unix(), "VolumeSize": 1,
		})
		return
	}
	if len(parts) == 4 && r.Method == http.MethodGet {
		idx, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil || idx < 0 {
			writeJSONError(w, http.StatusBadRequest, "ValidationException", "bad block index")
			return
		}
		if r.URL.Query().Get("blockToken") != blockToken(snapID, idx) {
			writeJSONError(w, http.StatusBadRequest, "ValidationException", "bad block token")
			return
		}
		b := common.DiskBytes(snap.VMID, snap.DiskID+":"+strconv.FormatInt(idx, 10), blockSize)
		sum := sha256.Sum256(b)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.Header().Set("x-amz-Checksum", base64.StdEncoding.EncodeToString(sum[:]))
		w.Header().Set("x-amz-Checksum-Algorithm", "SHA256")
		w.Header().Set("x-amz-Data-Length", strconv.Itoa(len(b)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) lookupVolume(id string) (common.VM, common.Disk, bool) {
	for _, vm := range s.Store.List() {
		for _, disk := range vm.Disks {
			if volumeID(disk.ID) == id {
				return vm, disk, true
			}
		}
	}
	return common.VM{}, common.Disk{}, false
}

type powerChange struct {
	InstanceID    string   `xml:"instanceId"`
	CurrentState  ec2State `xml:"currentState"`
	PreviousState ec2State `xml:"previousState"`
}

type startInstancesResponse struct {
	XMLName   xml.Name      `xml:"StartInstancesResponse"`
	Xmlns     string        `xml:"xmlns,attr"`
	RequestID string        `xml:"requestId"`
	Instances []powerChange `xml:"instancesSet>item"`
}

type stopInstancesResponse struct {
	XMLName   xml.Name      `xml:"StopInstancesResponse"`
	Xmlns     string        `xml:"xmlns,attr"`
	RequestID string        `xml:"requestId"`
	Instances []powerChange `xml:"instancesSet>item"`
}

type ec2State struct {
	Code int    `xml:"code"`
	Name string `xml:"name"`
}

func state(power string) ec2State {
	if strings.EqualFold(power, "ON") {
		return ec2State{Code: 16, Name: "running"}
	}
	return ec2State{Code: 80, Name: "stopped"}
}

func instanceID(id string) string { return "i-chimera" + strings.TrimPrefix(id, "vm-") }
func vmID(id string) string       { return "vm-" + strings.TrimPrefix(id, "i-chimera") }
func volumeID(id string) string   { return "vol-chimera" + strings.TrimPrefix(id, "disk-") }
func deviceName(i int) string {
	if i == 0 {
		return "/dev/sda1"
	}
	return fmt.Sprintf("/dev/sd%c", 'f'+rune(i-1))
}
func gib(bytes int64) int {
	g := int((bytes + (1 << 30) - 1) >> 30)
	if g < 1 {
		return 1
	}
	return g
}
func blockToken(snapID string, idx int64) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:chimera", snapID, idx)))
	return hex.EncodeToString(h[:12])
}
func requestID() string {
	h := sha256.Sum256([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(h[:16])
}

func indexed(v url.Values, prefix string) []string {
	type pair struct {
		i int
		v string
	}
	var out []pair
	for k, vv := range v {
		if !strings.HasPrefix(k, prefix+".") || len(vv) == 0 {
			continue
		}
		i, err := strconv.Atoi(strings.TrimPrefix(k, prefix+"."))
		if err == nil {
			out = append(out, pair{i: i, v: vv[0]})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].i < out[j].i })
	vals := make([]string, 0, len(out))
	for _, p := range out {
		vals = append(vals, p.v)
	}
	return vals
}

func (s *Server) verifySigV4(r *http.Request, body []byte, expectedService string) error {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256 ") {
		return fmt.Errorf("missing AWS Signature Version 4 authorization")
	}
	attrs := map[string]string{}
	for _, part := range strings.Split(strings.TrimPrefix(auth, "AWS4-HMAC-SHA256 "), ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 {
			attrs[kv[0]] = kv[1]
		}
	}
	cred := strings.Split(attrs["Credential"], "/")
	if len(cred) != 5 || cred[4] != "aws4_request" {
		return fmt.Errorf("invalid Credential scope")
	}
	if subtle.ConstantTimeCompare([]byte(cred[0]), []byte(s.AccessKey)) != 1 {
		return fmt.Errorf("unknown access key")
	}
	date, region, service := cred[1], cred[2], cred[3]
	if service != expectedService {
		return fmt.Errorf("signature service %q does not match %q", service, expectedService)
	}
	signedHeaders := attrs["SignedHeaders"]
	if signedHeaders == "" || attrs["Signature"] == "" {
		return fmt.Errorf("missing signed headers or signature")
	}
	amzDate := r.Header.Get("X-Amz-Date")
	if amzDate == "" {
		return fmt.Errorf("missing X-Amz-Date")
	}
	if len(amzDate) < 8 || amzDate[:8] != date {
		return fmt.Errorf("credential date does not match X-Amz-Date")
	}
	payloadHash := r.Header.Get("X-Amz-Content-Sha256")
	if payloadHash == "" {
		h := sha256.Sum256(body)
		payloadHash = hex.EncodeToString(h[:])
	}

	names := strings.Split(signedHeaders, ";")
	canonicalHeaders := strings.Builder{}
	for _, name := range names {
		name = strings.ToLower(strings.TrimSpace(name))
		var value string
		if name == "host" {
			value = r.Host
		} else {
			value = r.Header.Get(name)
		}
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteByte(':')
		canonicalHeaders.WriteString(strings.Join(strings.Fields(value), " "))
		canonicalHeaders.WriteByte('\n')
	}
	canonicalURI := r.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalRequest := strings.Join([]string{
		r.Method,
		canonicalURI,
		canonicalQuery(r.URL.Query()),
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")
	crHash := sha256.Sum256([]byte(canonicalRequest))
	scope := strings.Join([]string{date, region, service, "aws4_request"}, "/")
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(crHash[:])
	kDate := hmacSHA256([]byte("AWS4"+s.SecretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	expected := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))
	got := strings.ToLower(attrs["Signature"])
	if subtle.ConstantTimeCompare([]byte(expected), []byte(got)) != 1 {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func canonicalQuery(v url.Values) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		vals := append([]string(nil), v[k]...)
		sort.Strings(vals)
		for _, val := range vals {
			parts = append(parts, awsEncode(k)+"="+awsEncode(val))
		}
	}
	return strings.Join(parts, "&")
}

func awsEncode(v string) string {
	e := url.QueryEscape(v)
	e = strings.ReplaceAll(e, "+", "%20")
	e = strings.ReplaceAll(e, "%7E", "~")
	return e
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	_, _ = m.Write([]byte(data))
	return m.Sum(nil)
}

func writeXML(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "text/xml;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(v)
}

func writeEC2Error(w http.ResponseWriter, status int, code, message string) {
	resp := struct {
		XMLName xml.Name `xml:"Response"`
		Errors  []struct {
			Code    string `xml:"Code"`
			Message string `xml:"Message"`
		} `xml:"Errors>Error"`
		RequestID string `xml:"RequestID"`
	}{RequestID: requestID()}
	resp.Errors = append(resp.Errors, struct {
		Code    string `xml:"Code"`
		Message string `xml:"Message"`
	}{Code: code, Message: message})
	w.Header().Set("Content-Type", "text/xml;charset=UTF-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"Code": code, "Message": message})
}
