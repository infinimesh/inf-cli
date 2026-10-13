package cmd

import (
	"errors"
	"fmt"

	nspb "github.com/infinimesh/proto/node/namespaces"
	"github.com/spf13/cobra"
)

func namespaceID(args []string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", errors.New("namespace is required")
	}
	return args[0], nil
}

func fillNamespace(cmd *cobra.Command, ns *nspb.Namespace) {
	if cmd.Flags().Changed("title") {
		ns.Title, _ = cmd.Flags().GetString("title")
	}
	if cmd.Flags().Changed("provider") || cmd.Flags().Changed("ota") || cmd.Flags().Changed("webhook-url") {
		if ns.Ota == nil {
			ns.Ota = &nspb.OTAConfig{}
		}
		if cmd.Flags().Changed("provider") {
			ns.Ota.Provider, _ = cmd.Flags().GetString("provider")
		}
		if cmd.Flags().Changed("ota") {
			ns.Ota.Enabled, _ = cmd.Flags().GetBool("ota")
		}
		if cmd.Flags().Changed("webhook-url") {
			ns.Ota.WebhookUrl, _ = cmd.Flags().GetString("webhook-url")
		}
	}
	if cmd.Flags().Changed("max-devices") || cmd.Flags().Changed("device-mps") || cmd.Flags().Changed("namespace-mps") {
		if ns.Quota == nil {
			ns.Quota = &nspb.Quota{}
		}
		if cmd.Flags().Changed("max-devices") {
			n, _ := cmd.Flags().GetInt32("max-devices")
			ns.Quota.MaxDevices = n
		}
		if cmd.Flags().Changed("device-mps") {
			n, _ := cmd.Flags().GetFloat64("device-mps")
			ns.Quota.DeviceMessagesPerSecond = n
		}
		if cmd.Flags().Changed("namespace-mps") {
			n, _ := cmd.Flags().GetFloat64("namespace-mps")
			ns.Quota.NamespaceMessagesPerSecond = n
		}
	}
}

func nsFlags(cmd *cobra.Command) {
	cmd.Flags().String("title", "", "Namespace title")
	cmd.Flags().String("provider", "", "OTA provider: none, hawkbit, mender, or webhook")
	cmd.Flags().Bool("ota", false, "Turn the OTA switch on or off")
	cmd.Flags().String("webhook-url", "", "Webhook URL when the provider is webhook")
	cmd.Flags().Int32("max-devices", 0, "Device quota. Zero means no limit")
	cmd.Flags().Float64("device-mps", 0, "Per-device messages per second. Zero means no limit")
	cmd.Flags().Float64("namespace-mps", 0, "Namespace messages per second. Zero means no limit")
}

func printNamespace(ns *nspb.Namespace) {
	fmt.Printf("UUID: %s\nTitle: %s\n", ns.GetUuid(), ns.GetTitle())
	if ota := ns.GetOta(); ota != nil {
		fmt.Printf("OTA: enabled=%t provider=%s addon=%t\n", ota.GetEnabled(), ota.GetProvider(), ota.GetAddonRegistered())
		if ota.GetWebhookUrl() != "" {
			fmt.Printf("Webhook: %s\n", ota.GetWebhookUrl())
		}
	}
	if q := ns.GetQuota(); q != nil {
		fmt.Printf("Quota: devices=%d device-mps=%g namespace-mps=%g\n", q.GetMaxDevices(), q.GetDeviceMessagesPerSecond(), q.GetNamespaceMessagesPerSecond())
	}
}

var getNsCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a namespace",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := namespaceID(args)
		if err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeNamespacesServiceClient(ctx)
		if err != nil {
			return err
		}
		ns, err := client.Get(ctx, &nspb.Namespace{Uuid: id})
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(ns)
		}
		printNamespace(ns)
		return nil
	},
}

var createNsCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a namespace",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			return errors.New("--title is required")
		}
		ns := &nspb.Namespace{Title: title}
		fillNamespace(cmd, ns)
		ctx := makeContextWithBearerToken()
		client, err := makeNamespacesServiceClient(ctx)
		if err != nil {
			return err
		}
		res, err := client.Create(ctx, ns)
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(res)
		}
		printNamespace(res)
		return nil
	},
}

var updateNsCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a namespace, including OTA and quota",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := namespaceID(args)
		if err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeNamespacesServiceClient(ctx)
		if err != nil {
			return err
		}
		ns, err := client.Get(ctx, &nspb.Namespace{Uuid: id})
		if err != nil {
			return err
		}
		fillNamespace(cmd, ns)
		res, err := client.Update(ctx, ns)
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(res)
		}
		printNamespace(res)
		return nil
	},
}

var deleteNsCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a namespace",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := namespaceID(args)
		if err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeNamespacesServiceClient(ctx)
		if err != nil {
			return err
		}
		if _, err = client.Delete(ctx, &nspb.Namespace{Uuid: id}); err != nil {
			return err
		}
		fmt.Println("deleted", id)
		return nil
	},
}

func init() {
	nsFlags(createNsCmd)
	nsFlags(updateNsCmd)
	nsCmd.AddCommand(getNsCmd, createNsCmd, updateNsCmd, deleteNsCmd)
}
