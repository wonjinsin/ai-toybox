package service_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence"
	"github.com/wonjinsin/ledger/internal/core/service"
)

type askAI struct {
	sqlReply      string
	analysisReply string
	prompts       []string
}

func (a *askAI) Run(_ context.Context, prompt string) (string, error) {
	a.prompts = append(a.prompts, prompt)
	switch {
	case strings.Contains(prompt, "translate a Korean question"):
		return a.sqlReply, nil
	case strings.Contains(prompt, "가계부 분석 도우미"):
		return a.analysisReply, nil
	default:
		return "", nil
	}
}

func setupAsk(t *testing.T) (*askAI, *service.AskService) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := persistence.Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	cat, err := db.Client.Category.Create().SetName("카페/간식").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	src, err := db.Client.Source.Create().SetName("신한체크").SetKind("card").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Client.Transaction.Create().
		SetSourceID(src.ID).SetTxDate("2026-07-01").SetAmount(-4500).
		SetMerchant("스타벅스").SetMemo("").SetRawLine("r1").SetHash("h1").
		SetCategoryID(cat.ID).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}

	roExec, err := persistence.NewReadOnlyExecutor(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { roExec.Close() })

	ai := &askAI{analysisReply: "스타벅스에서 4,500원 썼습니다."}
	return ai, service.NewAskService(roExec, ai)
}

func TestAskPipeline(t *testing.T) {
	ai, svc := setupAsk(t)
	ai.sqlReply = "SELECT merchant, amount FROM transactions"

	res, err := svc.Ask(context.Background(), "어디에 썼어?")
	if err != nil {
		t.Fatal(err)
	}
	if res.SQL != "SELECT merchant, amount FROM transactions" {
		t.Errorf("SQL = %q", res.SQL)
	}
	if res.Answer != "스타벅스에서 4,500원 썼습니다." {
		t.Errorf("Answer = %q", res.Answer)
	}

	var sqlGen, analysis string
	for _, p := range ai.prompts {
		if strings.Contains(p, "translate a Korean question") {
			sqlGen = p
		}
		if strings.Contains(p, "분석 도우미") {
			analysis = p
		}
	}
	// SQL-gen prompt carries schema DDL, categories and the question.
	for _, want := range []string{"CREATE TABLE transactions", "카페/간식", "어디에 썼어?"} {
		if !strings.Contains(sqlGen, want) {
			t.Errorf("sql-gen prompt missing %q", want)
		}
	}
	// Analysis prompt carries the executed SQL and actual result rows.
	for _, want := range []string{"SELECT merchant, amount FROM transactions", "스타벅스", "-4500", "어디에 썼어?"} {
		if !strings.Contains(analysis, want) {
			t.Errorf("analysis prompt missing %q", want)
		}
	}
}

func TestAskRejectsUnsafeGeneratedSQL(t *testing.T) {
	ai, svc := setupAsk(t)
	ai.sqlReply = "DELETE FROM transactions"

	if _, err := svc.Ask(context.Background(), "다 지워줘"); err == nil {
		t.Fatal("unsafe SQL must be rejected")
	}
	// Analysis must never run after a rejected query.
	for _, p := range ai.prompts {
		if strings.Contains(p, "분석 도우미") {
			t.Error("analysis prompt should not be sent")
		}
	}
}
