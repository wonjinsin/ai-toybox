package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/wonjinsin/ledger/internal/adapter/out/ai"
	"github.com/wonjinsin/ledger/internal/adapter/out/persistence"
	"github.com/wonjinsin/ledger/internal/config"
	"github.com/wonjinsin/ledger/internal/core/service"
)

// deps carries wired usecases for commands; built lazily on first use
// so that --help never touches the DB.
type deps struct {
	db       *persistence.DB
	source   *service.SourceService
	importer *service.ImportService
	rules    *persistence.RuleRepo
}

func NewRootCmd() *cobra.Command {
	cfg := &config.Config{}
	var d *deps

	root := &cobra.Command{
		Use:           "ledger",
		Short:         "AI-powered personal ledger",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.PersistentFlags().StringVar(&cfg.DBPath, "db", config.DefaultDBPath(), "path to SQLite database file")
	root.PersistentFlags().StringVar(&cfg.AI, "ai", "claude", "AI backend: claude | codex")

	open := func(ctx context.Context) (*deps, error) {
		if d != nil {
			return d, nil
		}
		db, err := persistence.Open(ctx, cfg.DBPath)
		if err != nil {
			return nil, err
		}
		runner, err := ai.NewRunner(cfg.AI)
		if err != nil {
			return nil, err
		}
		sourceRepo := persistence.NewSourceRepo(db.Client)
		ruleRepo := persistence.NewRuleRepo(db.Client)
		prompter := NewTerminalPrompter(os.Stdin, os.Stdout)
		d = &deps{
			db:     db,
			source: service.NewSourceService(sourceRepo),
			importer: service.NewImportService(sourceRepo,
				persistence.NewMappingRepo(db.Client), persistence.NewTransactionRepo(db.Client),
				ruleRepo, persistence.NewCategoryRepo(db.Client), persistence.NewMerchantCategoryRepo(db.Client),
				runner, prompter),
			rules: ruleRepo,
		}
		return d, nil
	}
	root.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if d != nil {
			return d.db.Close()
		}
		return nil
	}

	root.AddCommand(newSourcesCmd(open), newImportCmd(open), newRulesCmd(open))
	return root
}
