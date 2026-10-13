package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"cel.dev/cel-go/cel"
	pb "github.com/infinimesh/proto/node"
	rulespb "github.com/infinimesh/proto/node/rules"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func makeRulesClient(ctx context.Context) (pb.RulesServiceClient, error) {
	conn, err := makeConnection(ctx)
	if err != nil {
		return nil, err
	}
	return pb.NewRulesServiceClient(conn), nil
}

func ruleFromFlags(cmd *cobra.Command, rule *rulespb.Rule) {
	if cmd.Flags().Changed("namespace") {
		rule.Namespace, _ = cmd.Flags().GetString("namespace")
	}
	if cmd.Flags().Changed("name") {
		rule.Name, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("when") {
		rule.When, _ = cmd.Flags().GetString("when")
	}
	if cmd.Flags().Changed("clear") {
		rule.Clear, _ = cmd.Flags().GetString("clear")
	}
	if cmd.Flags().Changed("cooldown") {
		rule.Cooldown, _ = cmd.Flags().GetString("cooldown")
	}
	if cmd.Flags().Changed("repeat") {
		rule.Repeat, _ = cmd.Flags().GetString("repeat")
	}
	if cmd.Flags().Changed("on-clear") {
		rule.OnClear, _ = cmd.Flags().GetBool("on-clear")
	}
}

func ruleFlags(cmd *cobra.Command) {
	cmd.Flags().String("namespace", "", "Namespace")
	cmd.Flags().String("name", "", "Rule name")
	cmd.Flags().String("when", "", "CEL expression that raises the alarm")
	cmd.Flags().String("clear", "", "CEL expression that clears the alarm")
	cmd.Flags().String("cooldown", "", "Shortest time between raises, for example 10m")
	cmd.Flags().String("repeat", "", "Repeat interval while raised and unacked, for example 1h")
	cmd.Flags().Bool("on-clear", false, "Run actions when the alarm clears")
}

func formatRule(rule *rulespb.Rule) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\t%s\t%s\nwhen: %s\nclear: %s\ncooldown: %s\nrepeat: %s\n",
		rule.GetUuid(), rule.GetNamespace(), rule.GetName(), rule.GetWhen(), rule.GetClear(), rule.GetCooldown(), rule.GetRepeat())
	if rule.GetOnClear() {
		fmt.Fprintf(&b, "on_clear: true\n")
	}
	return b.String()
}

func printRule(rule *rulespb.Rule) {
	fmt.Print(formatRule(rule))
}

var rulesCmd = &cobra.Command{Use: "rules", Short: "Manage rules", Aliases: []string{"rule"}, RunE: listRulesCmd.RunE}

var listRulesCmd = &cobra.Command{
	Use:   "list",
	Short: "List rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		ctx := makeContextWithBearerToken()
		client, err := makeRulesClient(ctx)
		if err != nil {
			return err
		}
		var rules []*rulespb.Rule
		token := ""
		for {
			res, err := client.List(ctx, &rulespb.ListRulesRequest{Namespace: ns, PageToken: token})
			if err != nil {
				return err
			}
			rules = append(rules, res.GetRules()...)
			if token = res.GetNextPageToken(); token == "" {
				if printJson, _ := cmd.Flags().GetBool("json"); printJson {
					return printJsonResponse(&rulespb.Rules{Rules: rules})
				}
				break
			}
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"UUID", "Namespace", "Name", "When"})
		for _, rule := range rules {
			t.AppendRow(table.Row{rule.GetUuid(), rule.GetNamespace(), rule.GetName(), rule.GetWhen()})
		}
		t.Render()
		return nil
	},
}

var getRuleCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("rule is required")
		}
		ctx := makeContextWithBearerToken()
		client, err := makeRulesClient(ctx)
		if err != nil {
			return err
		}
		rule, err := client.Get(ctx, &rulespb.Rule{Uuid: args[0]})
		if err != nil {
			return err
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			return printJsonResponse(rule)
		}
		printRule(rule)
		return nil
	},
}

var createRuleCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		rule := &rulespb.Rule{}
		ruleFromFlags(cmd, rule)
		if rule.Namespace == "" || rule.Name == "" || rule.When == "" {
			return errors.New("--namespace, --name, and --when are required")
		}
		if err := checkRule(rule.When, rule.Clear); err != nil {
			return err
		}
		ctx := makeContextWithBearerToken()
		client, err := makeRulesClient(ctx)
		if err != nil {
			return err
		}
		res, err := client.Create(ctx, rule)
		if err != nil {
			return err
		}
		printRule(res)
		return nil
	},
}

var updateRuleCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("rule is required")
		}
		ctx := makeContextWithBearerToken()
		client, err := makeRulesClient(ctx)
		if err != nil {
			return err
		}
		rule, err := client.Get(ctx, &rulespb.Rule{Uuid: args[0]})
		if err != nil {
			return err
		}
		ruleFromFlags(cmd, rule)
		if err := checkRule(rule.When, rule.Clear); err != nil {
			return err
		}
		res, err := client.Update(ctx, rule)
		if err != nil {
			return err
		}
		printRule(res)
		return nil
	},
}

var deleteRuleCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a rule and its alarms",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("rule is required")
		}
		ctx := makeContextWithBearerToken()
		client, err := makeRulesClient(ctx)
		if err != nil {
			return err
		}
		if _, err = client.Delete(ctx, &rulespb.Rule{Uuid: args[0]}); err != nil {
			return err
		}
		fmt.Println("deleted", args[0])
		return nil
	},
}

var dryRunRuleCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Compile a rule without saving it",
	RunE: func(cmd *cobra.Command, args []string) error {
		when, _ := cmd.Flags().GetString("when")
		clear, _ := cmd.Flags().GetString("clear")
		if when == "" {
			return errors.New("--when is required")
		}
		if err := checkRule(when, clear); err != nil {
			return err
		}
		fmt.Println("ok")
		return nil
	},
}

func checkRule(when, clear string) error {
	env, err := cel.NewEnv(
		cel.Variable("reported", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("desired", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("connected", cel.BoolType),
		cel.Variable("device", cel.StringType),
		cel.Variable("title", cel.StringType),
		cel.Variable("tags", cel.ListType(cel.StringType)),
		cel.Variable("ns", cel.StringType),
		cel.CrossTypeNumericComparisons(true),
	)
	if err != nil {
		return err
	}
	if err := compileCondition(env, when); err != nil {
		return fmt.Errorf("when: %w", err)
	}
	if clear != "" {
		if err := compileCondition(env, clear); err != nil {
			return fmt.Errorf("clear: %w", err)
		}
	}
	return nil
}

func compileCondition(env *cel.Env, expr string) error {
	ast, iss := env.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return iss.Err()
	}
	if ast.OutputType() != cel.BoolType && ast.OutputType() != cel.DynType {
		return fmt.Errorf("%q must be true or false", expr)
	}
	_, err := env.Program(ast)
	return err
}

func init() {
	listRulesCmd.Flags().String("namespace", "", "Namespace")
	ruleFlags(createRuleCmd)
	ruleFlags(updateRuleCmd)
	ruleFlags(dryRunRuleCmd)
	rulesCmd.AddCommand(listRulesCmd, getRuleCmd, createRuleCmd, updateRuleCmd, deleteRuleCmd, dryRunRuleCmd)
	rootCmd.AddCommand(rulesCmd)
}
