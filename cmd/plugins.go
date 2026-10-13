package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	plugpb "github.com/infinimesh/proto/plugins"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func makePluginsClient(ctx context.Context) (plugpb.PluginsServiceClient, error) {
	conn, err := makeConnection(ctx)
	if err != nil {
		return nil, err
	}
	return plugpb.NewPluginsServiceClient(conn), nil
}

func pluginKind(name string) (plugpb.PluginKind, error) {
	switch strings.ToLower(name) {
	case "", "embedded":
		return plugpb.PluginKind_EMBEDDED, nil
	case "ota":
		return plugpb.PluginKind_OTA, nil
	default:
		return plugpb.PluginKind_UNKNOWN, errors.New("kind must be embedded or ota")
	}
}

func printPlugin(p *plugpb.Plugin) {
	fmt.Printf("%s\t%s\t%s\tpublic=%t\n", p.GetUuid(), p.GetKind().String(), p.GetTitle(), p.GetPublic())
	if p.GetProvider() != "" {
		fmt.Printf("provider: %s\n", p.GetProvider())
	}
	if p.GetToken() != "" {
		fmt.Printf("token (shown once): %s\n", p.GetToken())
	}
}

var pluginsCmd = &cobra.Command{
	Use:     "plugins",
	Short:   "Manage plugins",
	Aliases: []string{"plugin"},
}

var listPluginsCmd = &cobra.Command{
	Use:   "list",
	Short: "List plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := makeContextWithBearerToken()
		client, err := makePluginsClient(ctx)
		if err != nil {
			return err
		}
		req := &plugpb.ListRequest{}
		if cmd.Flags().Changed("namespace") {
			ns, _ := cmd.Flags().GetString("namespace")
			req.Namespace = &ns
		}
		var plugins []*plugpb.Plugin
		for {
			res, err := client.List(ctx, req)
			if err != nil {
				return err
			}
			plugins = append(plugins, res.GetPlugins()...)
			if req.PageToken = res.GetNextPageToken(); req.PageToken == "" {
				if printJson, _ := cmd.Flags().GetBool("json"); printJson {
					return printJsonResponse(&plugpb.Plugins{Plugins: plugins})
				}
				break
			}
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"UUID", "Kind", "Title", "Provider", "Public"})
		for _, p := range plugins {
			t.AppendRow(table.Row{p.GetUuid(), p.GetKind().String(), p.GetTitle(), p.GetProvider(), p.GetPublic()})
		}
		t.Render()
		return nil
	},
}

var getPluginCmd = &cobra.Command{
	Use:   "get <uuid>",
	Short: "Get a plugin",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "" {
			return errors.New("plugin is required")
		}
		ctx := makeContextWithBearerToken()
		client, err := makePluginsClient(ctx)
		if err != nil {
			return err
		}
		res, err := client.Get(ctx, &plugpb.Plugin{Uuid: args[0]})
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(res)
		}
		printPlugin(res)
		return nil
	},
}

var createPluginCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a plugin",
	Long: `Create needs a root token. Mint one with POST /v1/token and "inf": true.

The addon bearer is printed once and is not stored.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			return errors.New("--title is required")
		}
		kind, err := pluginKind(mustString(cmd, "kind"))
		if err != nil {
			return err
		}
		provider, _ := cmd.Flags().GetString("provider")
		if kind == plugpb.PluginKind_OTA && provider == "" {
			return errors.New("--provider is required for an OTA plugin")
		}
		plugin := &plugpb.Plugin{
			Title:    title,
			Kind:     kind,
			Provider: provider,
			Public:   mustBool(cmd, "public"),
		}
		if cmd.Flags().Changed("description") {
			plugin.Description, _ = cmd.Flags().GetString("description")
		}
		if cmd.Flags().Changed("namespace") {
			ns, _ := cmd.Flags().GetString("namespace")
			plugin.Namespace = &ns
		}
		if cmd.Flags().Changed("frame-url") {
			frame, _ := cmd.Flags().GetString("frame-url")
			plugin.EmbeddedConf = &plugpb.EmbededPluginConf{FrameUrl: frame}
		}
		ctx := makeContextWithBearerToken()
		client, err := makePluginsClient(ctx)
		if err != nil {
			return err
		}
		res, err := client.Create(ctx, plugin)
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(res)
		}
		printPlugin(res)
		return nil
	},
}

var deletePluginCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete a plugin",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "" {
			return errors.New("plugin is required")
		}
		ctx := makeContextWithBearerToken()
		client, err := makePluginsClient(ctx)
		if err != nil {
			return err
		}
		if _, err = client.Delete(ctx, &plugpb.Plugin{Uuid: args[0]}); err != nil {
			return err
		}
		fmt.Println("deleted", args[0])
		return nil
	},
}

func mustString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}

func mustBool(cmd *cobra.Command, name string) bool {
	v, _ := cmd.Flags().GetBool(name)
	return v
}

func init() {
	listPluginsCmd.Flags().String("namespace", "", "Namespace")
	createPluginCmd.Flags().String("title", "", "Title")
	createPluginCmd.Flags().String("kind", "embedded", "embedded or ota")
	createPluginCmd.Flags().String("provider", "", "hawkbit, mender, or webhook")
	createPluginCmd.Flags().String("namespace", "", "Namespace. Empty with --public registers a public plugin")
	createPluginCmd.Flags().Bool("public", false, "Visible in every namespace")
	createPluginCmd.Flags().String("description", "", "Description")
	createPluginCmd.Flags().String("frame-url", "", "Embedded frame URL")
	pluginsCmd.AddCommand(listPluginsCmd, getPluginCmd, createPluginCmd, deletePluginCmd)
	rootCmd.AddCommand(pluginsCmd)
}
