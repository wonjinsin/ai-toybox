package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newRulesCmd(open func(ctx context.Context) (*deps, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage import rules created from 'always' decisions",
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List import rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := open(cmd.Context())
			if err != nil {
				return err
			}
			rules, err := d.rules.List(cmd.Context())
			if err != nil {
				return err
			}
			if len(rules) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no import rules")
				return nil
			}
			for _, r := range rules {
				scope := "all sources"
				if r.SourceID != nil {
					scope = fmt.Sprintf("source %d", *r.SourceID)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s %q (%s)\n", r.ID, r.Action, r.Pattern, scope)
			}
			return nil
		},
	}

	del := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an import rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid rule id %q", args[0])
			}
			d, err := open(cmd.Context())
			if err != nil {
				return err
			}
			if err := d.rules.Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "rule %d deleted\n", id)
			return nil
		},
	}

	cmd.AddCommand(list, del)
	return cmd
}
