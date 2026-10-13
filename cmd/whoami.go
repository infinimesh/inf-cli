package cmd

import (
	"fmt"

	accpb "github.com/infinimesh/proto/node/accounts"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Print the account for the stored token",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := makeContextWithBearerToken()
		client, err := makeAccountsServiceClient(ctx)
		if err != nil {
			return err
		}
		r, err := client.Get(ctx, &accpb.Account{Uuid: "me"})
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(r)
		}
		fmt.Printf("%s\t%s\n", r.GetUuid(), r.GetTitle())
		return nil
	},
}
