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
