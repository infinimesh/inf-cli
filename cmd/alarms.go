package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	rulespb "github.com/infinimesh/proto/node/rules"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var alarmsCmd = &cobra.Command{Use: "alarms", Short: "List and update alarms", Aliases: []string{"alarm"}, RunE: listAlarmsCmd.RunE}

var listAlarmsCmd = &cobra.Command{
	Use:   "list",
	Short: "List alarms. The default is raised and acknowledged",
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		device, _ := cmd.Flags().GetString("device")
		stateName, _ := cmd.Flags().GetString("state")
		req := &rulespb.ListAlarmsRequest{Namespace: ns, Device: device}
		switch strings.ToLower(stateName) {
		case "", "raised":
		case "acked":
			req.State = rulespb.AlarmState_ACKED
		case "cleared":
			req.State = rulespb.AlarmState_CLEARED
		case "any":
			req.AnyState = true
		default:
			return errors.New("state must be raised, acked, cleared, or any")
		}
		ctx := makeContextWithBearerToken()
		client, err := makeRulesClient(ctx)
		if err != nil {
			return err
		}
		var alarms []*rulespb.Alarm
		for {
			res, err := client.ListAlarms(ctx, req)
			if err != nil {
				return err
			}
			alarms = append(alarms, res.GetAlarms()...)
			if req.PageToken = res.GetNextPageToken(); req.PageToken == "" {
				if printJson, _ := cmd.Flags().GetBool("json"); printJson {
					return printJsonResponse(&rulespb.Alarms{Alarms: alarms})
				}
				break
			}
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Rule", "Device", "Title", "State", "Namespace"})
		for _, alarm := range alarms {
			t.AppendRow(table.Row{alarm.GetRule(), alarm.GetDevice(), alarm.GetTitle(), alarm.GetState().String(), alarm.GetNamespace()})
		}
		t.Render()
		return nil
	},
}

func alarmRef(cmd *cobra.Command) (*rulespb.AlarmRef, error) {
	rule, _ := cmd.Flags().GetString("rule")
	device, _ := cmd.Flags().GetString("device")
	if rule == "" || device == "" {
		return nil, errors.New("--rule and --device are required")
	}
	return &rulespb.AlarmRef{Rule: rule, Device: device}, nil
}

var ackAlarmCmd = &cobra.Command{
	Use:   "ack",
	Short: "Acknowledge an alarm",
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := alarmRef(cmd)
		if err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeRulesClient(ctx)
		if err != nil {
			return err
		}
		alarm, err := client.AckAlarm(ctx, ref)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s %s\n", alarm.GetRule(), alarm.GetDevice(), alarm.GetState().String())
		return nil
	},
}

var clearAlarmCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear an alarm",
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := alarmRef(cmd)
		if err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeRulesClient(ctx)
		if err != nil {
			return err
		}
		alarm, err := client.ClearAlarm(ctx, ref)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s %s\n", alarm.GetRule(), alarm.GetDevice(), alarm.GetState().String())
		return nil
	},
}

func init() {
	listAlarmsCmd.Flags().StringP("namespace", "n", "", "Namespace")
	listAlarmsCmd.Flags().String("device", "", "Device")
	listAlarmsCmd.Flags().String("state", "raised", "raised, acked, cleared, or any")
	for _, c := range []*cobra.Command{ackAlarmCmd, clearAlarmCmd} {
		c.Flags().String("rule", "", "Rule uuid")
		c.Flags().String("device", "", "Device uuid")
	}
	alarmsCmd.AddCommand(listAlarmsCmd, ackAlarmCmd, clearAlarmCmd)
	rootCmd.AddCommand(alarmsCmd)
}
