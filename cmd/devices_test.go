package cmd

import (
	"context"
	"net"
	"testing"

	pb "github.com/infinimesh/proto/node"
	"github.com/infinimesh/proto/node/access"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type mockDevicesServer struct {
	pb.UnimplementedDevicesServiceServer
	last map[string]access.Level
}

func (s *mockDevicesServer) MakeDevicesToken(ctx context.Context, req *pb.DevicesTokenRequest) (*pb.TokenResponse, error) {
	s.last = req.Devices
	return &pb.TokenResponse{Token: "ok"}, nil
}

func TestMakeDevicesTokenUsesAccessMap(t *testing.T) {
	viper.Reset()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	mock := &mockDevicesServer{}
	pb.RegisterDevicesServiceServer(srv, mock)
	go srv.Serve(lis)
	t.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})

	viper.Set("infinimesh", lis.Addr().String())
	viper.Set("insecure", true)
	viper.Set("token", "test-token")

	if err := makeDeviceTokenCmd.RunE(makeDeviceTokenCmd, []string{"dev-one", "dev-two"}); err != nil {
		t.Fatalf("run make device token: %v", err)
	}

	if mock.last == nil {
		t.Fatalf("server did not receive request")
	}
	if len(mock.last) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(mock.last))
	}
	if mock.last["dev-one"] != access.Level_NONE {
		t.Fatalf("expected dev-one to default to NONE, got %v", mock.last["dev-one"])
	}
	if mock.last["dev-two"] != access.Level_NONE {
		t.Fatalf("expected dev-two to default to NONE, got %v", mock.last["dev-two"])
	}
}

func TestStateTokenRequest(t *testing.T) {
	read := stateTokenRequest([]string{"dev"}, "", "", "")
	if read.Post || read.Devices["dev"] != access.Level_READ {
		t.Fatalf("get token = post %v level %v", read.Post, read.Devices["dev"])
	}

	reported := stateTokenRequest([]string{"dev"}, "", `{"t":1}`, "")
	if !reported.Post || reported.Devices["dev"] != access.Level_READ {
		t.Fatalf("reported token = post %v level %v", reported.Post, reported.Devices["dev"])
	}

	desired := stateTokenRequest([]string{"dev"}, `{"t":1}`, "", "")
	if desired.Post || desired.Devices["dev"] != access.Level_MGMT {
		t.Fatalf("desired token = post %v level %v", desired.Post, desired.Devices["dev"])
	}

	both := stateTokenRequest([]string{"dev"}, `{"t":1}`, `{"t":1}`, "")
	if !both.Post || both.Devices["dev"] != access.Level_MGMT {
		t.Fatalf("patch and report token = post %v level %v", both.Post, both.Devices["dev"])
	}

	dropReported := stateTokenRequest([]string{"dev"}, "", "", "reported.temp")
	if !dropReported.Post || dropReported.Devices["dev"] != access.Level_READ {
		t.Fatalf("remove reported token = post %v level %v", dropReported.Post, dropReported.Devices["dev"])
	}

	dropDesired := stateTokenRequest([]string{"dev"}, "", "", "desired.temp")
	if dropDesired.Post || dropDesired.Devices["dev"] != access.Level_MGMT {
		t.Fatalf("remove desired token = post %v level %v", dropDesired.Post, dropDesired.Devices["dev"])
	}
}
