package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/wonjinsin/ledger/internal/core/domain"
	"github.com/wonjinsin/ledger/internal/core/port/out"
)

type ImportService struct {
	sources  out.SourceRepo
	mappings out.MappingRepo
	txs      out.TransactionRepo
	ai       out.AIRunner
}

func NewImportService(sources out.SourceRepo, mappings out.MappingRepo, txs out.TransactionRepo, ai out.AIRunner) *ImportService {
	return &ImportService{sources: sources, mappings: mappings, txs: txs, ai: ai}
}

// Import parses raw CSV bytes and stores transactions for the named source.
func (s *ImportService) Import(ctx context.Context, data []byte, sourceName string) (*domain.ImportResult, error) {
	source, err := s.sources.GetByName(ctx, sourceName)
	if err != nil {
		return nil, fmt.Errorf("source %q not registered (use: ledger sources add): %w", sourceName, err)
	}

	utf8Data, err := decodeToUTF8(data)
	if err != nil {
		return nil, err
	}
	records, err := readCSV(utf8Data)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV")
	}

	hash := headerHash(records)
	mapping, cached, err := s.mappings.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	if !cached {
		mapping, err = s.generateMapping(ctx, records)
		if err != nil {
			return nil, err
		}
		if err := s.mappings.Save(ctx, hash, mapping); err != nil {
			return nil, err
		}
	}

	txs, failed := parseRows(records, mapping, source.ID)
	saved, dup, err := s.txs.BulkInsert(ctx, txs)
	if err != nil {
		return nil, err
	}
	return &domain.ImportResult{Saved: saved, DupSkipped: dup, Failed: failed, MappingCached: cached}, nil
}

const mappingPromptFmt = `You are a CSV column mapper for a Korean personal ledger.
Given the first rows of a bank/card transaction CSV, identify the columns.

Rules:
- header_rows: number of leading rows before the first data row (preamble + header line).
- Columns are 0-indexed.
- amount_mode "split": separate withdraw (출금) and deposit (입금) columns. Set withdraw_col and deposit_col; amount_col=-1.
- amount_mode "single": one amount column. Set amount_col. sign is "positive_expense" if a positive number means spending (typical card CSV), else "negative_expense". Set withdraw_col=-1, deposit_col=-1.
- merchant_col: the column best describing where money went (가맹점/거래처/적요/내용).
- memo_col: a secondary descriptive column, or -1.
- questions: normally []. ONLY if genuinely ambiguous (e.g. two equally plausible merchant columns), add {"field","prompt","options"} entries. prompt and options in Korean.

Reply with ONLY this JSON:
{"header_rows":int,"date_col":int,"merchant_col":int,"memo_col":int,"amount_mode":"single|split","amount_col":int,"sign":"positive_expense|negative_expense","withdraw_col":int,"deposit_col":int,"questions":[]}

CSV first rows:
%s`

func (s *ImportService) generateMapping(ctx context.Context, records [][]string) (*domain.ColumnMapping, error) {
	sample := records[:min(len(records), 8)]
	var b strings.Builder
	for _, row := range sample {
		b.WriteString(strings.Join(row, ","))
		b.WriteByte('\n')
	}
	var m domain.ColumnMapping
	if err := runJSON(ctx, s.ai, fmt.Sprintf(mappingPromptFmt, b.String()), &m); err != nil {
		return nil, fmt.Errorf("generate column mapping: %w", err)
	}
	if m.AmountMode != domain.AmountModeSingle && m.AmountMode != domain.AmountModeSplit {
		return nil, fmt.Errorf("AI mapping has invalid amount_mode %q", m.AmountMode)
	}
	return &m, nil
}
