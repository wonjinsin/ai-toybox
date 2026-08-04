package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

const classifyPromptFmt = `You are classifying Korean merchants/counterparties from a personal
ledger into spending categories.

Allowed categories (use EXACTLY these strings):
%s

Guidelines:
- 급여/수입 for salary and other income entries.
- 경조/이체 for person-to-person transfers, congratulation/condolence money.
- Omit merchants you genuinely cannot classify — do NOT guess wildly.

Reply with ONLY this JSON:
{"classifications":[{"merchant":"<input merchant verbatim>","category":"<allowed category>"}]}

Merchants:
%s`

// categorize assigns categories to uncategorized transactions: cached
// merchant->category entries apply directly; unknown merchants go to the AI
// in one batch, and its answers are cached for future imports.
func (s *ImportService) categorize(ctx context.Context) (int, error) {
	merchants, err := s.txs.UncategorizedMerchants(ctx)
	if err != nil || len(merchants) == 0 {
		return 0, err
	}
	cache, err := s.merchantCats.All(ctx)
	if err != nil {
		return 0, err
	}

	apply := make(map[string]int)
	var unknown []string
	for _, m := range merchants {
		if id, ok := cache[m]; ok {
			apply[m] = id
		} else {
			unknown = append(unknown, m)
		}
	}

	if len(unknown) > 0 {
		classified, err := s.classifyWithAI(ctx, unknown)
		if err != nil {
			return 0, err
		}
		if err := s.merchantCats.SaveBatch(ctx, classified); err != nil {
			return 0, err
		}
		for m, id := range classified {
			apply[m] = id
		}
	}
	return s.txs.ApplyMerchantCategories(ctx, apply)
}

func (s *ImportService) classifyWithAI(ctx context.Context, merchants []string) (map[string]int, error) {
	valid := make(map[string]bool, len(domain.AllowedCategories))
	for _, c := range domain.AllowedCategories {
		valid[c] = true
	}
	inputs := make(map[string]bool, len(merchants))
	for _, m := range merchants {
		inputs[m] = true
	}

	var reply struct {
		Classifications []struct {
			Merchant string `json:"merchant"`
			Category string `json:"category"`
		} `json:"classifications"`
	}
	prompt := fmt.Sprintf(classifyPromptFmt,
		strings.Join(domain.AllowedCategories, ", "), strings.Join(merchants, "\n"))
	if err := runJSON(ctx, s.ai, prompt, &reply); err != nil {
		return nil, fmt.Errorf("classify merchants: %w", err)
	}

	result := make(map[string]int)
	for _, c := range reply.Classifications {
		// Ignore hallucinated merchants or categories outside the allowed set.
		if !inputs[c.Merchant] || !valid[c.Category] {
			continue
		}
		cat, err := s.categories.GetOrCreate(ctx, c.Category)
		if err != nil {
			return nil, err
		}
		result[c.Merchant] = cat.ID
	}
	return result, nil
}
