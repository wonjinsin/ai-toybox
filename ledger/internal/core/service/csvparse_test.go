package service

import (
	"testing"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

func TestNormalizeDateFormats(t *testing.T) {
	cases := map[string]string{
		"2026-07-01":          "2026-07-01",
		"2026.07.01":          "2026-07-01",
		"2026/07/01":          "2026-07-01",
		"20260701":            "2026-07-01",
		"2026-07-01 12:34:56": "2026-07-01",
	}
	for in, want := range cases {
		got, err := normalizeDate(in)
		if err != nil || got != want {
			t.Errorf("normalizeDate(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := normalizeDate("잔액"); err == nil {
		t.Error("want error for non-date string")
	}
}

func TestParseAmount(t *testing.T) {
	cases := map[string]int64{
		"4,500":       4500,
		"₩12,000원":    12000,
		"":            0,
		"-":           0,
		"-3000":       -3000,
		" 1,234,567 ": 1234567,
	}
	for in, want := range cases {
		got, err := parseAmount(in)
		if err != nil || got != want {
			t.Errorf("parseAmount(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
}

func TestParseRowsSplitMode(t *testing.T) {
	records := [][]string{
		{"거래일자", "적요", "내용", "출금액", "입금액"},
		{"2026-07-01", "체크카드", "스타벅스", "4,500", ""},
		{"2026-07-02", "급여", "회사", "", "3,000,000"},
	}
	m := &domain.ColumnMapping{
		HeaderRows: 1, DateCol: 0, MerchantCol: 2, MemoCol: 1,
		AmountMode: domain.AmountModeSplit, WithdrawCol: 3, DepositCol: 4,
	}
	txs, failed := parseRows(records, m, 1)
	if len(failed) != 0 || len(txs) != 2 {
		t.Fatalf("txs=%d failed=%v", len(txs), failed)
	}
	if txs[0].Amount != -4500 {
		t.Errorf("expense should be negative: %d", txs[0].Amount)
	}
	if txs[1].Amount != 3000000 {
		t.Errorf("income should be positive: %d", txs[1].Amount)
	}
}

func TestParseRowsSingleModePositiveExpense(t *testing.T) {
	records := [][]string{
		{"이용일자", "가맹점", "이용금액"},
		{"2026.07.01", "쿠팡", "35,000"},
	}
	m := &domain.ColumnMapping{
		HeaderRows: 1, DateCol: 0, MerchantCol: 1, MemoCol: -1,
		AmountMode: domain.AmountModeSingle, AmountCol: 2, Sign: domain.SignPositiveExpense,
	}
	txs, failed := parseRows(records, m, 1)
	if len(failed) != 0 || len(txs) != 1 {
		t.Fatalf("txs=%d failed=%v", len(txs), failed)
	}
	if txs[0].Amount != -35000 {
		t.Errorf("card spend should be negative: %d", txs[0].Amount)
	}
}

func TestParseRowsReportsBadRows(t *testing.T) {
	records := [][]string{
		{"이용일자", "가맹점", "이용금액"},
		{"이월잔액입니다", "-", "-"},
		{"2026.07.01", "쿠팡", "35,000"},
	}
	m := &domain.ColumnMapping{
		HeaderRows: 1, DateCol: 0, MerchantCol: 1, MemoCol: -1,
		AmountMode: domain.AmountModeSingle, AmountCol: 2, Sign: domain.SignPositiveExpense,
	}
	txs, failed := parseRows(records, m, 1)
	if len(txs) != 1 || len(failed) != 1 {
		t.Fatalf("want 1 tx + 1 failed, got %d/%d", len(txs), len(failed))
	}
	if failed[0].Line != 2 {
		t.Errorf("failed line = %d, want 2", failed[0].Line)
	}
}
