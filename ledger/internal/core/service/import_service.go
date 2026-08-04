package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/wonjinsin/ledger/internal/core/domain"
	"github.com/wonjinsin/ledger/internal/core/port/out"
)

type ImportService struct {
	sources      out.SourceRepo
	mappings     out.MappingRepo
	txs          out.TransactionRepo
	rules        out.RuleRepo
	categories   out.CategoryRepo
	merchantCats out.MerchantCategoryRepo
	ai           out.AIRunner
	prompter     out.Prompter
}

func NewImportService(
	sources out.SourceRepo, mappings out.MappingRepo, txs out.TransactionRepo,
	rules out.RuleRepo, categories out.CategoryRepo, merchantCats out.MerchantCategoryRepo,
	ai out.AIRunner, prompter out.Prompter,
) *ImportService {
	return &ImportService{
		sources: sources, mappings: mappings, txs: txs, rules: rules,
		categories: categories, merchantCats: merchantCats, ai: ai, prompter: prompter,
	}
}

// Import parses raw CSV bytes and stores transactions for the named source.
func (s *ImportService) Import(ctx context.Context, data []byte, sourceName string, opts domain.ImportOptions) (*domain.ImportResult, error) {
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
		if err := s.resolveMappingQuestions(ctx, mapping); err != nil {
			return nil, err
		}
	}

	txs, failed := parseRows(records, mapping, source.ID)
	// Cache the mapping only once it proved able to parse at least one row;
	// a broken mapping must not poison every future import of this format.
	if !cached && (len(txs) > 0 || len(failed) == 0) {
		if err := s.mappings.Save(ctx, hash, mapping); err != nil {
			return nil, err
		}
	}
	res := &domain.ImportResult{Failed: failed, MappingCached: cached}

	rules, err := s.rules.List(ctx)
	if err != nil {
		return nil, err
	}
	txs, ruleSkipped, exempt := applyRules(txs, rules, source.ID)
	res.RuleSkipped = ruleSkipped

	// Drop already-stored rows before review so a re-import never re-asks.
	txs, err = s.dropExisting(ctx, txs, res)
	if err != nil {
		return nil, err
	}

	if !opts.AutoYes && len(txs) > 0 {
		txs, err = s.reviewRows(ctx, txs, exempt, source.ID, res)
		if err != nil {
			return nil, err
		}
	}

	saved, dup, err := s.txs.BulkInsert(ctx, txs)
	if err != nil {
		return nil, err
	}
	res.Saved = saved
	res.DupSkipped += dup

	categorized, err := s.categorize(ctx)
	if err != nil {
		return nil, err
	}
	res.Categorized = categorized
	return res, nil
}

func (s *ImportService) dropExisting(ctx context.Context, txs []domain.Transaction, res *domain.ImportResult) ([]domain.Transaction, error) {
	hashes := make([]string, len(txs))
	for i, tx := range txs {
		hashes[i] = tx.Hash
	}
	existing, err := s.txs.ExistingHashes(ctx, hashes)
	if err != nil {
		return nil, err
	}
	kept := txs[:0]
	for _, tx := range txs {
		if existing[tx.Hash] {
			res.DupSkipped++
			continue
		}
		kept = append(kept, tx)
	}
	return kept, nil
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
- questions: normally []. ONLY if genuinely ambiguous (e.g. two equally plausible merchant columns), add {"field","prompt","options"} entries. prompt in Korean. Every option MUST start with the machine value followed by ": " and a Korean description — for column fields the value is the 0-indexed column number (e.g. "2: 내용 컬럼"), for sign/amount_mode it is the enum literal (e.g. "positive_expense: 양수가 지출").

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
	if m.HeaderRows < 0 || m.HeaderRows >= len(records) {
		return nil, fmt.Errorf("AI mapping has invalid header_rows %d for %d rows", m.HeaderRows, len(records))
	}
	return &m, nil
}

// resolveMappingQuestions lets the user settle AI-flagged ambiguities,
// then clears them so the cached mapping is final.
func (s *ImportService) resolveMappingQuestions(ctx context.Context, m *domain.ColumnMapping) error {
	for _, q := range m.Questions {
		answer, err := s.prompter.AskMapping(ctx, q)
		if err != nil {
			return fmt.Errorf("resolve mapping question %q: %w", q.Field, err)
		}
		if err := applyMappingAnswer(m, q.Field, answer); err != nil {
			return err
		}
	}
	m.Questions = nil
	return nil
}

// applyMappingAnswer sets the field named by the question to the answered
// option's value. Options use the format "<value>: <description>".
func applyMappingAnswer(m *domain.ColumnMapping, field, answer string) error {
	value := strings.TrimSpace(strings.SplitN(answer, ":", 2)[0])
	switch field {
	case "sign":
		m.Sign = value
		return nil
	case "amount_mode":
		m.AmountMode = value
		return nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("mapping answer %q for %s: expected leading integer", answer, field)
	}
	switch field {
	case "header_rows":
		m.HeaderRows = n
	case "date_col":
		m.DateCol = n
	case "merchant_col":
		m.MerchantCol = n
	case "memo_col":
		m.MemoCol = n
	case "amount_col":
		m.AmountCol = n
	case "withdraw_col":
		m.WithdrawCol = n
	case "deposit_col":
		m.DepositCol = n
	default:
		return fmt.Errorf("unknown mapping question field %q", field)
	}
	return nil
}
