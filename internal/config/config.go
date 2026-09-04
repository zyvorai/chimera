// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Persona      string        `json:"persona"`
	Listen       string        `json:"listen"`
	PublicHost   string        `json:"public_host"`
	TLS          bool          `json:"tls"`
	Username     string        `json:"username"`
	Password     string        `json:"password"`
	Datacenters  int           `json:"datacenters"`
	Clusters     int           `json:"clusters"`
	HostsPerDC   int           `json:"hosts_per_dc"`
	HostsPerClus int           `json:"hosts_per_cluster"`
	Datastores   int           `json:"datastores"`
	VMsPerPool   int           `json:"vms_per_pool"`
	AutostartVMs bool          `json:"autostart_vms"`
	SOAPDelay    time.Duration `json:"-"`
	SOAPDelayMS  int           `json:"soap_delay_ms"`
	AdminToken   string        `json:"admin_token"`
	// AdminUsername/AdminPassword gate the Chimera dashboard's own login —
	// distinct from Username/Password above, which are the simulated
	// vCenter login used by API clients like govc/hyperexport.
	AdminUsername  string `json:"admin_username"`
	AdminPassword  string `json:"admin_password"`
	FixtureVMDK    string `json:"fixture_vmdk"`
	FixtureVMDKDir string `json:"fixture_vmdk_dir"`
	// FixtureVMDKDirs are additional, read-only directories scanned the same
	// way as FixtureVMDKDir — browser uploads always land in FixtureVMDKDir,
	// these only ever contribute files an operator staged directly on disk.
	FixtureVMDKDirs []string `json:"fixture_vmdk_dirs"`
	FixtureSizeMB   int      `json:"fixture_size_mb"`
}

func Default() Config {
	return Config{
		Persona:        "vsphere",
		Listen:         "0.0.0.0:8989",
		PublicHost:     "",
		TLS:            false,
		Username:       "administrator@vsphere.local",
		Password:       "vmware",
		Datacenters:    1,
		Clusters:       1,
		HostsPerDC:     0,
		HostsPerClus:   1,
		Datastores:     1,
		VMsPerPool:     3,
		AutostartVMs:   false,
		SOAPDelayMS:    0,
		AdminToken:     "chimera-admin",
		AdminUsername:  "admin",
		AdminPassword:  "admin",
		FixtureVMDK:    "",
		FixtureVMDKDir: "",
		FixtureSizeMB:  16,
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return cfg, err
		}
		if err := json.Unmarshal(b, &cfg); err != nil {
			return cfg, fmt.Errorf("decode config: %w", err)
		}
	}
	applyEnv(&cfg)
	cfg.SOAPDelay = time.Duration(cfg.SOAPDelayMS) * time.Millisecond
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	switch strings.ToLower(strings.TrimSpace(c.Persona)) {
	case "vsphere", "nutanix", "hyperv":
	default:
		return fmt.Errorf("persona must be one of vsphere, nutanix, hyperv; got %q", c.Persona)
	}
	if strings.TrimSpace(c.Listen) == "" {
		return errors.New("listen cannot be empty")
	}
	if c.Username == "" || c.Password == "" {
		return errors.New("username and password are required")
	}
	if c.AdminUsername == "" || c.AdminPassword == "" {
		return errors.New("admin_username and admin_password are required")
	}
	if c.Datacenters < 1 || c.Datastores < 1 || c.VMsPerPool < 1 {
		return errors.New("datacenters, datastores, and vms_per_pool must be >= 1")
	}
	if c.Clusters < 0 || c.HostsPerDC < 0 || c.HostsPerClus < 0 {
		return errors.New("cluster and host counts cannot be negative")
	}
	if c.FixtureSizeMB < 1 {
		return errors.New("fixture_size_mb must be >= 1")
	}
	if strings.TrimSpace(c.FixtureVMDK) != "" && strings.TrimSpace(c.FixtureVMDKDir) != "" {
		return errors.New("fixture_vmdk and fixture_vmdk_dir are mutually exclusive; set only one")
	}
	if strings.TrimSpace(c.FixtureVMDK) != "" && len(c.FixtureVMDKDirs) > 0 {
		return errors.New("fixture_vmdk and fixture_vmdk_dirs are mutually exclusive; set only one")
	}
	for _, d := range c.FixtureVMDKDirs {
		if strings.TrimSpace(d) == "" {
			return errors.New("fixture_vmdk_dirs entries cannot be empty")
		}
	}
	return nil
}

func applyEnv(c *Config) {
	str := func(key string, dst *string) {
		if v := os.Getenv(key); v != "" {
			*dst = v
		}
	}
	boolean := func(key string, dst *bool) {
		if v := os.Getenv(key); v != "" {
			if x, err := strconv.ParseBool(v); err == nil {
				*dst = x
			}
		}
	}
	integer := func(key string, dst *int) {
		if v := os.Getenv(key); v != "" {
			if x, err := strconv.Atoi(v); err == nil {
				*dst = x
			}
		}
	}
	stringList := func(key string, dst *[]string) {
		if v := os.Getenv(key); v != "" {
			var out []string
			for part := range strings.SplitSeq(v, ",") {
				if part = strings.TrimSpace(part); part != "" {
					out = append(out, part)
				}
			}
			*dst = out
		}
	}
	str("CHIMERA_PERSONA", &c.Persona)
	str("CHIMERA_LISTEN", &c.Listen)
	str("CHIMERA_PUBLIC_HOST", &c.PublicHost)
	str("CHIMERA_USERNAME", &c.Username)
	str("CHIMERA_PASSWORD", &c.Password)
	str("CHIMERA_ADMIN_TOKEN", &c.AdminToken)
	str("CHIMERA_ADMIN_USERNAME", &c.AdminUsername)
	str("CHIMERA_ADMIN_PASSWORD", &c.AdminPassword)
	str("CHIMERA_FIXTURE_VMDK", &c.FixtureVMDK)
	str("CHIMERA_FIXTURE_VMDK_DIR", &c.FixtureVMDKDir)
	stringList("CHIMERA_FIXTURE_VMDK_DIRS", &c.FixtureVMDKDirs)
	boolean("CHIMERA_TLS", &c.TLS)
	boolean("CHIMERA_AUTOSTART_VMS", &c.AutostartVMs)
	integer("CHIMERA_DATACENTERS", &c.Datacenters)
	integer("CHIMERA_CLUSTERS", &c.Clusters)
	integer("CHIMERA_HOSTS_PER_DC", &c.HostsPerDC)
	integer("CHIMERA_HOSTS_PER_CLUSTER", &c.HostsPerClus)
	integer("CHIMERA_DATASTORES", &c.Datastores)
	integer("CHIMERA_VMS_PER_POOL", &c.VMsPerPool)
	integer("CHIMERA_SOAP_DELAY_MS", &c.SOAPDelayMS)
	integer("CHIMERA_FIXTURE_SIZE_MB", &c.FixtureSizeMB)
}
