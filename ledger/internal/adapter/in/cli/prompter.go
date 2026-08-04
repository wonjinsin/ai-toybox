package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

// TerminalPrompter implements out.Prompter over stdin/stdout.
type TerminalPrompter struct {
	in  *bufio.Reader
	out io.Writer
}

func NewTerminalPrompter(in io.Reader, out io.Writer) *TerminalPrompter {
	return &TerminalPrompter{in: bufio.NewReader(in), out: out}
}

func (p *TerminalPrompter) AskRowGroup(_ context.Context, q domain.RowGroupQuestion) (domain.Decision, error) {
	fmt.Fprintf(p.out, "\n[%s] %s — %d건\n", q.Kind, q.Reason, len(q.Rows))
	shown := min(len(q.Rows), 5)
	for _, tx := range q.Rows[:shown] {
		fmt.Fprintf(p.out, "  %s  %d  %s\n", tx.TxDate, tx.Amount, tx.Merchant)
	}
	if len(q.Rows) > shown {
		fmt.Fprintf(p.out, "  ... 외 %d건\n", len(q.Rows)-shown)
	}
	for {
		fmt.Fprint(p.out, "처리: [i]포함 [s]스킵 [I]항상 포함 [S]항상 스킵 (기본 i): ")
		line, err := p.in.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read answer: %w", err)
		}
		switch strings.TrimSpace(line) {
		case "", "i":
			return domain.DecisionInclude, nil
		case "s":
			return domain.DecisionSkip, nil
		case "I":
			return domain.DecisionAlwaysInclude, nil
		case "S":
			return domain.DecisionAlwaysSkip, nil
		}
		fmt.Fprintln(p.out, "잘못된 입력입니다.")
	}
}

func (p *TerminalPrompter) AskMapping(_ context.Context, q domain.MappingQuestion) (string, error) {
	fmt.Fprintf(p.out, "\n%s\n", q.Prompt)
	for i, opt := range q.Options {
		fmt.Fprintf(p.out, "  [%d] %s\n", i+1, opt)
	}
	for {
		fmt.Fprintf(p.out, "선택 (1-%d): ", len(q.Options))
		line, err := p.in.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read answer: %w", err)
		}
		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err == nil && n >= 1 && n <= len(q.Options) {
			return q.Options[n-1], nil
		}
		fmt.Fprintln(p.out, "잘못된 입력입니다.")
	}
}
