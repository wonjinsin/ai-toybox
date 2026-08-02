package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence"
	"github.com/wonjinsin/ledger/internal/core/domain"
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

// scriptedPrompter answers mapping questions from a canned list and
// records what was asked. answers empty = fail the test if called.
type scriptedPrompter struct {
	t       *testing.T
	answers []string
	asked   []domain.MappingQuestion
}

func (p *scriptedPrompter) AskMapping(_ context.Context, q domain.MappingQuestion) (string, error) {
	p.asked = append(p.asked, q)
	if len(p.answers) == 0 {
		p.t.Fatalf("unexpected mapping question: %+v", q)
	}
	answer := p.answers[0]
	p.answers = p.answers[1:]
	return answer, nil
}

func setupImport(t *testing.T) (*persistence.DB, *countingRunner, *service.ImportService) {
	db, runner, _, svc := setupImportWithPrompter(t, nil)
	return db, runner, svc
}

func setupImportWithPrompter(t *testing.T, answers []string) (*persistence.DB, *countingRunner, *scriptedPrompter, *service.ImportService) {
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
	prompter := &scriptedPrompter{t: t, answers: answers}
	svc := service.NewImportService(sourceRepo, persistence.NewMappingRepo(db.Client), persistence.NewTransactionRepo(db.Client), runner, prompter)
	return db, runner, prompter, svc
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

func TestAmbiguousMappingAsksUserAndCachesResolution(t *testing.T) {
	// AI is unsure whether column 1 (적요) or 2 (내용) is the merchant.
	ambiguous := `{"header_rows":2,"date_col":0,"merchant_col":1,"memo_col":1,` +
		`"amount_mode":"split","amount_col":-1,"sign":"negative_expense",` +
		`"withdraw_col":3,"deposit_col":4,` +
		`"questions":[{"field":"merchant_col","prompt":"가맹점 컬럼은?","options":["1: 적요","2: 내용"]}]}`

	db, runner, prompter, svc := setupImportWithPrompter(t, []string{"2: 내용"})
	runner.reply = ambiguous

	res, err := svc.Import(context.Background(), toEUCKR(t, bankCSV), "신한체크")
	if err != nil {
		t.Fatal(err)
	}
	if len(prompter.asked) != 1 || prompter.asked[0].Field != "merchant_col" {
		t.Fatalf("prompter asked = %+v, want 1 merchant_col question", prompter.asked)
	}
	if res.Saved != 3 {
		t.Fatalf("saved = %d, want 3", res.Saved)
	}

	// The user's choice (column 2, 내용) must be applied...
	var count int
	if err := db.SQL.QueryRow(`SELECT COUNT(*) FROM transactions WHERE merchant = '스타벅스 강남점'`).Scan(&count); err != nil || count != 1 {
		t.Errorf("merchant from chosen column not applied (count=%d, err=%v)", count, err)
	}
	// ...and the cached mapping must be final: re-import asks nothing.
	prompter.answers = nil
	if _, err := svc.Import(context.Background(), toEUCKR(t, bankCSV), "신한체크"); err != nil {
		t.Fatal(err)
	}
	if len(prompter.asked) != 1 {
		t.Errorf("cached mapping should not re-ask; asked %d times", len(prompter.asked))
	}
}

func TestImportUnknownSourceFails(t *testing.T) {
	_, _, svc := setupImport(t)
	if _, err := svc.Import(context.Background(), []byte("a,b\n"), "없는소스"); err == nil {
		t.Fatal("want error for unregistered source")
	}
}
