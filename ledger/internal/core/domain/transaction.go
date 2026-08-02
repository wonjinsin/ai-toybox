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
	Failed        []FailedRow
	MappingCached bool
}
