package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rulespb "github.com/infinimesh/proto/node/rules"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
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

func TestRotatePluginRequiresUUID(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"plugins", "rotate"})
	if err != nil || cmd == nil || cmd.Name() != "rotate" {
		t.Fatalf("rotate: %v %v", cmd, err)
	}
	if err := cmd.RunE(cmd, nil); err == nil || !strings.Contains(err.Error(), "plugin is required") {
		t.Fatalf("%v", err)
	}
}

func TestOTAProviderUsesRequest(t *testing.T) {
	if otaProviderUsesRequest("none") || otaProviderUsesRequest("") || otaProviderUsesRequest("other") {
		t.Fatal("only hawkbit, mender, and webhook request")
	}
	for _, provider := range []string{"hawkbit", "mender", "webhook"} {
		if !otaProviderUsesRequest(provider) {
			t.Fatal(provider)
		}
	}
}

func TestAssignAndAuditRequireFilters(t *testing.T) {
	_ = assignOTACmd.Flags().Set("namespace", "")
	_ = assignOTACmd.Flags().Set("version", "")
	_ = assignOTACmd.Flags().Set("tag", "")
	_ = assignOTACmd.Flags().Set("device", "")
	if err := assignOTACmd.RunE(assignOTACmd, nil); err == nil || !strings.Contains(err.Error(), "--device") {
		t.Fatalf("ota: %v", err)
	}
	if err := assignOTACmd.Flags().Set("version", "1.2.3"); err != nil {
		t.Fatal(err)
	}
	if err := assignOTACmd.Flags().Set("device", "dev-1"); err != nil {
		t.Fatal(err)
	}
	viper.Set("infinimesh", "")
	err := assignOTACmd.RunE(assignOTACmd, nil)
	if err == nil || strings.Contains(err.Error(), "--version and one of") {
		t.Fatalf("device assign was rejected: %v", err)
	}
	if err := auditCmd.RunE(auditCmd, nil); err == nil || !strings.Contains(err.Error(), "--device") {
		t.Fatalf("audit: %v", err)
	}
}

func TestFormatRulePrintsClearCooldownRepeat(t *testing.T) {
	out := formatRule(&rulespb.Rule{Uuid: "u", When: "w", Clear: "c", Cooldown: "10m", Repeat: "1h", OnClear: true})
	for _, part := range []string{"clear: c", "cooldown: 10m", "repeat: 1h", "on_clear: true"} {
		if !strings.Contains(out, part) {
			t.Fatalf("missing %q in %q", part, out)
		}
	}
}

func TestTokenFlagOverridesStored(t *testing.T) {
	viper.Set("token", "stored")
	if err := rootCmd.PersistentFlags().Set("token", "override"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = rootCmd.PersistentFlags().Set("token", "")
		rootCmd.PersistentFlags().Lookup("token").Changed = false
		viper.Set("token", "")
	})
	ctx := makeContextWithBearerToken()
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(md.Get("authorization")) != 1 || md.Get("authorization")[0] != "Bearer override" {
		t.Fatalf("%v", md)
	}
}

func TestLoginErrorIsNotPrintedByCobra(t *testing.T) {
	if !rootCmd.SilenceErrors {
		t.Fatal("cobra prints the error and CheckErr prints it again")
	}
	initialized = false
	cfgFile = filepath.Join(t.TempDir(), "config.yaml")
	viper.Reset()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"login"})
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		initialized = false
	})
	err := rootCmd.Execute()
	if err == nil || strings.Contains(buf.String(), err.Error()) {
		t.Fatalf("cobra printed %q (%v)", buf.String(), err)
	}
}

func TestNamespaceShortFlag(t *testing.T) {
	for _, cmd := range []*cobra.Command{createDeviceCmd, createRuleCmd, listAlarmsCmd} {
		flag := cmd.Flags().ShorthandLookup("n")
		if flag == nil || flag.Name != "namespace" {
			t.Fatalf("%s -n is %v", cmd.CommandPath(), flag)
		}
	}
}

func TestInitConfigCreatesMissingFolder(t *testing.T) {
	prevInit, prevCfg := initialized, cfgFile
	dir := filepath.Join(t.TempDir(), "missing")
	initialized = false
	cfgFile = filepath.Join(dir, "config.yaml")
	viper.Reset()
	t.Cleanup(func() {
		initialized = prevInit
		cfgFile = prevCfg
		viper.Reset()
	})
	initConfig()
	info, err := os.Stat(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		t.Fatal("config path is a directory")
	}
}

func TestMetricsRequiresAddr(t *testing.T) {
	if err := metricsCmd.Flags().Set("addr", ""); err != nil {
		t.Fatal(err)
	}
	err := metricsCmd.RunE(metricsCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--addr") {
		t.Fatalf("%v", err)
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
