package cmd

import (
	"context"
	"net"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/infinimesh/proto/node"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type mockAccountsServer struct {
	pb.UnimplementedAccountsServiceServer
	last *pb.TokenRequest
}

func (s *mockAccountsServer) Token(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	s.last = req
	return &pb.TokenResponse{
		Token: "token-" + req.Auth.Data[0],
	}, nil
}

func startAccounts(t *testing.T) *mockAccountsServer {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	impl := &mockAccountsServer{}
	srv := grpc.NewServer()
	pb.RegisterAccountsServiceServer(srv, impl)
	go srv.Serve(lis)
	t.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})
	t.Cleanup(func() {
		_ = loginCmd.Flags().Set("api", "")
		_ = loginCmd.Flags().Set("username", "")
		_ = loginCmd.Flags().Set("password", "")
		_ = loginCmd.Flags().Set("root", "false")
		_ = rootCmd.PersistentFlags().Set("insecure", "false")
		_ = rootCmd.PersistentFlags().Set("insecure-skip-verify", "false")
		_ = rootCmd.PersistentFlags().Set("ca-file", "")
		rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	})
	if err := loginCmd.Flags().Set("api", lis.Addr().String()); err != nil {
		t.Fatal(err)
	}
	if err := loginCmd.Flags().Set("username", "root"); err != nil {
		t.Fatal(err)
	}
	if err := loginCmd.Flags().Set("password", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := rootCmd.PersistentFlags().Set("insecure", "true"); err != nil {
		t.Fatal(err)
	}
	return impl
}

func TestLoginStoresContext(t *testing.T) {
	tempDir := t.TempDir()
	cfg := filepath.Join(tempDir, "config.yaml")

	// isolate viper/config for the test
	initialized = false
	cfgFile = cfg
	viper.Reset()
	initConfig()

	impl := startAccounts(t)
	host := loginCmd.Flags().Lookup("api").Value.String()

	if err := loginCmd.RunE(loginCmd, nil); err != nil {
		t.Fatalf("login run: %v", err)
	}
	if impl.last.GetInf() || impl.last.GetInfinite() {
		t.Fatal("plain login requested a root token")
	}

	if got := viper.GetString("infinimesh"); got != host {
		t.Fatalf("expected host %s, got %s", host, got)
	}
	if got := viper.GetString("token"); got != "token-root" {
		t.Fatalf("expected token token-root, got %s", got)
	}
	if !viper.GetBool("insecure") {
		t.Fatalf("expected insecure flag to persist")
	}
}

func TestLoginCAFileClearsSkipVerify(t *testing.T) {
	tempDir := t.TempDir()
	cfg := filepath.Join(tempDir, "config.yaml")
	initialized = false
	cfgFile = cfg
	viper.Reset()
	initConfig()
	viper.Set("insecure_skip_verify", true)

	startAccounts(t)
	if err := rootCmd.PersistentFlags().Set("ca-file", "ca.pem"); err != nil {
		t.Fatal(err)
	}
	if err := loginCmd.RunE(loginCmd, nil); err != nil {
		t.Fatal(err)
	}
	if viper.GetBool("insecure_skip_verify") {
		t.Fatal("ca-file login kept insecure_skip_verify")
	}
	if got := viper.GetString("ca_file"); !filepath.IsAbs(got) || !strings.HasSuffix(got, "ca.pem") {
		t.Fatalf("ca_file = %q", got)
	}
}

func TestLoginRootSetsBothFlags(t *testing.T) {
	tempDir := t.TempDir()
	cfg := filepath.Join(tempDir, "config.yaml")
	initialized = false
	cfgFile = cfg
	viper.Reset()
	initConfig()

	impl := startAccounts(t)
	if err := loginCmd.Flags().Set("root", "true"); err != nil {
		t.Fatal(err)
	}
	if err := loginCmd.RunE(loginCmd, nil); err != nil {
		t.Fatal(err)
	}
	if !impl.last.GetInf() || !impl.last.GetInfinite() {
		t.Fatalf("inf=%v infinite=%v", impl.last.GetInf(), impl.last.GetInfinite())
	}
}

func TestNormalizeAPIHost(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"api.infinimesh.local", "api.infinimesh.local:443"},
		{"api.infinimesh.local:9000", "api.infinimesh.local:9000"},
		{"http://api.infinimesh.local", "api.infinimesh.local:443"},
		{"http://api.infinimesh.local:8080", "api.infinimesh.local:8080"},
	}
	for _, tt := range tests {
		got, err := normalizeAPIHost(tt.in)
		if err != nil {
			t.Fatalf("normalizeAPIHost(%q) returned error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("normalizeAPIHost(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeAPIHostRejectsBrokenHost(t *testing.T) {
	for _, in := range []string{"[", ":", "/"} {
		if _, err := normalizeAPIHost(in); err == nil {
			t.Fatalf("normalizeAPIHost(%q) returned a host", in)
		}
	}
}
