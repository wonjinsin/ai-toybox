package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newTxCmd(open func(ctx context.Context) (*deps, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Inspect and delete stored transactions",
	}

	var match string
	list := &cobra.Command{
		Use:   "list",
		Short: "List transactions, optionally filtered by merchant/memo substring",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := open(cmd.Context())
			if err != nil {
				return err
			}
			txs, err := d.txs.Search(cmd.Context(), match)
			if err != nil {
				return err
			}
			if len(txs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no transactions")
				return nil
			}
			for _, t := range txs {
				memo := ""
				if t.Memo != "" {
					memo = " · " + t.Memo
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s %d %s%s (source %d)\n",
					t.ID, t.TxDate, t.Amount, t.Merchant, memo, t.SourceID)
			}
			return nil
		},
	}
	list.Flags().StringVar(&match, "match", "", "only rows whose merchant or memo contains this text")

	del := &cobra.Command{
		Use:   "delete <id>...",
		Short: "Delete transactions by id",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := make([]int, len(args))
			for i, a := range args {
				id, err := strconv.Atoi(a)
				if err != nil {
					return fmt.Errorf("invalid transaction id %q", a)
				}
				ids[i] = id
			}
			d, err := open(cmd.Context())
			if err != nil {
				return err
			}
			n, err := d.txs.Delete(cmd.Context(), ids)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d transaction(s) deleted\n", n)
			return nil
		},
	}

	cmd.AddCommand(list, del)
	return cmd
}
