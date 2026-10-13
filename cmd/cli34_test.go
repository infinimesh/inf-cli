package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestDeviceCommandsRequireAnID(t *testing.T) {
	if err := getDeviceCmd.RunE(getDeviceCmd, nil); err == nil || !strings.Contains(err.Error(), "device is required") {
		t.Fatalf("get: %v", err)
	}
	if err := makeDeviceTokenCmd.RunE(makeDeviceTokenCmd, nil); err == nil || !strings.Contains(err.Error(), "device is required") {
		t.Fatalf("token: %v", err)
	}
	if err := mgmtDeviceStateCmd.RunE(mgmtDeviceStateCmd, nil); err == nil || !strings.Contains(err.Error(), "device is required") {
		t.Fatalf("state: %v", err)
	}
	if err := getAccountCmd.RunE(getAccountCmd, nil); err == nil || !strings.Contains(err.Error(), "account is required") {
		t.Fatalf("account: %v", err)
	}
}

func TestDevIsTheKit(t *testing.T) {
	for _, alias := range devicesCmd.Aliases {
		if alias == "dev" {
			t.Fatal("devices still aliases dev")
		}
	}
	cmd, _, err := rootCmd.Find([]string{"dev"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != devCmd {
		t.Fatalf("inf dev resolved to %s", cmd.Name())
	}
}

func TestTLSSkipVerifyFollowsTheFlag(t *testing.T) {
	viper.Reset()
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	cfg, err := tlsConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.InsecureSkipVerify {
		t.Fatal("verification is off by default")
	}
	viper.Set("insecure_skip_verify", true)
	cfg, err = tlsConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.InsecureSkipVerify {
		t.Fatal("stored insecure_skip_verify did not apply")
	}
	viper.Set("insecure_skip_verify", false)
}

func TestDryRunCompilesCEL(t *testing.T) {
	if err := dryRunRuleCmd.Flags().Set("when", ""); err != nil {
		t.Fatal(err)
	}
	if err := dryRunRuleCmd.RunE(dryRunRuleCmd, nil); err == nil {
		t.Fatal("missing when was accepted")
	}
	if err := dryRunRuleCmd.Flags().Set("when", "reported.temp > 80"); err != nil {
		t.Fatal(err)
	}
	if err := dryRunRuleCmd.Flags().Set("clear", ""); err != nil {
		t.Fatal(err)
	}
	if err := dryRunRuleCmd.RunE(dryRunRuleCmd, nil); err != nil {
		t.Fatal(err)
	}
	if err := dryRunRuleCmd.Flags().Set("when", "nope("); err != nil {
		t.Fatal(err)
	}
	if err := dryRunRuleCmd.RunE(dryRunRuleCmd, nil); err == nil {
		t.Fatal("bad expression compiled")
	}
}

func TestAssignAndAuditRequireFilters(t *testing.T) {
	if err := assignOTACmd.RunE(assignOTACmd, nil); err == nil || !strings.Contains(err.Error(), "--namespace") {
		t.Fatalf("ota: %v", err)
	}
	if err := auditCmd.RunE(auditCmd, nil); err == nil || !strings.Contains(err.Error(), "--device") {
		t.Fatalf("audit: %v", err)
	}
}

func TestYamlTemplateBecomesJSON(t *testing.T) {
	b, err := yamlToJSON([]byte("title: pump\nenabled: true\ntags:\n  - plant\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"title":"pump"`) || !strings.Contains(string(b), `"plant"`) {
		t.Fatalf("%s", b)
	}
}
