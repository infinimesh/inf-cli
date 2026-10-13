package cmd

import (
	"errors"
	"fmt"
	"strings"

	devpb "github.com/infinimesh/proto/node/devices"
	"github.com/spf13/cobra"
)

func deviceID(args []string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", errors.New("device is required")
	}
	return args[0], nil
}

func setDeviceEnabled(on bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		id, err := deviceID(args)
		if err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeDevicesServiceClient(ctx)
		if err != nil {
			return err
		}
		cur, err := client.Get(ctx, &devpb.Device{Uuid: id})
		if err != nil {
			return err
		}
		if cur.GetEnabled() == on {
			fmt.Printf("%s enabled=%t\n", id, on)
			return nil
		}
		if _, err = client.Toggle(ctx, &devpb.Device{Uuid: id}); err != nil {
			return err
		}
		fmt.Printf("%s enabled=%t\n", id, on)
		return nil
	}
}

var enableDeviceCmd = &cobra.Command{Use: "enable <id>", Short: "Enable a device", RunE: setDeviceEnabled(true)}
var disableDeviceCmd = &cobra.Command{Use: "disable <id>", Short: "Disable a device", RunE: setDeviceEnabled(false)}

var deleteDeviceCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a device",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := deviceID(args)
		if err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeDevicesServiceClient(ctx)
		if err != nil {
			return err
		}
		if _, err = client.Delete(ctx, &devpb.Device{Uuid: id}); err != nil {
			return err
		}
		fmt.Println("deleted", id)
		return nil
	},
}

var updateDeviceCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a device title or tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := deviceID(args)
		if err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeDevicesServiceClient(ctx)
		if err != nil {
			return err
		}
		cur, err := client.Get(ctx, &devpb.Device{Uuid: id})
		if err != nil {
			return err
		}
		title := cur.GetTitle()
		if cmd.Flags().Changed("title") {
			title, _ = cmd.Flags().GetString("title")
		}
		tags := cur.GetTags()
		if cmd.Flags().Changed("tags") {
			raw, _ := cmd.Flags().GetString("tags")
			tags = splitTags(raw)
		}
		res, err := client.Update(ctx, &devpb.Device{Uuid: id, Title: title, Tags: tags})
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(res)
		}
		PrintSingleDevice(res)
		return nil
	},
}

func splitTags(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func init() {
	updateDeviceCmd.Flags().String("title", "", "Device title")
	updateDeviceCmd.Flags().String("tags", "", "Comma-separated tags. Replaces the current tags.")
	devicesCmd.AddCommand(enableDeviceCmd, disableDeviceCmd, deleteDeviceCmd, updateDeviceCmd)
}
