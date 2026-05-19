// ===================
// © AngelaMos | 2026
// lengthbias_test.go
// ===================

package audit

import (
	"strings"
	"testing"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

func q(id, correct int, opts ...string) parse.Question {
	return parse.Question{ID: id, CorrectAnswerIndex: correct, Options: opts}
}

func TestAnalyzeLengthBias_CharLongest(t *testing.T) {
	qs := []parse.Question{
		// correct (idx 1) is the strict longest by chars.
		q(
			1,
			1,
			"short",
			"this option is clearly the very longest one here",
			"mid sized",
			"tiny",
		),
	}
	lb := AnalyzeLengthBias(qs)
	if lb.CorrectIsLongestByCharCount != 1 {
		t.Fatalf(
			"CorrectIsLongestByCharCount = %d, want 1",
			lb.CorrectIsLongestByCharCount,
		)
	}
	if len(lb.QuestionsFlagged) != 1 || lb.QuestionsFlagged[0] != 1 {
		t.Fatalf("QuestionsFlagged = %v, want [1]", lb.QuestionsFlagged)
	}
	joined := strings.Join(lb.PerQuestion[1], " | ")
	if !strings.Contains(joined, "character count") {
		t.Errorf("expected a character-count issue, got %q", joined)
	}
}

func TestAnalyzeLengthBias_TieIsNotLongest(t *testing.T) {
	// correct ties another option for longest -> not a giveaway, not flagged.
	qs := []parse.Question{
		q(1, 0, "AAAAAAAAAA", "BBBBBBBBBB", "cc", "dd"),
	}
	lb := AnalyzeLengthBias(qs)
	if lb.CorrectIsLongestByCharCount != 0 {
		t.Errorf(
			"tie should not count as longest, got %d",
			lb.CorrectIsLongestByCharCount,
		)
	}
	if len(lb.QuestionsFlagged) != 0 {
		t.Errorf("tie should not be flagged, got %v", lb.QuestionsFlagged)
	}
}

func TestAnalyzeLengthBias_WordLongestNotCharLongest(t *testing.T) {
	// correct has the most WORDS but not the most CHARACTERS.
	qs := []parse.Question{
		q(
			1,
			0,
			"a b c d e f g h",
			"supercalifragilisticexpialidociousAAAAAAAA",
			"x",
			"y",
		),
	}
	lb := AnalyzeLengthBias(qs)
	if lb.CorrectIsLongestByCharCount != 0 {
		t.Errorf("CharCount should be 0, got %d", lb.CorrectIsLongestByCharCount)
	}
	if lb.CorrectIsLongestByWordCount != 1 {
		t.Errorf("WordCount should be 1, got %d", lb.CorrectIsLongestByWordCount)
	}
	if !strings.Contains(strings.Join(lb.PerQuestion[1], " "), "word") {
		t.Errorf("expected a word-count issue: %v", lb.PerQuestion[1])
	}
}

func TestAnalyzeLengthBias_QualifierOnly(t *testing.T) {
	// correct is the only option with a parenthetical AND a hedge word.
	qs := []parse.Question{
		q(
			1,
			2,
			"Alpha protocol",
			"Beta protocol",
			"Gamma protocol (typically used here)",
			"Delta protocol",
		),
		// correct is the only one with a "because" clause.
		q(
			2,
			0,
			"It works because the key is rotated",
			"Static keys",
			"No keys",
			"Shared keys",
		),
		// correct is the only multi-sentence option.
		q(
			3,
			1,
			"One thing.",
			"Two things happen. Then a third occurs.",
			"Three.",
			"Four.",
		),
	}
	lb := AnalyzeLengthBias(qs)
	if got := strings.Join(
		lb.PerQuestion[1],
		" ",
	); !strings.Contains(got, "parenthetical") ||
		!strings.Contains(got, "hedge") {
		t.Errorf("q1 expected parenthetical+hedge issues, got %q", got)
	}
	if got := strings.Join(
		lb.PerQuestion[2],
		" ",
	); !strings.Contains(
		got,
		"because",
	) {
		t.Errorf("q2 expected because-clause issue, got %q", got)
	}
	if got := strings.Join(
		lb.PerQuestion[3],
		" ",
	); !strings.Contains(
		got,
		"sentence",
	) {
		t.Errorf("q3 expected extra-sentence issue, got %q", got)
	}
}

func TestAnalyzeLengthBias_NotOnlyWhenDistractorShares(t *testing.T) {
	// A distractor also has a parenthetical -> correct is NOT the only one -> no flag.
	qs := []parse.Question{
		q(
			1,
			0,
			"Correct (with note)",
			"Distractor (also note)",
			"Plain",
			"Plain two",
		),
	}
	lb := AnalyzeLengthBias(qs)
	if j := strings.Join(
		lb.PerQuestion[1],
		" ",
	); strings.Contains(
		j,
		"parenthetical",
	) {
		t.Errorf("should not flag parenthetical when a distractor shares it: %q", j)
	}
}

func TestAnalyzeLengthBias_CleanQuestionNoFlags(t *testing.T) {
	qs := []parse.Question{
		// correct (idx 2) is the SHORTEST — genuinely clean, no qualifiers.
		q(
			1,
			2,
			"Intrusion detection and prevention systems",
			"Network access control list configuration",
			"Segmentation",
			"Role based access control policies",
		),
	}
	lb := AnalyzeLengthBias(qs)
	if lb.CorrectIsLongestByCharCount != 0 || lb.CorrectIsLongestByWordCount != 0 ||
		len(lb.PerQuestion[1]) != 0 {
		t.Errorf("clean question flagged: %+v", lb)
	}
}
