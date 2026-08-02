package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newSourcesCmd(open func(ctx context.Context) (*deps, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage account sources (bank accounts, cards)",
	}

	var kind string
	add := &cobra.Command{
		Use:   "add <name>",
		Short: "Register a new source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := open(cmd.Context())
			if err != nil {
				return err
			}
			s, err := d.source.Add(cmd.Context(), args[0], kind)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "source added: [%d] %s (%s)\n", s.ID, s.Name, s.Kind)
			return nil
		},
	}
	add.Flags().StringVar(&kind, "kind", "card", "source kind: bank | card")

	list := &cobra.Command{
		Use:   "list",
		Short: "List registered sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := open(cmd.Context())
			if err != nil {
				return err
			}
			sources, err := d.source.List(cmd.Context())
			if err != nil {
				return err
			}
			if len(sources) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no sources registered")
				return nil
			}
			for _, s := range sources {
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s (%s)\n", s.ID, s.Name, s.Kind)
			}
			return nil
		},
	}

	cmd.AddCommand(add, list)
	return cmd
}
