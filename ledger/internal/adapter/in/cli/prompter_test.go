package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

func TestTerminalPrompterAskMapping(t *testing.T) {
	var out bytes.Buffer
	p := NewTerminalPrompter(strings.NewReader("9\n2\n"), &out) // invalid then valid
	q := domain.MappingQuestion{
		Field:   "merchant_col",
		Prompt:  "가맹점 컬럼은?",
		Options: []string{"1: 적요", "2: 내용"},
	}
	got, err := p.AskMapping(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if got != "2: 내용" {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(out.String(), "잘못된 입력") {
		t.Error("invalid input should be rejected with a message")
	}
}
