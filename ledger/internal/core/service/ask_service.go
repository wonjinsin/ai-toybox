package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/wonjinsin/ledger/internal/core/port/out"
)

type AskService struct {
	queries out.QueryExecutor
	ai      out.AIRunner
	now     func() time.Time
}

func NewAskService(queries out.QueryExecutor, ai out.AIRunner) *AskService {
	return &AskService{queries: queries, ai: ai, now: time.Now}
}

type AskResult struct {
	SQL    string
	Answer string
}

const schemaDDL = `CREATE TABLE sources(id INTEGER PRIMARY KEY, name TEXT UNIQUE, kind TEXT); -- kind: 'bank' | 'card'
CREATE TABLE categories(id INTEGER PRIMARY KEY, name TEXT UNIQUE);
CREATE TABLE transactions(
  id INTEGER PRIMARY KEY,
  source_id INTEGER,       -- references sources.id
  tx_date TEXT,            -- 'YYYY-MM-DD'; use date() functions
  amount INTEGER,          -- KRW; expense NEGATIVE, income POSITIVE
  merchant TEXT,
  memo TEXT,
  category_id INTEGER,     -- references categories.id; NULL = uncategorized
  raw_line TEXT,
  hash TEXT
);`

const sqlGenPromptFmt = `You translate a Korean question about personal ledger data
into ONE SQLite SELECT query.

Schema:
%s

Existing categories: %s
Today: %s

Rules:
- Exactly one SELECT statement (WITH ... SELECT allowed). Nothing else.
- Expenses are NEGATIVE amounts. For "지출" totals use SUM(-amount) with amount < 0,
  or ABS() — never mix income in.
- For relative dates (지난달, 이번주 ...) compute concrete ranges from Today.
- Unless the question implies aggregation, add LIMIT 200.
- Reply with ONLY the SQL. No code fences, no prose, no trailing semicolon.

Question: %s`

const analysisPromptFmt = `당신은 개인 가계부 분석 도우미입니다.

사용자 질문: %s

실행한 SQL:
%s

쿼리 결과 (탭 구분):
%s

위 결과만 근거로 질문에 한국어로 답하세요. 금액은 원 단위 콤마 표기.
핵심 답 먼저, 그다음 눈에 띄는 패턴이 있으면 1-2문장 덧붙이기.
결과가 비어있으면 해당 내역이 없다고 답하세요.`

// Ask turns a natural-language question into SQL, runs it read-only,
// and asks the AI to explain the result in Korean.
func (s *AskService) Ask(ctx context.Context, question string) (*AskResult, error) {
	categories := s.categoryNames(ctx)

	reply, err := s.ai.Run(ctx, fmt.Sprintf(sqlGenPromptFmt,
		schemaDDL, categories, s.now().Format("2006-01-02"), question))
	if err != nil {
		return nil, fmt.Errorf("generate SQL: %w", err)
	}
	query := cleanSQL(reply)
	if err := validateSQL(query); err != nil {
		return nil, fmt.Errorf("AI generated unsafe SQL (%s): %w", query, err)
	}

	cols, rows, err := s.queries.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("run query %q: %w", query, err)
	}

	answer, err := s.ai.Run(ctx, fmt.Sprintf(analysisPromptFmt,
		question, query, formatTable(cols, rows)))
	if err != nil {
		return nil, fmt.Errorf("analyze result: %w", err)
	}
	return &AskResult{SQL: query, Answer: answer}, nil
}

func (s *AskService) categoryNames(ctx context.Context) string {
	_, rows, err := s.queries.Query(ctx, "SELECT name FROM categories ORDER BY name")
	if err != nil || len(rows) == 0 {
		return "(없음)"
	}
	names := make([]string, len(rows))
	for i, r := range rows {
		names[i] = r[0]
	}
	return strings.Join(names, ", ")
}

// cleanSQL strips code fences and surrounding prose the AI may add.
func cleanSQL(reply string) string {
	s := strings.TrimSpace(reply)
	if i := strings.Index(s, "```"); i != -1 {
		s = s[i+3:]
		s = strings.TrimPrefix(s, "sql")
		if j := strings.Index(s, "```"); j != -1 {
			s = s[:j]
		}
	}
	return strings.TrimSuffix(strings.TrimSpace(s), ";")
}

var forbiddenSQL = regexp.MustCompile(`(?i)\b(pragma|attach|detach|insert|update|delete|drop|alter|create|replace|vacuum|reindex)\b`)

// validateSQL is defense-in-depth on top of the read-only connection.
func validateSQL(query string) error {
	low := strings.ToLower(strings.TrimSpace(query))
	if !strings.HasPrefix(low, "select") && !strings.HasPrefix(low, "with") {
		return fmt.Errorf("only SELECT statements are allowed")
	}
	if strings.Contains(query, ";") {
		return fmt.Errorf("multiple statements are not allowed")
	}
	if m := forbiddenSQL.FindString(query); m != "" {
		return fmt.Errorf("forbidden keyword %q", m)
	}
	return nil
}

func formatTable(cols []string, rows [][]string) string {
	var b strings.Builder
	b.WriteString(strings.Join(cols, "\t"))
	b.WriteByte('\n')
	const maxRows = 300
	for i, row := range rows {
		if i == maxRows {
			fmt.Fprintf(&b, "... (%d rows total, truncated)\n", len(rows))
			break
		}
		b.WriteString(strings.Join(row, "\t"))
		b.WriteByte('\n')
	}
	if len(rows) == 0 {
		b.WriteString("(0 rows)\n")
	}
	return b.String()
}
