package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newAskCmd(open func(ctx context.Context) (*deps, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "ask <질문>",
		Short: "Ask a natural-language question about your ledger",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := open(cmd.Context())
			if err != nil {
				return err
			}
			res, err := d.ask.Ask(cmd.Context(), strings.Join(args, " "))
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "SQL: %s\n\n%s\n", res.SQL, res.Answer)
			return nil
		},
	}
}
