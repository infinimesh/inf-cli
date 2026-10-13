package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Print Prometheus metrics",
	Long: `Each core process serves GET /metrics on METRICS_ADDR. The default is :2112 and it stays inside the process.

Compose does not publish 2112. Pass --addr with a host you can reach.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, _ := cmd.Flags().GetString("addr")
		if addr == "" {
			addr = "127.0.0.1:2112"
		}
		raw := addr
		if !strings.Contains(addr, "://") {
			raw = "http://" + addr
		}
		raw = strings.TrimRight(raw, "/") + "/metrics"
		client, err := httpClient()
		if err != nil {
			return err
		}
		res, err := client.Get(raw)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if res.StatusCode >= 300 {
			return fmt.Errorf("%s: %s", res.Status, strings.TrimSpace(string(body)))
		}
		fmt.Print(string(body))
		return nil
	},
}

func init() {
	metricsCmd.Flags().String("addr", "127.0.0.1:2112", "Host and port, or a full URL, of a process metrics listener")
	rootCmd.AddCommand(metricsCmd)
}
