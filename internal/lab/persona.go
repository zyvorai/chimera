// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package lab

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zyvorai/chimera/internal/config"
	"github.com/zyvorai/chimera/internal/personas/hyperv"
	"github.com/zyvorai/chimera/internal/personas/nutanix"
)

// StartHTTPPersona starts the Nutanix Prism or Hyper-V WS-Man persona.
// vSphere continues to use the existing govmomi-backed Start path.
func StartHTTPPersona(ctx context.Context, cfg config.Config) (*Lab, error) {
	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return nil, err
	}
	host := cfg.PublicHost
	if host == "" {
		host = ln.Addr().String()
		host = strings.Replace(host, "[::]", "127.0.0.1", 1)
		host = strings.Replace(host, "0.0.0.0", "127.0.0.1", 1)
	}
	scheme := "http"
	if cfg.TLS {
		scheme = "https"
	}

	var h http.Handler
	var endpoint string
	switch strings.ToLower(cfg.Persona) {
	case "nutanix":
		h = nutanix.New(cfg.Username, cfg.Password, cfg.VMsPerPool)
		endpoint = "/api/nutanix/v3"
	case "hyperv":
		h = hyperv.New(cfg.Username, cfg.Password, cfg.VMsPerPool)
		endpoint = "/wsman"
	default:
		ln.Close()
		return nil, fmt.Errorf("unsupported HTTP persona %q", cfg.Persona)
	}
	hs := &http.Server{Handler: h, ReadHeaderTimeout: 15 * time.Second}
	serveLn := ln
	if cfg.TLS {
		cert, err := selfSigned(host)
		if err != nil {
			ln.Close()
			return nil, err
		}
		serveLn = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
	}
	go func() { _ = hs.Serve(serveLn) }()
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = hs.Shutdown(c)
	}()
	u, _ := url.Parse(scheme + "://" + host + endpoint)
	return &Lab{Config: cfg, HTTPServer: hs, Listener: ln, URL: u}, nil
}
