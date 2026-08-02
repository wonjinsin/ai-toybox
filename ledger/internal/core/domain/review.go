package domain

// ImportRule persists a user's "always" decision about questionable rows.
type ImportRule struct {
	ID       int
	SourceID *int // nil = all sources
	Pattern  string
	Action   string // RuleSkip | RuleInclude
}

const (
	RuleSkip    = "skip"
	RuleInclude = "include"
)

// RowGroupQuestion asks the user what to do with a group of same-kind
// questionable rows found by the AI review pass.
type RowGroupQuestion struct {
	Kind    string // cancel_pair | internal_transfer | zero_amount | duplicate_suspect | other
	Reason  string // Korean explanation for the user
	Pattern string // merchant/memo substring identifying the group ("" = no rule possible)
	Rows    []Transaction
}

type Decision string

const (
	DecisionInclude       Decision = "include"
	DecisionSkip          Decision = "skip"
	DecisionAlwaysSkip    Decision = "always_skip"
	DecisionAlwaysInclude Decision = "always_include"
)

type ImportOptions struct {
	AutoYes bool // --yes: no AI review pass, no questions, include everything
}
