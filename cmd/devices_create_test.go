package cmd

import (
	"strings"
	"testing"
)

func TestCreateDeviceWithoutTemplateReturnsError(t *testing.T) {
	cmd := createDeviceCmd
	if err := cmd.Flags().Set("crt", "cert.pem"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Flags().Set("crt", "") })

	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Fatal("Args should reject a missing template")
	}
	err := cmd.RunE(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "device template") {
		t.Fatalf("want template error, got %v", err)
	}
}

func TestCreateDeviceWithoutCertReturnsError(t *testing.T) {
	cmd := createDeviceCmd
	_ = cmd.Flags().Set("crt", "")
	err := cmd.RunE(cmd, []string{"device.json"})
	if err == nil || !strings.Contains(err.Error(), "--crt") {
		t.Fatalf("want --crt error, got %v", err)
	}
}
