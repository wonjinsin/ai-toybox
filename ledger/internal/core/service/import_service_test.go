package service_test

import (
	"context"
	"path/filepath"
	"strings"
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

const emptyReview = `{"groups":[]}`

// fakeAI routes prompts to canned replies by pipeline stage.
type fakeAI struct {
	mappingReply  string
	reviewReply   string
	classifyReply string
	mappingCalls  int
	reviewCalls   int
	classifyCalls int
	lastClassify  string
}

func (f *fakeAI) Run(_ context.Context, prompt string) (string, error) {
	switch {
	case strings.Contains(prompt, "CSV column mapper"):
		f.mappingCalls++
		return f.mappingReply, nil
	case strings.Contains(prompt, "reviewing parsed transactions"):
		f.reviewCalls++
		if f.reviewReply == "" {
			return emptyReview, nil
		}
		return f.reviewReply, nil
	case strings.Contains(prompt, "classifying Korean merchants"):
		f.classifyCalls++
		f.lastClassify = prompt
		if f.classifyReply == "" {
			return `{"classifications":[]}`, nil
		}
		return f.classifyReply, nil
	default:
		return "", nil
	}
}

// scriptedPrompter answers questions from canned lists; unexpected questions
// fail the test.
type scriptedPrompter struct {
	t            *testing.T
	answers      []string
	rowDecisions []domain.Decision
	asked        []domain.MappingQuestion
	rowAsked     []domain.RowGroupQuestion
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

func (p *scriptedPrompter) AskRowGroup(_ context.Context, q domain.RowGroupQuestion) (domain.Decision, error) {
	p.rowAsked = append(p.rowAsked, q)
	if len(p.rowDecisions) == 0 {
		p.t.Fatalf("unexpected row group question: %+v", q)
	}
	d := p.rowDecisions[0]
	p.rowDecisions = p.rowDecisions[1:]
	return d, nil
}

type fixture struct {
	db       *persistence.DB
	ai       *fakeAI
	prompter *scriptedPrompter
	svc      *service.ImportService
}

func setup(t *testing.T) *fixture {
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
	ai := &fakeAI{mappingReply: bankMappingJSON}
	prompter := &scriptedPrompter{t: t}
	svc := service.NewImportService(sourceRepo,
		persistence.NewMappingRepo(db.Client),
		persistence.NewTransactionRepo(db.Client),
		persistence.NewRuleRepo(db.Client),
		persistence.NewCategoryRepo(db.Client),
		persistence.NewMerchantCategoryRepo(db.Client),
		ai, prompter)
	return &fixture{db: db, ai: ai, prompter: prompter, svc: svc}
}

func toEUCKR(t *testing.T, s string) []byte {
	t.Helper()
	out, _, err := transform.Bytes(korean.EUCKR.NewEncoder(), []byte(s))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func (f *fixture) importCSV(t *testing.T, csv string, opts domain.ImportOptions) *domain.ImportResult {
	t.Helper()
	res, err := f.svc.Import(context.Background(), toEUCKR(t, csv), "신한체크", opts)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestImportEUCKRBankCSV(t *testing.T) {
	f := setup(t)
	res := f.importCSV(t, bankCSV, domain.ImportOptions{})

	if res.Saved != 3 || res.DupSkipped != 0 || len(res.Failed) != 0 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if f.ai.mappingCalls != 1 {
		t.Errorf("mapping calls = %d, want 1", f.ai.mappingCalls)
	}
	if res.MappingCached {
		t.Error("first import should not be a cache hit")
	}

	var amount int64
	err := f.db.SQL.QueryRow(`SELECT amount FROM transactions WHERE merchant = '스타벅스 강남점'`).Scan(&amount)
	if err != nil || amount != -4500 {
		t.Errorf("expense amount = %d, %v; want -4500", amount, err)
	}
	err = f.db.SQL.QueryRow(`SELECT amount FROM transactions WHERE merchant = '회사급여'`).Scan(&amount)
	if err != nil || amount != 3000000 {
		t.Errorf("income amount = %d, %v; want 3000000", amount, err)
	}
}

func TestReimportSameFileSkipsAllAsDuplicates(t *testing.T) {
	f := setup(t)
	f.importCSV(t, bankCSV, domain.ImportOptions{})
	res := f.importCSV(t, bankCSV, domain.ImportOptions{})

	if res.Saved != 0 || res.DupSkipped != 3 {
		t.Fatalf("reimport: saved=%d dup=%d, want 0/3", res.Saved, res.DupSkipped)
	}
	if !res.MappingCached {
		t.Error("second import should hit the mapping cache")
	}
}

func TestMappingCacheAvoidsSecondAICall(t *testing.T) {
	f := setup(t)
	f.importCSV(t, bankCSV, domain.ImportOptions{})

	// Same header, different data rows: must reuse the cached mapping.
	otherMonth := "신한은행 거래내역조회\n" +
		"거래일자,적요,내용,출금액,입금액,잔액\n" +
		"2026-08-01,체크카드,이마트,\"50,000\",,\"3,943,500\"\n"
	res := f.importCSV(t, otherMonth, domain.ImportOptions{})

	if f.ai.mappingCalls != 1 {
		t.Errorf("mapping calls = %d, want 1 (cache should prevent second call)", f.ai.mappingCalls)
	}
	if res.Saved != 1 {
		t.Errorf("saved = %d, want 1", res.Saved)
	}
}

func TestAmbiguousMappingAsksUserAndCachesResolution(t *testing.T) {
	f := setup(t)
	// AI is unsure whether column 1 (적요) or 2 (내용) is the merchant.
	f.ai.mappingReply = `{"header_rows":2,"date_col":0,"merchant_col":1,"memo_col":1,` +
		`"amount_mode":"split","amount_col":-1,"sign":"negative_expense",` +
		`"withdraw_col":3,"deposit_col":4,` +
		`"questions":[{"field":"merchant_col","prompt":"가맹점 컬럼은?","options":["1: 적요","2: 내용"]}]}`
	f.prompter.answers = []string{"2: 내용"}

	res := f.importCSV(t, bankCSV, domain.ImportOptions{})
	if len(f.prompter.asked) != 1 || f.prompter.asked[0].Field != "merchant_col" {
		t.Fatalf("prompter asked = %+v, want 1 merchant_col question", f.prompter.asked)
	}
	if res.Saved != 3 {
		t.Fatalf("saved = %d, want 3", res.Saved)
	}

	// The user's choice (column 2, 내용) must be applied...
	var count int
	if err := f.db.SQL.QueryRow(`SELECT COUNT(*) FROM transactions WHERE merchant = '스타벅스 강남점'`).Scan(&count); err != nil || count != 1 {
		t.Errorf("merchant from chosen column not applied (count=%d, err=%v)", count, err)
	}
	// ...and the cached mapping must be final: re-import asks nothing.
	f.importCSV(t, bankCSV, domain.ImportOptions{})
	if len(f.prompter.asked) != 1 {
		t.Errorf("cached mapping should not re-ask; asked %d times", len(f.prompter.asked))
	}
}

const cancelCSV = "신한은행 거래내역조회\n" +
	"거래일자,적요,내용,출금액,입금액,잔액\n" +
	"2026-07-01,체크카드,스타벅스 강남점,\"4,500\",,\"995,500\"\n" +
	"2026-07-01,승인취소,스타벅스 강남점 승인취소,,\"4,500\",\"1,000,000\"\n" +
	"2026-07-03,체크카드,GS25 역삼점,\"2,000\",,\"998,000\"\n"

const cancelReview = `{"groups":[{"kind":"cancel_pair",` +
	`"reason":"승인취소로 원거래와 상쇄됩니다","pattern":"승인취소","row_indexes":[1]}]}`

func TestReviewGroupAlwaysSkipCreatesRule(t *testing.T) {
	f := setup(t)
	f.ai.reviewReply = cancelReview
	f.prompter.rowDecisions = []domain.Decision{domain.DecisionAlwaysSkip}

	res := f.importCSV(t, cancelCSV, domain.ImportOptions{})
	if len(f.prompter.rowAsked) != 1 {
		t.Fatalf("row group questions = %d, want 1 (grouped, not per-row)", len(f.prompter.rowAsked))
	}
	if res.Saved != 2 || res.UserSkipped != 1 || res.RulesCreated != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}

	var count int
	if err := f.db.SQL.QueryRow(`SELECT COUNT(*) FROM import_rules WHERE pattern = '승인취소' AND action = 'skip'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("rule not persisted (count=%d, err=%v)", count, err)
	}

	// Next import: the rule drops matching rows BEFORE review — no question.
	res = f.importCSV(t, cancelCSV, domain.ImportOptions{})
	if len(f.prompter.rowAsked) != 1 {
		t.Errorf("rule should pre-drop rows; asked %d times", len(f.prompter.rowAsked))
	}
	if res.RuleSkipped != 1 {
		t.Errorf("rule-skipped = %d, want 1", res.RuleSkipped)
	}
}

func TestAutoYesSkipsReviewEntirely(t *testing.T) {
	f := setup(t)
	f.ai.reviewReply = cancelReview // would flag rows, but must never be called

	res := f.importCSV(t, cancelCSV, domain.ImportOptions{AutoYes: true})
	if f.ai.reviewCalls != 0 {
		t.Errorf("review AI calls = %d, want 0 with --yes", f.ai.reviewCalls)
	}
	if len(f.prompter.rowAsked) != 0 {
		t.Errorf("no questions expected with --yes")
	}
	if res.Saved != 3 {
		t.Errorf("saved = %d, want 3 (everything included)", res.Saved)
	}
}

func TestCategorizeNewMerchantsAndReuseCache(t *testing.T) {
	f := setup(t)
	f.ai.classifyReply = `{"classifications":[` +
		`{"merchant":"스타벅스 강남점","category":"카페/간식"},` +
		`{"merchant":"GS25 역삼점","category":"생활/마트"},` +
		`{"merchant":"회사급여","category":"급여/수입"}]}`

	res := f.importCSV(t, bankCSV, domain.ImportOptions{})
	if f.ai.classifyCalls != 1 {
		t.Fatalf("classify calls = %d, want 1", f.ai.classifyCalls)
	}
	if res.Categorized != 3 {
		t.Errorf("categorized = %d, want 3", res.Categorized)
	}

	var name string
	err := f.db.SQL.QueryRow(`SELECT c.name FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.merchant = '스타벅스 강남점'`).Scan(&name)
	if err != nil || name != "카페/간식" {
		t.Errorf("category = %q, %v; want 카페/간식", name, err)
	}
	var cached int
	if err := f.db.SQL.QueryRow(`SELECT COUNT(*) FROM merchant_categories`).Scan(&cached); err != nil || cached != 3 {
		t.Errorf("merchant_categories cache = %d, %v; want 3", cached, err)
	}

	// New file, same merchants: cache applies, AI not called again.
	augustCSV := "신한은행 거래내역조회\n" +
		"거래일자,적요,내용,출금액,입금액,잔액\n" +
		"2026-08-01,체크카드,스타벅스 강남점,\"5,000\",,\"990,500\"\n"
	res = f.importCSV(t, augustCSV, domain.ImportOptions{})
	if f.ai.classifyCalls != 1 {
		t.Errorf("classify calls = %d, want still 1 (cache hit)", f.ai.classifyCalls)
	}
	if res.Categorized != 1 {
		t.Errorf("categorized = %d, want 1", res.Categorized)
	}
}

func TestClassifyOnlySendsUnknownMerchants(t *testing.T) {
	f := setup(t)
	f.ai.classifyReply = `{"classifications":[{"merchant":"스타벅스 강남점","category":"카페/간식"}]}`
	f.importCSV(t, bankCSV, domain.ImportOptions{})

	// 스타벅스 is now cached; a new file with 스타벅스 + a new merchant must
	// only send the new one to the AI.
	f.ai.classifyReply = `{"classifications":[{"merchant":"이마트","category":"생활/마트"}]}`
	augustCSV := "신한은행 거래내역조회\n" +
		"거래일자,적요,내용,출금액,입금액,잔액\n" +
		"2026-08-01,체크카드,스타벅스 강남점,\"5,000\",,\"990,500\"\n" +
		"2026-08-02,체크카드,이마트,\"70,000\",,\"920,500\"\n"
	f.importCSV(t, augustCSV, domain.ImportOptions{})

	if strings.Contains(f.ai.lastClassify, "스타벅스") {
		t.Error("cached merchant must not be sent to the AI again")
	}
	if !strings.Contains(f.ai.lastClassify, "이마트") {
		t.Error("new merchant missing from classify prompt")
	}
}

func TestUnclassifiableMerchantStaysNull(t *testing.T) {
	f := setup(t)
	// AI omits 회사급여 and hallucinates an unknown category for GS25.
	f.ai.classifyReply = `{"classifications":[` +
		`{"merchant":"스타벅스 강남점","category":"카페/간식"},` +
		`{"merchant":"GS25 역삼점","category":"없는카테고리"}]}`

	res := f.importCSV(t, bankCSV, domain.ImportOptions{})
	if res.Categorized != 1 {
		t.Errorf("categorized = %d, want 1 (invalid category rejected)", res.Categorized)
	}
	var nulls int
	if err := f.db.SQL.QueryRow(`SELECT COUNT(*) FROM transactions WHERE category_id IS NULL`).Scan(&nulls); err != nil || nulls != 2 {
		t.Errorf("uncategorized rows = %d, %v; want 2", nulls, err)
	}
}

func TestImportUnknownSourceFails(t *testing.T) {
	f := setup(t)
	if _, err := f.svc.Import(context.Background(), []byte("a,b\n"), "없는소스", domain.ImportOptions{}); err == nil {
		t.Fatal("want error for unregistered source")
	}
}
