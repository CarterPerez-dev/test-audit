// ===================
// © AngelaMos | 2026
// distribution_test.go
// ===================

package audit

import (
	"strings"
	"testing"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

func mkQ(id int, stem string, opts []string, qtype, trap string) parse.Question {
	return parse.Question{
		ID:                 id,
		Question:           stem,
		Options:            opts,
		CorrectAnswerIndex: 0,
		QuestionType:       qtype,
		TrapType:           trap,
		StemLength:         "short",
		Tags: []string{
			"a",
			"b",
			"c",
		},
		Explanation: "x",
		ExamTip:     "y",
		Domain:      "1. X",
	}
}

func TestDistTrapType_OverCapAndMissing(t *testing.T) {
	var qs []parse.Question
	for i := 1; i <= 30; i++ {
		qs = append(
			qs,
			mkQ(
				i,
				"Short?",
				[]string{"a", "b", "c", "d"},
				"recall",
				"adjacent concept",
			),
		)
	}
	for i := 31; i <= 40; i++ {
		qs = append(
			qs,
			mkQ(i, "Short?", []string{"a", "b", "c", "d"}, "recall", "reversal"),
		)
	}
	counts, flags := DistTrapType(qs)
	if counts["adjacent concept"] != 30 {
		t.Errorf("adjacent concept count = %d, want 30", counts["adjacent concept"])
	}
	j := strings.Join(flags, " | ")
	if !strings.Contains(j, "adjacent concept") || !strings.Contains(j, "25") {
		t.Errorf("expected >25%% cap flag for adjacent concept, got %q", j)
	}
	if !strings.Contains(j, "overgeneralization") {
		t.Errorf("expected missing-type flag for overgeneralization, got %q", j)
	}
}

func TestDistQuestionType_DeviationFlag(t *testing.T) {
	var qs []parse.Question
	for i := 1; i <= 100; i++ {
		qs = append(
			qs,
			mkQ(i, "Short?", []string{"a", "b", "c", "d"}, "recall", "reversal"),
		)
	}
	counts, flags := DistQuestionType(qs, DefaultTargets())
	if counts["recall"] != 100 {
		t.Errorf("recall = %d, want 100", counts["recall"])
	}
	if len(flags) == 0 || !strings.Contains(strings.Join(flags, " "), "recall") {
		t.Errorf("expected recall over-target flag, got %v", flags)
	}
}

func TestDistStemLength_UsesComputedNotDeclared(t *testing.T) {
	// Declared "short" but stems are actually 4 sentences (long).
	var qs []parse.Question
	long := "One thing happens. Two things happen. Three things happen. What is next?"
	for i := 1; i <= 100; i++ {
		q := mkQ(i, long, []string{"a", "b", "c", "d"}, "application", "reversal")
		q.StemLength = "short"
		qs = append(qs, q)
	}
	counts, flags := DistStemLength(qs, DefaultTargets())
	if counts["long"] != 100 || counts["short"] != 0 {
		t.Errorf("computed counts wrong: %+v", counts)
	}
	if !strings.Contains(strings.Join(flags, " "), "long") {
		t.Errorf("expected long-over-target flag, got %v", flags)
	}
}
