package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	pb "github.com/infinimesh/proto/node"
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
The namespace must have OTA enabled and a registered addon whose provider matches.`,
}

var assignOTACmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a firmware version",
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		version, _ := cmd.Flags().GetString("version")
		tag, _ := cmd.Flags().GetString("tag")
		devices, _ := cmd.Flags().GetStringSlice("device")
		if ns == "" || version == "" {
			return errors.New("--namespace and --version are required")
		}
		ctx := makeContextWithBearerToken()
		client, err := makeNodeShadowClient(ctx)
		if err != nil {
			return err
		}
		res, err := client.AssignOTA(ctx, &shadowpb.AssignOTARequest{
			Namespace: ns,
			Tag:       tag,
			Devices:   devices,
			Version:   version,
		})
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

func init() {
	assignOTACmd.Flags().String("namespace", "", "Namespace")
	assignOTACmd.Flags().String("version", "", "Firmware version")
	assignOTACmd.Flags().String("tag", "", "Only devices with this tag")
	assignOTACmd.Flags().StringSlice("device", nil, "Device uuid (repeatable). Empty means every device in the namespace")
	otaCmd.AddCommand(assignOTACmd)
	rootCmd.AddCommand(otaCmd)
}
