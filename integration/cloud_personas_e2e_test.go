// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	awspersona "github.com/zyvorai/chimera/internal/personas/aws"
	azurepersona "github.com/zyvorai/chimera/internal/personas/azure"
)

func TestAWSEC2DiscoveryPowerSnapshotAndBlockExportE2E(t *testing.T) {
	const accessKey = "AKIDCHIMERATEST"
	const secretKey = "chimera-secret-key"
	ts := httptest.NewServer(awspersona.New(accessKey, secretKey, 2))
	defer ts.Close()

	call := func(service string, method string, target string, form url.Values) *http.Response {
		t.Helper()
		var body []byte
		if form != nil {
			body = []byte(form.Encode())
		}
		req, err := http.NewRequest(method, ts.URL+target, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if form != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		}
		signAWSRequest(req, body, accessKey, secretKey, "us-east-1", service, time.Date(2026, 9, 4, 7, 0, 0, 0, time.UTC))
		res, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	res := call("ec2", http.MethodPost, "/", url.Values{"Action": {"DescribeInstances"}, "Version": {"2016-11-15"}})
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !bytes.Contains(body, []byte("i-chimera0001")) || !bytes.Contains(body, []byte("vol-chimera0001")) {
		t.Fatalf("DescribeInstances status=%d body=%s", res.StatusCode, body)
	}

	res = call("ec2", http.MethodPost, "/", url.Values{"Action": {"StartInstances"}, "Version": {"2016-11-15"}, "InstanceId.1": {"i-chimera0001"}})
	body, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !bytes.Contains(body, []byte("running")) {
		t.Fatalf("StartInstances status=%d body=%s", res.StatusCode, body)
	}

	res = call("ec2", http.MethodPost, "/", url.Values{"Action": {"CreateSnapshot"}, "Version": {"2016-11-15"}, "VolumeId": {"vol-chimera0001"}, "Description": {"chimera e2e"}})
	body, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("CreateSnapshot status=%d body=%s", res.StatusCode, body)
	}
	snapID := regexp.MustCompile(`snap-chimera[0-9a-f]+`).FindString(string(body))
	if snapID == "" {
		t.Fatalf("snapshot id missing: %s", body)
	}

	res = call("ebs", http.MethodGet, "/snapshots/"+snapID+"/blocks", nil)
	var blocks struct {
		Blocks []struct {
			BlockIndex int64  `json:"BlockIndex"`
			BlockToken string `json:"BlockToken"`
		} `json:"Blocks"`
		BlockSize int `json:"BlockSize"`
	}
	if err := json.NewDecoder(res.Body).Decode(&blocks); err != nil {
		res.Body.Close()
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK || blocks.BlockSize != 512*1024 || len(blocks.Blocks) != 1 {
		t.Fatalf("ListSnapshotBlocks status=%d response=%+v", res.StatusCode, blocks)
	}

	target := fmt.Sprintf("/snapshots/%s/blocks/0?blockToken=%s", snapID, url.QueryEscape(blocks.Blocks[0].BlockToken))
	res = call("ebs", http.MethodGet, target, nil)
	block, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK || len(block) != 512*1024 || res.Header.Get("x-amz-Checksum-Algorithm") != "SHA256" {
		t.Fatalf("GetSnapshotBlock status=%d bytes=%d checksum=%q", res.StatusCode, len(block), res.Header.Get("x-amz-Checksum-Algorithm"))
	}
}

func TestAzureARMDiscoveryPowerAndManagedDiskAccessE2E(t *testing.T) {
	const subscription = "11111111-2222-3333-4444-555555555555"
	const token = "chimera-azure-token"
	ts := httptest.NewServer(azurepersona.New(subscription, token, 2))
	defer ts.Close()

	do := func(method, target string, body io.Reader) *http.Response {
		t.Helper()
		req, err := http.NewRequest(method, ts.URL+target, body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		res, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	base := "/subscriptions/" + subscription + "/resourceGroups/chimera-rg/providers/Microsoft.Compute"
	res := do(http.MethodGet, base+"/virtualMachines?api-version=2024-11-01", nil)
	var list struct {
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		res.Body.Close()
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK || len(list.Value) != 2 || list.Value[0].Name != "chimera-vm-01" {
		t.Fatalf("VM list status=%d response=%+v", res.StatusCode, list)
	}

	vmPath := base + "/virtualMachines/chimera-vm-01"
	res = do(http.MethodPost, vmPath+"/start?api-version=2024-11-01", nil)
	opURL := res.Header.Get("Azure-AsyncOperation")
	res.Body.Close()
	if res.StatusCode != http.StatusAccepted || opURL == "" {
		t.Fatalf("start status=%d operation=%q", res.StatusCode, opURL)
	}
	res = do(http.MethodGet, opURL, nil)
	var op map[string]any
	_ = json.NewDecoder(res.Body).Decode(&op)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || op["status"] != "Succeeded" {
		t.Fatalf("operation status=%d response=%v", res.StatusCode, op)
	}

	res = do(http.MethodGet, vmPath+"/instanceView?api-version=2024-11-01", nil)
	view, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !bytes.Contains(view, []byte("PowerState/running")) {
		t.Fatalf("instance view status=%d body=%s", res.StatusCode, view)
	}

	diskPath := base + "/disks/chimera-vm-01-osdisk"
	res = do(http.MethodPost, diskPath+"/beginGetAccess?api-version=2025-01-02", strings.NewReader(`{"access":"Read","durationInSeconds":3600}`))
	diskOp := res.Header.Get("Azure-AsyncOperation")
	res.Body.Close()
	if res.StatusCode != http.StatusAccepted || diskOp == "" {
		t.Fatalf("beginGetAccess status=%d operation=%q", res.StatusCode, diskOp)
	}
	res = do(http.MethodGet, diskOp, nil)
	var access struct {
		Properties struct {
			Output struct {
				AccessSAS string `json:"accessSAS"`
			} `json:"output"`
		} `json:"properties"`
	}
	if err := json.NewDecoder(res.Body).Decode(&access); err != nil {
		res.Body.Close()
		t.Fatal(err)
	}
	res.Body.Close()
	if access.Properties.Output.AccessSAS == "" {
		t.Fatal("missing disk accessSAS")
	}

	req, _ := http.NewRequest(http.MethodGet, access.Properties.Output.AccessSAS, nil)
	req.Header.Set("Range", "bytes=0-4095")
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	diskBytes, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusPartialContent || len(diskBytes) != 4096 || res.Header.Get("x-ms-blob-type") != "PageBlob" {
		t.Fatalf("disk range status=%d bytes=%d blobType=%q", res.StatusCode, len(diskBytes), res.Header.Get("x-ms-blob-type"))
	}
}

func signAWSRequest(req *http.Request, body []byte, accessKey, secretKey, region, service string, now time.Time) {
	amzDate := now.UTC().Format("20060102T150405Z")
	date := now.UTC().Format("20060102")
	req.Header.Set("X-Amz-Date", amzDate)
	payload := sha256.Sum256(body)
	payloadHash := hex.EncodeToString(payload[:])

	signed := []string{"host", "x-amz-date"}
	if req.Header.Get("Content-Type") != "" {
		signed = append([]string{"content-type"}, signed...)
	}
	canonicalHeaders := strings.Builder{}
	for _, h := range signed {
		value := ""
		if h == "host" {
			value = req.URL.Host
		} else {
			value = req.Header.Get(h)
		}
		canonicalHeaders.WriteString(h + ":" + strings.Join(strings.Fields(value), " ") + "\n")
	}
	signedHeaders := strings.Join(signed, ";")
	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalAWSQuery(req.URL.Query()),
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")
	hash := sha256.Sum256([]byte(canonicalRequest))
	scope := strings.Join([]string{date, region, service, "aws4_request"}, "/")
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(hash[:])
	kDate := hmacBytes([]byte("AWS4"+secretKey), date)
	kRegion := hmacBytes(kDate, region)
	kService := hmacBytes(kRegion, service)
	kSigning := hmacBytes(kService, "aws4_request")
	signature := hex.EncodeToString(hmacBytes(kSigning, stringToSign))
	req.Header.Set("Authorization", fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s", accessKey, scope, signedHeaders, signature))
}

func canonicalAWSQuery(v url.Values) string {
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
			parts = append(parts, awsQueryEncode(k)+"="+awsQueryEncode(val))
		}
	}
	return strings.Join(parts, "&")
}

func awsQueryEncode(v string) string {
	e := url.QueryEscape(v)
	e = strings.ReplaceAll(e, "+", "%20")
	e = strings.ReplaceAll(e, "%7E", "~")
	return e
}

func hmacBytes(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	_, _ = m.Write([]byte(data))
	return m.Sum(nil)
}
