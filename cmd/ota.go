package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	pb "github.com/infinimesh/proto/node"
	devpb "github.com/infinimesh/proto/node/devices"
	nspb "github.com/infinimesh/proto/node/namespaces"
	shadowpb "github.com/infinimesh/proto/shadow"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func makeNodeShadowClient(ctx context.Context) (pb.ShadowServiceClient, error) {
	conn, err := makeConnection(ctx)
	if err != nil {
		return nil, err
	}
	return pb.NewShadowServiceClient(conn), nil
}

var otaCmd = &cobra.Command{
	Use:   "ota",
	Short: "Assign firmware",
	Long: `Bulk assign writes desired.ota for every selected device.

The call needs an account token or an addon token. A device token is rejected.
The namespace must have OTA enabled and a registered addon whose provider matches.
When the namespace provider is hawkbit, mender, or webhook, assign stores requested_version
instead of version. The addon writes version.`,
}

var assignOTACmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a firmware version",
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		version, _ := cmd.Flags().GetString("version")
		tag, _ := cmd.Flags().GetString("tag")
		devices, _ := cmd.Flags().GetStringSlice("device")
		if version == "" || (ns == "" && tag == "" && len(devices) == 0) {
			return errors.New("--version and one of --namespace, --tag, or --device are required")
		}
		ctx := makeContextWithBearerToken()
		client, err := makeNodeShadowClient(ctx)
		if err != nil {
			return err
		}
		req := &shadowpb.AssignOTARequest{
			Namespace: ns,
			Tag:       tag,
			Devices:   devices,
			Version:   version,
		}
		useRequest, err := otaAssignUsesRequest(ctx, ns, devices)
		if err != nil {
			return err
		}
		var res *shadowpb.AssignOTAResponse
		if useRequest {
			res, err = client.RequestOTA(ctx, req)
		} else {
			res, err = client.AssignOTA(ctx, req)
		}
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(res)
		}
		fmt.Printf("updated %d\n", res.GetUpdated())
		if len(res.GetResults()) == 0 {
			return nil
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Device", "Error"})
		for _, row := range res.GetResults() {
			t.AppendRow(table.Row{row.GetDevice(), row.GetError()})
		}
		t.Render()
		return nil
	},
}

func otaProviderUsesRequest(provider string) bool {
	switch strings.TrimSpace(provider) {
	case "hawkbit", "mender", "webhook":
		return true
	default:
		return false
	}
}

func otaAssignUsesRequest(ctx context.Context, namespace string, devices []string) (bool, error) {
	if namespace != "" {
		client, err := makeNamespacesServiceClient(ctx)
		if err != nil {
			return false, err
		}
		ns, err := client.Get(ctx, &nspb.Namespace{Uuid: namespace})
		if err != nil {
			return false, err
		}
		return otaProviderUsesRequest(ns.GetOta().GetProvider()), nil
	}
	if len(devices) == 0 {
		return false, nil
	}
	devClient, err := makeDevicesServiceClient(ctx)
	if err != nil {
		return false, err
	}
	nsClient, err := makeNamespacesServiceClient(ctx)
	if err != nil {
		return false, err
	}
	seen := map[string]bool{}
	var decided, request bool
	for _, id := range devices {
		dev, err := devClient.Get(ctx, &devpb.Device{Uuid: id})
		if err != nil {
			return false, err
		}
		nsID := dev.GetAccess().GetNamespace()
		if nsID == "" {
			return false, fmt.Errorf("device %s has no namespace", id)
		}
		if seen[nsID] {
			continue
		}
		seen[nsID] = true
		ns, err := nsClient.Get(ctx, &nspb.Namespace{Uuid: nsID})
		if err != nil {
			return false, err
		}
		this := otaProviderUsesRequest(ns.GetOta().GetProvider())
		if !decided {
			request = this
			decided = true
			continue
		}
		if this != request {
			return false, errors.New("selected devices use more than one OTA write path; pass --namespace")
		}
	}
	return request, nil
}

func init() {
	assignOTACmd.Flags().String("namespace", "", "Namespace")
	assignOTACmd.Flags().String("version", "", "Firmware version")
	assignOTACmd.Flags().String("tag", "", "Only devices with this tag")
	assignOTACmd.Flags().StringSlice("device", nil, "Device uuid (repeatable). With no namespace or tag, these are the devices to assign")
	otaCmd.AddCommand(assignOTACmd)
	rootCmd.AddCommand(otaCmd)
}
