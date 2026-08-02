package domain

const (
	AmountModeSingle = "single" // one signed amount column
	AmountModeSplit  = "split"  // separate withdraw/deposit columns

	SignPositiveExpense = "positive_expense" // card CSVs: positive number means spending
	SignNegativeExpense = "negative_expense" // already-signed amounts
)

// ColumnMapping maps CSV columns to transaction fields.
// JSON tags define the contract with the AI mapping prompt.
type ColumnMapping struct {
	HeaderRows  int    `json:"header_rows"` // leading rows to skip (preamble + header line)
	DateCol     int    `json:"date_col"`
	MerchantCol int    `json:"merchant_col"`
	MemoCol     int    `json:"memo_col"` // -1 = none
	AmountMode  string `json:"amount_mode"`
	AmountCol   int    `json:"amount_col"`   // single mode
	Sign        string `json:"sign"`         // single mode
	WithdrawCol int    `json:"withdraw_col"` // split mode
	DepositCol  int    `json:"deposit_col"`  // split mode

	// Questions is non-empty when the AI is unsure (e.g. two merchant column
	// candidates); the user resolves them interactively before caching.
	Questions []MappingQuestion `json:"questions"`
}

type MappingQuestion struct {
	Field   string   `json:"field"`   // e.g. "merchant_col"
	Prompt  string   `json:"prompt"`  // human-readable question
	Options []string `json:"options"` // candidate answers, e.g. column indices with names
}
