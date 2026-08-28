// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package config

import "testing"

func TestDefaultValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultListensOnAllInterfaces(t *testing.T) {
	if got := Default().Listen; got != "0.0.0.0:8989" {
		t.Fatalf("Listen=%q, want 0.0.0.0:8989", got)
	}
}

func TestDefaultAdminCredentials(t *testing.T) {
	cfg := Default()
	if cfg.AdminUsername != "admin" || cfg.AdminPassword != "admin" {
		t.Fatalf("AdminUsername/AdminPassword=%q/%q, want admin/admin", cfg.AdminUsername, cfg.AdminPassword)
	}
}

func TestValidateRejectsBlankAdminCredentials(t *testing.T) {
	cfg := Default()
	cfg.AdminUsername = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for blank admin_username")
	}
	cfg = Default()
	cfg.AdminPassword = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for blank admin_password")
	}
}
