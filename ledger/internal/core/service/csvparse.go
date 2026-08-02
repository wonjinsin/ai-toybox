package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

// decodeToUTF8 converts EUC-KR (common in Korean bank exports) to UTF-8;
// valid UTF-8 input passes through unchanged.
func decodeToUTF8(data []byte) ([]byte, error) {
	if utf8.Valid(data) {
		return data, nil
	}
	decoded, _, err := transform.Bytes(korean.EUCKR.NewDecoder(), data)
	if err != nil {
		return nil, fmt.Errorf("decode EUC-KR: %w", err)
	}
	return decoded, nil
}

func readCSV(data []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	return records, nil
}

// headerHash fingerprints the CSV format: the widest of the first rows is
// assumed to be the header line (bank exports often have preamble rows).
func headerHash(records [][]string) string {
	limit := min(len(records), 10)
	best := 0
	for i := 1; i < limit; i++ {
		if len(records[i]) > len(records[best]) {
			best = i
		}
	}
	sum := sha256.Sum256([]byte(strings.Join(records[best], ",")))
	return hex.EncodeToString(sum[:])
}

var dateLayouts = []string{"2006-01-02", "2006.01.02", "2006/01/02", "20060102"}

// normalizeDate parses common Korean bank/card date formats to YYYY-MM-DD.
// A trailing time part ("2026-08-01 12:34:56") is ignored.
func normalizeDate(s string) (string, error) {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, ' '); i > 0 {
		s = s[:i]
	}
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("unrecognized date %q", s)
}

func parseAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	for _, cut := range []string{",", "₩", "원", "\"", " "} {
		s = strings.ReplaceAll(s, cut, "")
	}
	if s == "" || s == "-" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unrecognized amount %q", s)
	}
	return n, nil
}

func cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// parseRows applies a confirmed mapping deterministically to all data rows.
func parseRows(records [][]string, m *domain.ColumnMapping, sourceID int) ([]domain.Transaction, []domain.FailedRow) {
	var txs []domain.Transaction
	var failed []domain.FailedRow

	for i := m.HeaderRows; i < len(records); i++ {
		row := records[i]
		raw := strings.Join(row, ",")
		if strings.TrimSpace(raw) == "" {
			continue
		}
		lineNo := i + 1

		date, err := normalizeDate(cell(row, m.DateCol))
		if err != nil {
			failed = append(failed, domain.FailedRow{Line: lineNo, Raw: raw, Reason: err.Error()})
			continue
		}

		amount, err := rowAmount(row, m)
		if err != nil {
			failed = append(failed, domain.FailedRow{Line: lineNo, Raw: raw, Reason: err.Error()})
			continue
		}

		merchant := cell(row, m.MerchantCol)
		memo := ""
		if m.MemoCol >= 0 {
			memo = cell(row, m.MemoCol)
		}
		if merchant == "" {
			merchant = memo
		}

		txs = append(txs, domain.Transaction{
			SourceID: sourceID,
			TxDate:   date,
			Amount:   amount,
			Merchant: merchant,
			Memo:     memo,
			RawLine:  raw,
			Hash:     txHash(sourceID, date, amount, merchant),
		})
	}
	return txs, failed
}

func rowAmount(row []string, m *domain.ColumnMapping) (int64, error) {
	switch m.AmountMode {
	case domain.AmountModeSplit:
		withdraw, err := parseAmount(cell(row, m.WithdrawCol))
		if err != nil {
			return 0, err
		}
		deposit, err := parseAmount(cell(row, m.DepositCol))
		if err != nil {
			return 0, err
		}
		return deposit - withdraw, nil
	case domain.AmountModeSingle:
		amount, err := parseAmount(cell(row, m.AmountCol))
		if err != nil {
			return 0, err
		}
		if m.Sign == domain.SignPositiveExpense {
			return -amount, nil
		}
		return amount, nil
	default:
		return 0, fmt.Errorf("unknown amount_mode %q", m.AmountMode)
	}
}

func txHash(sourceID int, date string, amount int64, merchant string) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%d|%s|%d|%s", sourceID, date, amount, merchant))
	return hex.EncodeToString(sum[:])
}
