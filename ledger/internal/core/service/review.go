package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

const reviewPromptFmt = `You are reviewing parsed transactions from a Korean bank/card CSV
before they are saved to a personal ledger. Flag rows that may NOT belong:
- cancel_pair: cancellations/refunds that offset an original transaction (승인취소, 거래취소, 환불)
- internal_transfer: transfers between the user's own accounts/pay services (내계좌이체, 카카오페이 충전 등)
- card_settlement: card bill payments withdrawn from a bank account (카드대금, 카드결제, OO카드 출금 등) — these double-count spending when the card's own transactions are imported separately
- zero_amount: zero-amount rows
- duplicate_suspect: rows that look like duplicates of each other
- other: anything else worth confirming

Group flagged rows by kind. For each group give:
- reason: one short Korean sentence for the user
- pattern: a merchant/memo substring that identifies this group for future auto-handling, or ""
- row_indexes: indexes from the list below

Most rows are normal — when nothing is suspicious reply {"groups":[]}.
Reply with ONLY this JSON:
{"groups":[{"kind":"...","reason":"...","pattern":"...","row_indexes":[0]}]}

Rows (index|date|amount|merchant|memo):
%s`

type reviewGroup struct {
	Kind       string `json:"kind"`
	Reason     string `json:"reason"`
	Pattern    string `json:"pattern"`
	RowIndexes []int  `json:"row_indexes"`
}

// applyRules drops rows matching a skip rule and exempts rows matching an
// include rule from AI review. Returns (kept, skippedCount, exemptHashes).
func applyRules(txs []domain.Transaction, rules []*domain.ImportRule, sourceID int) ([]domain.Transaction, int, map[string]bool) {
	if len(rules) == 0 {
		return txs, 0, map[string]bool{}
	}
	kept := make([]domain.Transaction, 0, len(txs))
	exempt := make(map[string]bool)
	skipped := 0
	for _, tx := range txs {
		action, matched := matchRule(tx, rules, sourceID)
		switch {
		case matched && action == domain.RuleSkip:
			skipped++
		case matched && action == domain.RuleInclude:
			exempt[tx.Hash] = true
			kept = append(kept, tx)
		default:
			kept = append(kept, tx)
		}
	}
	return kept, skipped, exempt
}

func matchRule(tx domain.Transaction, rules []*domain.ImportRule, sourceID int) (string, bool) {
	haystack := strings.ToLower(tx.Merchant + " " + tx.Memo)
	for _, r := range rules {
		if r.SourceID != nil && *r.SourceID != sourceID {
			continue
		}
		if strings.Contains(haystack, strings.ToLower(r.Pattern)) {
			return r.Action, true
		}
	}
	return "", false
}

// reviewRows asks the AI to flag questionable rows, then the user to decide
// per group. Returns the rows to keep and updates counters on res.
func (s *ImportService) reviewRows(ctx context.Context, txs []domain.Transaction, exempt map[string]bool, sourceID int, res *domain.ImportResult) ([]domain.Transaction, error) {
	var b strings.Builder
	for i, tx := range txs {
		fmt.Fprintf(&b, "%d|%s|%d|%s|%s\n", i, tx.TxDate, tx.Amount, tx.Merchant, tx.Memo)
	}
	var reply struct {
		Groups []reviewGroup `json:"groups"`
	}
	if err := runJSON(ctx, s.ai, fmt.Sprintf(reviewPromptFmt, b.String()), &reply); err != nil {
		return nil, fmt.Errorf("review rows: %w", err)
	}

	drop := make(map[string]bool)
	for _, g := range validGroups(reply.Groups, len(txs)) {
		var rows []domain.Transaction
		for _, idx := range g.RowIndexes {
			if tx := txs[idx]; !exempt[tx.Hash] {
				rows = append(rows, tx)
			}
		}
		if len(rows) == 0 {
			continue
		}
		decision, err := s.prompter.AskRowGroup(ctx, domain.RowGroupQuestion{
			Kind: g.Kind, Reason: g.Reason, Pattern: g.Pattern, Rows: rows,
		})
		if err != nil {
			return nil, err
		}
		switch decision {
		case domain.DecisionSkip, domain.DecisionAlwaysSkip:
			for _, tx := range rows {
				drop[tx.Hash] = true
			}
			res.UserSkipped += len(rows)
		}
		if decision == domain.DecisionAlwaysSkip || decision == domain.DecisionAlwaysInclude {
			if g.Pattern == "" {
				continue // nothing to key a rule on; treat as one-off decision
			}
			action := domain.RuleSkip
			if decision == domain.DecisionAlwaysInclude {
				action = domain.RuleInclude
			}
			if _, err := s.rules.Create(ctx, domain.ImportRule{SourceID: &sourceID, Pattern: g.Pattern, Action: action}); err != nil {
				return nil, err
			}
			res.RulesCreated++
		}
	}

	kept := make([]domain.Transaction, 0, len(txs))
	for _, tx := range txs {
		if !drop[tx.Hash] {
			kept = append(kept, tx)
		}
	}
	return kept, nil
}

// validGroups drops groups whose row indexes are out of range (AI slip).
func validGroups(groups []reviewGroup, n int) []reviewGroup {
	valid := groups[:0]
	for _, g := range groups {
		ok := len(g.RowIndexes) > 0
		for _, idx := range g.RowIndexes {
			if idx < 0 || idx >= n {
				ok = false
				break
			}
		}
		if ok {
			valid = append(valid, g)
		}
	}
	return valid
}
