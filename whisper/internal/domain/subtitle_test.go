package domain

import (
	"strings"
	"testing"
	"time"
)

func TestCleanSubtitleCuesDropsLowConfidenceSpeech(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "残す", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "落とす", Probability: 0.31},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 1 || got[0].Text != "残す" {
		t.Fatalf("CleanSubtitleCues() = %#v, want only high-confidence cue", got)
	}
}

func TestCleanSubtitleCuesDropsJapaneseVocalizationButKeepsResponse(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "あー", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "はい", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 1 || got[0].Text != "はい" {
		t.Fatalf("CleanSubtitleCues() = %#v, want meaningful response only", got)
	}
}

func TestCleanSubtitleCuesDropsRepeatedKnownJapaneseHallucinations(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "ご視聴ありがとうございました。", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "今日はいい天気です", Probability: 0.9},
		{Start: 5 * time.Second, End: 6 * time.Second, Text: " ご視聴ありがとうございます ", Probability: 0.9},
		{Start: 7 * time.Second, End: 8 * time.Second, Text: "おめでとうございます！", Probability: 0.9},
		{Start: 9 * time.Second, End: 10 * time.Second, Text: "おめでとうございました", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 12*time.Second)
	if len(got) != 1 || got[0].Text != "今日はいい天気です" {
		t.Fatalf("CleanSubtitleCues() = %#v, want only unrelated speech", got)
	}
}

func TestCleanSubtitleCuesDropsRepeatedJapaneseNextVideoHallucinations(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "また次の動画でお会いしましょう。", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "残す", Probability: 0.9},
		{Start: 5 * time.Second, End: 6 * time.Second, Text: "次回の動画でお会いしましょう", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 8*time.Second)
	if len(got) != 1 || got[0].Text != "残す" {
		t.Fatalf("CleanSubtitleCues() = %#v, want only unrelated speech", got)
	}
}

func TestCleanSubtitleCuesDropsRepeatedJapaneseViewingThanksVariants(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "最後までご視聴いただきありがとうございました。", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "残す", Probability: 0.9},
		{Start: 5 * time.Second, End: 6 * time.Second, Text: "ご覧いただきありがとうございます", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 8*time.Second)
	if len(got) != 1 || got[0].Text != "残す" {
		t.Fatalf("CleanSubtitleCues() = %#v, want only unrelated speech", got)
	}
}

func TestCleanSubtitleCuesDropsJapaneseGoodNightHallucinationAtThreshold(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "おやすみなさい。", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "残す", Probability: 0.9},
		{Start: 5 * time.Second, End: 6 * time.Second, Text: "お休みなさい", Probability: 0.9},
		{Start: 7 * time.Second, End: 8 * time.Second, Text: "おやすみなさい", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 1 || got[0].Text != "残す" {
		t.Fatalf("CleanSubtitleCues() = %#v, want only unrelated speech", got)
	}
}

func TestCleanSubtitleCuesKeepsJapaneseGoodNightBelowThreshold(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "おやすみなさい。", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "お休みなさい", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 6*time.Second)
	if len(got) != len(cues) {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want %d legitimate phrases preserved", len(got), len(cues))
	}
}

func TestCleanSubtitleCuesDropsGenericJapaneseThanksAtThreshold(t *testing.T) {
	t.Parallel()

	for _, phrase := range []string{"ありがとうございました", "ありがとうございます"} {
		phrase := phrase
		t.Run(phrase, func(t *testing.T) {
			t.Parallel()

			cues := []Cue{
				{Start: time.Second, End: 2 * time.Second, Text: phrase + "。", Probability: 0.9},
				{Start: 3 * time.Second, End: 4 * time.Second, Text: "残す", Probability: 0.9},
				{Start: 5 * time.Second, End: 6 * time.Second, Text: phrase, Probability: 0.9},
				{Start: 7 * time.Second, End: 8 * time.Second, Text: " " + phrase + " ", Probability: 0.9},
				{Start: 9 * time.Second, End: 10 * time.Second, Text: phrase, Probability: 0.9},
			}

			got := CleanSubtitleCues(cues, "ja", nil, 12*time.Second)
			if len(got) != 1 || got[0].Text != "残す" {
				t.Fatalf("CleanSubtitleCues() = %#v, want only unrelated speech", got)
			}
		})
	}
}

func TestCleanSubtitleCuesDoesNotCombineGenericJapaneseThanksForms(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "ありがとうございました", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "ありがとうございました。", Probability: 0.9},
		{Start: 5 * time.Second, End: 6 * time.Second, Text: "ありがとうございます", Probability: 0.9},
		{Start: 7 * time.Second, End: 8 * time.Second, Text: "ありがとうございます。", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != len(cues) {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want %d separate legitimate thanks forms preserved", len(got), len(cues))
	}
}

func TestCleanSubtitleCuesKeepsGenericJapaneseThanksBelowThreshold(t *testing.T) {
	t.Parallel()

	for _, phrase := range []string{"ありがとうございました", "ありがとうございます"} {
		phrase := phrase
		t.Run(phrase, func(t *testing.T) {
			t.Parallel()

			cues := []Cue{
				{Start: time.Second, End: 2 * time.Second, Text: phrase, Probability: 0.9},
				{Start: 3 * time.Second, End: 4 * time.Second, Text: phrase + "。", Probability: 0.9},
				{Start: 5 * time.Second, End: 6 * time.Second, Text: " " + phrase + " ", Probability: 0.9},
			}

			got := CleanSubtitleCues(cues, "ja", nil, 8*time.Second)
			if len(got) != len(cues) {
				t.Fatalf("len(CleanSubtitleCues()) = %d, want %d legitimate thanks preserved", len(got), len(cues))
			}
		})
	}
}

func TestCleanSubtitleCuesKeepsSingleJapaneseNextVideoPhrase(t *testing.T) {
	t.Parallel()

	cues := []Cue{{
		Start: time.Second, End: 2 * time.Second, Text: "また次の動画でお会いしましょう", Probability: 0.9,
	}}

	got := CleanSubtitleCues(cues, "ja", nil, 4*time.Second)
	if len(got) != 1 || got[0].Text != cues[0].Text {
		t.Fatalf("CleanSubtitleCues() = %#v, want single legitimate phrase preserved", got)
	}
}

func TestCleanSubtitleCuesDropsRepeatedKnownJapaneseHallucinationsInAutoMode(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "ご視聴ありがとうございました", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "残す", Probability: 0.9},
		{Start: 5 * time.Second, End: 6 * time.Second, Text: "ご視聴ありがとうございました", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "auto", nil, 8*time.Second)
	if len(got) != 1 || got[0].Text != "残す" {
		t.Fatalf("CleanSubtitleCues() = %#v, want only unrelated speech", got)
	}
}

func TestCleanSubtitleCuesKeepsSingleKnownJapanesePhrase(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "ご視聴ありがとうございました", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "おめでとうございます", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 5*time.Second)
	if len(got) != 2 || got[0].Text != cues[0].Text || got[1].Text != cues[1].Text {
		t.Fatalf("CleanSubtitleCues() = %#v, want single legitimate phrases preserved", got)
	}
}

func TestCleanSubtitleCuesKeepsKnownJapanesePhraseWhenDuplicateIsLowConfidence(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "ご視聴ありがとうございました", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "ご視聴ありがとうございました", Probability: 0.31},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 5*time.Second)
	if len(got) != 1 || got[0].Text != cues[0].Text {
		t.Fatalf("CleanSubtitleCues() = %#v, want only high-confidence phrase preserved", got)
	}
}

func TestCleanSubtitleCuesKeepsRepeatedKnownJapanesePhraseForForcedOtherLanguage(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "ご視聴ありがとうございました", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "ご視聴ありがとうございました", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ko", nil, 5*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want forced non-Japanese language unchanged", len(got))
	}
}

func TestCleanSubtitleCuesKeepsSentencesContainingKnownJapanesePhrase(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "彼はご視聴ありがとうございましたと言いました", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "画面にご視聴ありがとうございましたと表示されます", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 5*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want sentences with embedded phrases preserved", len(got))
	}
}

func TestCleanSubtitleCuesKeepsKnownJapanesePhraseDecoratedWithSymbol(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "★ご視聴ありがとうございました", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "ご視聴ありがとうございました", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 5*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want symbol-decorated phrase preserved", len(got))
	}
}

func TestCleanSubtitleCuesKeepsJapaneseHallucinationPhraseWrappedInPunctuation(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "「次の動画でお会いしましょう」", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "「次の動画でお会いしましょう」", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 6*time.Second)
	if len(got) != len(cues) {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want punctuation-wrapped phrases preserved", len(got))
	}
}

func TestCleanSubtitleCuesKeepsNearbyDuplicateWithoutConfirmedBoundary(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 2 * time.Second, Text: "こんにちは", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "こんにちは", Probability: 0.9},
		{Start: 20 * time.Second, End: 21 * time.Second, Text: "こんにちは", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 30*time.Second)
	if len(got) != 3 || got[0].Start != time.Second || got[1].Start != 3*time.Second || got[2].Start != 20*time.Second {
		t.Fatalf("CleanSubtitleCues() = %#v, want every actual repeat", got)
	}
}

func TestCleanSubtitleCuesAppliesExternalCorrections(t *testing.T) {
	t.Parallel()

	cues := []Cue{{
		Start: time.Second, End: 2 * time.Second, Text: "レッサン帰り?", Probability: 0.9,
	}}

	got := CleanSubtitleCues(cues, "ja", map[string]string{"レッサン": "レッスン"}, 10*time.Second)
	if len(got) != 1 || got[0].Text != "レッスン帰り？" {
		t.Fatalf("CleanSubtitleCues() = %#v, want corrected Japanese text", got)
	}
}

func TestCleanSubtitleCuesCapsDisplayTimeBeforeNextCue(t *testing.T) {
	t.Parallel()

	cues := []Cue{
		{Start: time.Second, End: 20 * time.Second, Text: "첫 번째", Probability: 0.9},
		{Start: 5 * time.Second, End: 6 * time.Second, Text: "두 번째", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ko", nil, 10*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want 2", len(got))
	}
	if got[0].End != 4950*time.Millisecond {
		t.Errorf("first cue end = %s, want 4.95s", got[0].End)
	}
	if got[0].End > got[1].Start {
		t.Errorf("cues overlap: %#v", got)
	}
}

func TestCleanSubtitleCuesSplitsTextBeyondTwoLines(t *testing.T) {
	t.Parallel()

	text := "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十"
	cues := []Cue{{
		Start: time.Second, End: 5 * time.Second, Text: text, Probability: 0.9,
	}}

	got := CleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want 2", len(got))
	}
	if got[0].Text+got[1].Text != text {
		t.Errorf("split text = %q + %q, want original %q", got[0].Text, got[1].Text, text)
	}
	for _, cue := range got {
		if length := len([]rune(cue.Text)); length > 36 {
			t.Errorf("split cue length = %d, want at most 36", length)
		}
	}
}

func TestCleanSubtitleCuesSplitsLongCueAtActualTokenTimes(t *testing.T) {
	t.Parallel()

	text := "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十"
	cues := []Cue{{
		Start:       time.Second,
		End:         7 * time.Second,
		Text:        text,
		Probability: 0.9,
		Tokens: []Token{
			{Start: time.Second, End: 2 * time.Second, Text: "一二三四五六七八九十"},
			{Start: 2500 * time.Millisecond, End: 3800 * time.Millisecond, Text: "一二三四五六七八九十"},
			{Start: 5 * time.Second, End: 6 * time.Second, Text: "一二三四五六七八九十"},
			{Start: 6 * time.Second, End: 7 * time.Second, Text: "一二三四五六七八九十"},
		},
	}}

	got := CleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want 2", len(got))
	}
	if got[0].End != 3800*time.Millisecond {
		t.Errorf("first cue end = %s, want 3.8s", got[0].End)
	}
	if got[1].Start != 5*time.Second {
		t.Errorf("second cue start = %s, want 5s", got[1].Start)
	}
}

func TestCleanSubtitleCuesSortsTokenSplitCuesBeforeTiming(t *testing.T) {
	t.Parallel()

	firstHalf := strings.Repeat("前", 20)
	secondHalf := strings.Repeat("後", 20)
	cues := []Cue{
		{
			Start:       3 * time.Second,
			End:         9 * time.Second,
			Text:        firstHalf + secondHalf,
			Probability: 0.9,
			Tokens: []Token{
				{Start: 3 * time.Second, End: 6 * time.Second, Text: firstHalf},
				{Start: 7 * time.Second, End: 9 * time.Second, Text: secondHalf},
			},
		},
		{Start: 5 * time.Second, End: 5500 * time.Millisecond, Text: "割り込み", Probability: 0.9},
	}

	got := CleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if err := ValidateSubtitleCues(got, 10*time.Second); err != nil {
		t.Fatalf("ValidateSubtitleCues() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want 3", len(got))
	}
}

func TestCleanSubtitleCuesPreservesTokenTimingAroundCorrectionSource(t *testing.T) {
	t.Parallel()

	text := strings.Repeat("가", 15) + strings.Repeat("나", 10) + strings.Repeat("다", 15)
	source := strings.Repeat("가", 5) + strings.Repeat("나", 10) + strings.Repeat("다", 5)
	cues := []Cue{{
		Start:       time.Second,
		End:         7 * time.Second,
		Text:        text,
		Probability: 0.9,
		Tokens: []Token{
			{Start: time.Second, End: 2 * time.Second, Text: strings.Repeat("가", 10)},
			{Start: 2500 * time.Millisecond, End: 3800 * time.Millisecond, Text: strings.Repeat("가", 5) + strings.Repeat("나", 5)},
			{Start: 5 * time.Second, End: 6 * time.Second, Text: strings.Repeat("나", 5) + strings.Repeat("다", 5)},
			{Start: 6 * time.Second, End: 7 * time.Second, Text: strings.Repeat("다", 10)},
		},
	}}

	got := CleanSubtitleCues(cues, "ja", map[string]string{source: strings.Repeat("마", 20)}, 10*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want 2", len(got))
	}
	if got[0].End != 2*time.Second {
		t.Errorf("first cue end = %s, want 2s", got[0].End)
	}
	if got[1].Start != 2500*time.Millisecond {
		t.Errorf("second cue start = %s, want 2.5s", got[1].Start)
	}
	if got[1].Text != strings.Repeat("마", 20)+strings.Repeat("다", 10) {
		t.Errorf("second cue text = %q", got[1].Text)
	}
}

func TestCleanSubtitleCuesPrefersPunctuationTokenBoundary(t *testing.T) {
	t.Parallel()

	text := strings.Repeat("가", 9) + "。" + strings.Repeat("나", 10) + strings.Repeat("다", 10) + strings.Repeat("라", 10)
	cues := []Cue{{
		Start:       time.Second,
		End:         7 * time.Second,
		Text:        text,
		Probability: 0.9,
		Tokens: []Token{
			{Start: time.Second, End: 2 * time.Second, Text: strings.Repeat("가", 9) + "。"},
			{Start: 2500 * time.Millisecond, End: 3800 * time.Millisecond, Text: strings.Repeat("나", 10)},
			{Start: 5 * time.Second, End: 6 * time.Second, Text: strings.Repeat("다", 10)},
			{Start: 6 * time.Second, End: 7 * time.Second, Text: strings.Repeat("라", 10)},
		},
	}}

	got := CleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(CleanSubtitleCues()) = %d, want 2", len(got))
	}
	if got[0].End != 2*time.Second || got[1].Start != 2500*time.Millisecond {
		t.Errorf("cue timing = %s --> %s and %s --> %s, want 1s --> 2s and 2.5s --> 7s", got[0].Start, got[0].End, got[1].Start, got[1].End)
	}
}
