package service

import (
	"strings"
	"testing"
)

func TestValidateSQLAcceptsSelectAndCTE(t *testing.T) {
	for _, q := range []string{
		"SELECT * FROM transactions",
		"select sum(-amount) from transactions where amount < 0",
		"WITH m AS (SELECT 1) SELECT * FROM m",
	} {
		if err := validateSQL(q); err != nil {
			t.Errorf("validateSQL(%q) = %v, want nil", q, err)
		}
	}
}

func TestValidateSQLRejectsUnsafe(t *testing.T) {
	for _, q := range []string{
		"DELETE FROM transactions",
		"INSERT INTO transactions VALUES (1)",
		"PRAGMA writable_schema = 1",
		"SELECT 1; DROP TABLE transactions",
		"UPDATE transactions SET amount = 0",
		"SELECT * FROM t; PRAGMA integrity_check",
	} {
		if err := validateSQL(q); err == nil {
			t.Errorf("validateSQL(%q) should fail", q)
		}
	}
}

func TestCleanSQLStripsFencesAndSemicolon(t *testing.T) {
	in := "Here is the query:\n```sql\nSELECT 1;\n```"
	if got := cleanSQL(in); got != "SELECT 1" {
		t.Errorf("cleanSQL = %q", got)
	}
	if got := cleanSQL("  SELECT 2;  "); got != "SELECT 2" {
		t.Errorf("cleanSQL = %q", got)
	}
}

func TestFormatTableEmptyResult(t *testing.T) {
	out := formatTable([]string{"a", "b"}, nil)
	if !strings.Contains(out, "(0 rows)") {
		t.Errorf("empty result marker missing: %q", out)
	}
}
