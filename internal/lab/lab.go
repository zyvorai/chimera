package lab

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/zyvorai/chimera/internal/config"
	"github.com/zyvorai/chimera/internal/exportshim"
	"github.com/zyvorai/chimera/internal/faults"
	"github.com/zyvorai/chimera/internal/fixture"
	"github.com/zyvorai/chimera/internal/gateway"
)

type Lab struct {
	Config     config.Config
	Model      *simulator.Model
	Backend    *simulator.Server
	HTTPServer *http.Server
	Listener   net.Listener
	URL        *url.URL
	Faults     *faults.State
	Fixtures   *fixture.Store
}

func Start(ctx context.Context, cfg config.Config) (*Lab, error) {
	model := simulator.VPX()
	model.Datacenter = cfg.Datacenters
	model.Cluster = cfg.Clusters
	model.Host = cfg.HostsPerDC
	model.ClusterHost = cfg.HostsPerClus
	model.Datastore = cfg.Datastores
	model.Machine = cfg.VMsPerPool
	model.Autostart = cfg.AutostartVMs
	if cfg.SOAPDelay > 0 {
		model.DelayConfig.Delay = int(cfg.SOAPDelay / time.Millisecond)
	}
	if err := model.Create(); err != nil {
		return nil, fmt.Errorf("create vCenter model: %w", err)
	}

	model.Service.RegisterEndpoints = true
	model.Service.Listen = &url.URL{Host: "127.0.0.1:0", User: url.UserPassword(cfg.Username, cfg.Password)}
	backend := model.Service.NewServer()

	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		backend.Close()
		model.Remove()
		return nil, err
	}
	host := cfg.PublicHost
	if host == "" {
		host = ln.Addr().String()
		if strings.HasPrefix(host, "[::]") {
			host = strings.Replace(host, "[::]", "127.0.0.1", 1)
		}
		if strings.HasPrefix(host, "0.0.0.0") {
			host = strings.Replace(host, "0.0.0.0", "127.0.0.1", 1)
		}
	}
	scheme := "http"
	if cfg.TLS {
		scheme = "https"
	}
	publicBase := scheme + "://" + host
	publicURL, _ := url.Parse(publicBase + "/sdk")

	fixtures, err := fixture.New(cfg.FixtureVMDK, cfg.FixtureSizeMB)
	if err != nil {
		ln.Close()
		backend.Close()
		model.Remove()
		return nil, fmt.Errorf("prepare export fixture: %w", err)
	}
	shim := exportshim.New(fixtures, publicBase)
	registry := model.Service.Context.Map
	previousHandler := registry.Handler
	registry.Handler = func(c *simulator.Context, m *simulator.Method) (mo.Reference, types.BaseMethodFault) {
		h, fault := shim.Handler(c, m)
		if h != nil || fault != nil {
			return h, fault
		}
		if previousHandler != nil {
			return previousHandler(c, m)
		}
		return nil, nil
	}

	fs := faults.New()
	hosts := cfg.Datacenters * (cfg.HostsPerDC + cfg.Clusters*cfg.HostsPerClus)
	vmCount := cfg.Datacenters * max(1, cfg.Clusters) * cfg.VMsPerPool
	gw := gateway.New(backend.URL, publicBase, cfg.AdminToken, fs, fixtures, gateway.Meta{
		Version: "0.2.0", Persona: "vSphere", Username: cfg.Username, Datacenters: cfg.Datacenters,
		Clusters: cfg.Datacenters * cfg.Clusters, Hosts: hosts, Datastores: cfg.Datacenters * cfg.Datastores,
		VMs: vmCount, FixtureSizeMB: cfg.FixtureSizeMB, TLS: cfg.TLS,
	})
	hs := &http.Server{Handler: gw, ReadHeaderTimeout: 15 * time.Second}

	serveLn := ln
	if cfg.TLS {
		cert, err := selfSigned(host)
		if err != nil {
			ln.Close()
			_ = fixtures.Close()
			backend.Close()
			model.Remove()
			return nil, err
		}
		serveLn = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
	}
	errCh := make(chan error, 1)
	go func() { errCh <- hs.Serve(serveLn) }()
	go func() {
		select {
		case <-ctx.Done():
			shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = hs.Shutdown(shutCtx)
		case err := <-errCh:
			if err != nil && err != http.ErrServerClosed {
				log.Printf("public server stopped: %v", err)
			}
		}
	}()

	return &Lab{Config: cfg, Model: model, Backend: backend, HTTPServer: hs, Listener: ln, URL: publicURL, Faults: fs, Fixtures: fixtures}, nil
}

func (l *Lab) Close(ctx context.Context) error {
	var first error
	if l.HTTPServer != nil {
		if err := l.HTTPServer.Shutdown(ctx); err != nil && first == nil {
			first = err
		}
	}
	if l.Backend != nil {
		l.Backend.Close()
	}
	if l.Fixtures != nil {
		if err := l.Fixtures.Close(); err != nil && first == nil {
			first = err
		}
	}
	if l.Model != nil {
		l.Model.Remove()
	}
	return first
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func selfSigned(hostport string) (tls.Certificate, error) {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = hostport
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	tmpl := x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: "chimera"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(365 * 24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	if ip := net.ParseIP(host); ip != nil {
		tmpl.IPAddresses = []net.IP{ip}
	} else {
		tmpl.DNSNames = []string{host, "localhost"}
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return tls.X509KeyPair(certPEM, keyPEM)
}
