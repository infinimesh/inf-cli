package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Print Prometheus metrics",
	Long: `Each core process serves GET /metrics on METRICS_ADDR. That default is :2112 and it stays inside the process.

Compose publishes host port 2112 on the mqtt bridge only. Pass --addr for the process you want. A quiet process still exports infinimesh_up.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, _ := cmd.Flags().GetString("addr")
		if addr == "" {
			return errors.New("--addr is required. Compose publishes 2112 on the mqtt bridge. Other processes serve :2112 inside the container")
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
	metricsCmd.Flags().String("addr", "", "Host and port, or a full URL, of one process metrics listener")
	rootCmd.AddCommand(metricsCmd)
}
