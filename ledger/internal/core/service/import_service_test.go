package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence"
	"github.com/wonjinsin/ledger/internal/core/service"
)

const bankCSV = "신한은행 거래내역조회\n" +
	"거래일자,적요,내용,출금액,입금액,잔액\n" +
	"2026-07-01,체크카드,스타벅스 강남점,\"4,500\",,\"995,500\"\n" +
	"2026-07-02,급여,회사급여,,\"3,000,000\",\"3,995,500\"\n" +
	"2026-07-03,체크카드,GS25 역삼점,\"2,000\",,\"3,993,500\"\n"

const bankMappingJSON = `{"header_rows":2,"date_col":0,"merchant_col":2,"memo_col":1,` +
	`"amount_mode":"split","amount_col":-1,"sign":"negative_expense",` +
	`"withdraw_col":3,"deposit_col":4,"questions":[]}`

// countingRunner returns a fixed reply and counts invocations.
type countingRunner struct {
	reply string
	calls int
}

func (c *countingRunner) Run(_ context.Context, _ string) (string, error) {
	c.calls++
	return c.reply, nil
}

func setupImport(t *testing.T) (*persistence.DB, *countingRunner, *service.ImportService) {
	t.Helper()
	db, err := persistence.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	sourceRepo := persistence.NewSourceRepo(db.Client)
	if _, err := sourceRepo.Create(context.Background(), "신한체크", "card"); err != nil {
		t.Fatal(err)
	}
	runner := &countingRunner{reply: bankMappingJSON}
	svc := service.NewImportService(sourceRepo, persistence.NewMappingRepo(db.Client), persistence.NewTransactionRepo(db.Client), runner)
	return db, runner, svc
}

func toEUCKR(t *testing.T, s string) []byte {
	t.Helper()
	out, _, err := transform.Bytes(korean.EUCKR.NewEncoder(), []byte(s))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestImportEUCKRBankCSV(t *testing.T) {
	db, runner, svc := setupImport(t)

	res, err := svc.Import(context.Background(), toEUCKR(t, bankCSV), "신한체크")
	if err != nil {
		t.Fatal(err)
	}
	if res.Saved != 3 || res.DupSkipped != 0 || len(res.Failed) != 0 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if runner.calls != 1 {
		t.Errorf("AI mapping calls = %d, want 1", runner.calls)
	}
	if res.MappingCached {
		t.Error("first import should not be a cache hit")
	}

	var amount int64
	err = db.SQL.QueryRow(`SELECT amount FROM transactions WHERE merchant = '스타벅스 강남점'`).Scan(&amount)
	if err != nil || amount != -4500 {
		t.Errorf("expense amount = %d, %v; want -4500", amount, err)
	}
	err = db.SQL.QueryRow(`SELECT amount FROM transactions WHERE merchant = '회사급여'`).Scan(&amount)
	if err != nil || amount != 3000000 {
		t.Errorf("income amount = %d, %v; want 3000000", amount, err)
	}
}

func TestReimportSameFileSkipsAllAsDuplicates(t *testing.T) {
	_, _, svc := setupImport(t)
	data := toEUCKR(t, bankCSV)

	if _, err := svc.Import(context.Background(), data, "신한체크"); err != nil {
		t.Fatal(err)
	}
	res, err := svc.Import(context.Background(), data, "신한체크")
	if err != nil {
		t.Fatal(err)
	}
	if res.Saved != 0 || res.DupSkipped != 3 {
		t.Fatalf("reimport: saved=%d dup=%d, want 0/3", res.Saved, res.DupSkipped)
	}
	if !res.MappingCached {
		t.Error("second import should hit the mapping cache")
	}
}

func TestMappingCacheAvoidsSecondAICall(t *testing.T) {
	_, runner, svc := setupImport(t)

	if _, err := svc.Import(context.Background(), toEUCKR(t, bankCSV), "신한체크"); err != nil {
		t.Fatal(err)
	}
	// Same header, different data rows: must reuse the cached mapping.
	otherMonth := "신한은행 거래내역조회\n" +
		"거래일자,적요,내용,출금액,입금액,잔액\n" +
		"2026-08-01,체크카드,이마트,\"50,000\",,\"3,943,500\"\n"
	res, err := svc.Import(context.Background(), toEUCKR(t, otherMonth), "신한체크")
	if err != nil {
		t.Fatal(err)
	}
	if runner.calls != 1 {
		t.Errorf("AI calls = %d, want 1 (cache should prevent second call)", runner.calls)
	}
	if res.Saved != 1 {
		t.Errorf("saved = %d, want 1", res.Saved)
	}
}

func TestImportUnknownSourceFails(t *testing.T) {
	_, _, svc := setupImport(t)
	if _, err := svc.Import(context.Background(), []byte("a,b\n"), "없는소스"); err == nil {
		t.Fatal("want error for unregistered source")
	}
}
