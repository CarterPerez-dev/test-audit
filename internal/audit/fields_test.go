// ===================
// © AngelaMos | 2026
// fields_test.go
// ===================

package audit

import (
	"strings"
	"testing"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

func TestFieldAccuracy_StemLengthMismatch(t *testing.T) {
	qs := []parse.Question{
		{
			ID:                 1,
			Question:           "A team finds a leak. Legal demands action. Which risk category applies here today?",
			Options:            []string{"a", "b", "c", "d"},
			CorrectAnswerIndex: 0,
			StemLength:         "long",
			QuestionType:       "analysis",
			TrapType:           "reversal",
			Tags: []string{
				"a",
				"b",
				"c",
			},
			Explanation: "x",
			ExamTip:     "y",
			Domain:      "1. X",
		},
	}
	fa, _ := FieldFlags(qs)
	got := strings.Join(fa[1], " ")
	if !strings.Contains(got, "stemLength") || !strings.Contains(got, "long") {
		t.Fatalf("expected stemLength mismatch flag, got %q", got)
	}
}

func TestFieldAccuracy_TagsCount(t *testing.T) {
	qs := []parse.Question{
		{
			ID:                 1,
			Question:           "Short one?",
			Options:            []string{"a", "b", "c", "d"},
			CorrectAnswerIndex: 0,
			StemLength:         "short",
			QuestionType:       "recall",
			TrapType:           "reversal",
			Tags: []string{
				"only",
				"two",
			},
			Explanation: "x",
			ExamTip:     "y",
			Domain:      "1. X",
		},
		{
			ID:                 2,
			Question:           "Short two?",
			Options:            []string{"a", "b", "c", "d"},
			CorrectAnswerIndex: 0,
			StemLength:         "short",
			QuestionType:       "recall",
			TrapType:           "reversal",
			Tags: []string{
				"a",
				"b",
				"c",
				"d",
				"e",
			},
			Explanation: "x",
			ExamTip:     "y",
			Domain:      "1. X",
		},
	}
	fa, _ := FieldFlags(qs)
	if !strings.Contains(strings.Join(fa[1], " "), "tags") {
		t.Errorf("q1 expected tags-count flag, got %v", fa[1])
	}
	if !strings.Contains(strings.Join(fa[2], " "), "tags") {
		t.Errorf("q2 expected tags-count flag, got %v", fa[2])
	}
}

func TestFieldAccuracy_StructuralAndEnum(t *testing.T) {
	qs := []parse.Question{
		{
			ID:                 1,
			Question:           "Bad?",
			Options:            []string{"a", "b", "c"},
			CorrectAnswerIndex: 5,
			StemLength:         "tiny",
			QuestionType:       "guess",
			TrapType:           "nope",
			Tags: []string{
				"a",
				"b",
				"c",
			},
			Explanation: "",
			ExamTip:     "y",
			Domain:      "1. X",
		},
	}
	_, structural := FieldFlags(qs)
	j := strings.Join(structural[1], " | ")
	for _, want := range []string{"options", "correctAnswerIndex", "questionType", "stemLength", "trapType", "explanation"} {
		if !strings.Contains(j, want) {
			t.Errorf("expected structural issue mentioning %q, got %q", want, j)
		}
	}
}

func TestFieldAccuracy_CleanQuestion(t *testing.T) {
	qs := []parse.Question{
		{
			ID:                 1,
			Question:           "Which control enforces least privilege effectively?",
			Options:            []string{"RBAC", "DAC", "MAC", "ABAC"},
			CorrectAnswerIndex: 0,
			StemLength:         "short",
			QuestionType:       "recall",
			TrapType:           "adjacent concept",
			Tags: []string{
				"rbac",
				"access control",
				"least privilege",
			},
			Explanation: "x",
			ExamTip:     "y",
			Domain:      "1. X",
		},
	}
	fa, structural := FieldFlags(qs)
	if len(fa[1]) != 0 || len(structural[1]) != 0 {
		t.Errorf("clean question flagged: fa=%v struct=%v", fa[1], structural[1])
	}
}
