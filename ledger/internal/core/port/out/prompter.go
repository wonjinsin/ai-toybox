package out

import (
	"context"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

// Prompter asks the user to resolve ambiguities during interactive import.
// The core stays terminal-agnostic; the CLI adapter provides the implementation.
type Prompter interface {
	// AskMapping returns the chosen option string (format "<value>: <description>").
	AskMapping(ctx context.Context, q domain.MappingQuestion) (string, error)
	// AskRowGroup asks what to do with a group of questionable rows.
	AskRowGroup(ctx context.Context, q domain.RowGroupQuestion) (domain.Decision, error)
}
