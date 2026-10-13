package cmd

import (
	"context"
	"errors"
	"os"

	pb "github.com/infinimesh/proto/node"
	auditpb "github.com/infinimesh/proto/node/audit"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func makeAuditClient(ctx context.Context) (pb.AuditServiceClient, error) {
	conn, err := makeConnection(ctx)
	if err != nil {
		return nil, err
	}
	return pb.NewAuditServiceClient(conn), nil
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "List audit records",
	RunE: func(cmd *cobra.Command, args []string) error {
		device, _ := cmd.Flags().GetString("device")
		account, _ := cmd.Flags().GetString("account")
		ns, _ := cmd.Flags().GetString("namespace")
		if device == "" && account == "" && ns == "" {
			return errors.New("one of --device, --account, or --namespace is required")
		}
		ctx := makeContextWithBearerToken()
		client, err := makeAuditClient(ctx)
		if err != nil {
			return err
		}
		req := &auditpb.ListAuditRequest{Device: device, Account: account, Namespace: ns}
		var records []*auditpb.AuditRecord
		for {
			res, err := client.List(ctx, req)
			if err != nil {
				return err
			}
			records = append(records, res.GetRecords()...)
			if req.PageToken = res.GetNextPageToken(); req.PageToken == "" {
				if printJson, _ := cmd.Flags().GetBool("json"); printJson {
					return printJsonResponse(&auditpb.AuditRecords{Records: records})
				}
				break
			}
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"At", "Action", "Account", "Device", "Namespace", "Token"})
		for _, row := range records {
			at := ""
			if row.GetAt() != nil {
				at = row.GetAt().AsTime().Format("2006-01-02T15:04:05Z07:00")
			}
			t.AppendRow(table.Row{at, row.GetAction(), row.GetAccount(), row.GetDevice(), row.GetNamespace(), row.GetTokenId()})
		}
		t.Render()
		return nil
	},
}

func init() {
	auditCmd.Flags().String("device", "", "Device uuid")
	auditCmd.Flags().String("account", "", "Account uuid")
	auditCmd.Flags().String("namespace", "", "Namespace uuid")
	rootCmd.AddCommand(auditCmd)
}
