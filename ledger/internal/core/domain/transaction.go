package domain

type Transaction struct {
	ID         int
	SourceID   int
	TxDate     string // YYYY-MM-DD
	Amount     int64  // KRW; expense negative, income positive
	Merchant   string
	Memo       string
	CategoryID *int
	RawLine    string
	Hash       string
}

type FailedRow struct {
	Line   int
	Raw    string
	Reason string
}

type ImportResult struct {
	Saved         int
	DupSkipped    int
	RuleSkipped   int // dropped by persisted import rules
	UserSkipped   int // dropped by an interactive decision this run
	RulesCreated  int
	Categorized   int // transactions assigned a category this run
	Failed        []FailedRow
	MappingCached bool
}
