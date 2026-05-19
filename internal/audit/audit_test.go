// ===================
// © AngelaMos | 2026
// audit_test.go
// ===================

package audit

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

func TestAudit_RealCISSP6_Invariants(t *testing.T) {
	raw, err := os.ReadFile(
		"/home/yoshi/AngelaMos-LLC/CertGames-Content/cissp/tests/cissp_test_6.js",
	)
	if err != nil {
		t.Skipf("real fixture unavailable: %v", err)
	}
	tst, err := parse.ParseFile(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	rep := Audit("cissp_test_6.js", tst, DefaultTargets())

	if rep.TestFile != "cissp_test_6.js" {
		t.Errorf("TestFile = %q", rep.TestFile)
	}
	sl := rep.DistributionAudit.StemLength
	if sl.Short+sl.Medium+sl.Long != 100 {
		t.Errorf("stemLength counts sum = %d, want 100", sl.Short+sl.Medium+sl.Long)
	}
	if !strings.Contains(strings.ToLower(rep.Summary), "deterministic") {
		t.Errorf("summary must disclose deterministic scope, got: %q", rep.Summary)
	}
	// No pass/fail verdict: summary must not declare PASS/FAIL.
	if strings.Contains(rep.Summary, "PASS") || strings.Contains(rep.Summary, "FAIL") {
		t.Errorf("summary must not contain a pass/fail verdict, got: %q", rep.Summary)
	}

	// Schema fidelity: answerLengthBias is a key; overallPass and
	// correctAnswerPosition are NOT keys (no verdict; runtime shuffle).
	b, _ := json.Marshal(rep)
	js := string(b)
	if !strings.Contains(js, `"answerLengthBias"`) {
		t.Errorf("output missing answerLengthBias")
	}
	if strings.Contains(js, `"correctAnswerPosition"`) {
		t.Errorf("output must NOT contain a correctAnswerPosition key")
	}
	if strings.Contains(js, `"overallPass"`) {
		t.Errorf("output must NOT contain an overallPass key (no pass/fail verdict)")
	}
	for _, k := range []string{`"testFile"`, `"criticalFlags"`, `"distributionAudit"`, `"questionFlags"`, `"fieldAccuracyFlags"`, `"coverageAudit"`, `"summary"`} {
		if !strings.Contains(js, k) {
			t.Errorf("output missing key %s", k)
		}
	}
}

func TestAudit_CleanTestPasses(t *testing.T) {
	var qs []parse.Question
	types := []string{
		"recall",
		"comprehension",
		"application",
		"analysis",
		"evaluation",
	}
	traps := trapTypes
	for i := 0; i < 100; i++ {
		// Correct = idx 0 and the shortest; distractors longer. Single-sentence short.
		qs = append(qs, parse.Question{
			ID:       i + 1,
			Question: "Which control enforces least privilege?",
			Options: []string{
				"RBAC",
				"Discretionary access control lists",
				"Mandatory access labels",
				"Attribute based policies",
			},
			CorrectAnswerIndex: 0,
			Explanation:        "Role based access control maps permissions to roles, unlike the alternatives.",
			ExamTip:            "Roles bundle permissions for least privilege.",
			Domain:             "1. Security and Risk Management",
			QuestionType:       types[i%5],
			StemLength:         "short",
			TrapType:           traps[i%6],
			Tags: []string{
				"rbac",
				"access control",
				"least privilege",
			},
		})
	}
	rep := Audit("clean_test_1.js", qs2test(qs), DefaultTargets())
	if len(rep.CriticalFlags) != 0 {
		t.Fatalf("clean test produced critical flags: %v", rep.CriticalFlags)
	}
	if rep.DistributionAudit.AnswerLengthBias.CorrectIsLongestByCharCount != 0 {
		t.Errorf("clean test should have 0 char-longest, got %d",
			rep.DistributionAudit.AnswerLengthBias.CorrectIsLongestByCharCount)
	}
	// Schema fidelity: criticalFlags must serialize as [] not null when empty
	// (fixer.md and schema.md mandate an array — null breaks the consumer).
	b, _ := json.Marshal(rep)
	if strings.Contains(string(b), `"criticalFlags":null`) {
		t.Errorf("empty criticalFlags marshaled as null, must be []")
	}
}

func TestAudit_GuaranteedLengthBiasFails(t *testing.T) {
	// Every question: correct option is strictly the longest by far.
	var qs []parse.Question
	for i := 0; i < 100; i++ {
		qs = append(qs, parse.Question{
			ID:       i + 1,
			Question: "Which is the best control?",
			Options: []string{
				"Short A",
				"This correct answer is deliberately and obviously the single longest option on the entire question by a wide margin",
				"Short C",
				"Short D",
			},
			CorrectAnswerIndex: 1,
			Explanation:        "The correct choice fits the scenario better than the others.",
			ExamTip:            "Match the control to the requirement.",
			Domain:             "1. Security and Risk Management",
			QuestionType:       "application",
			StemLength:         "short",
			TrapType:           trapTypes[i%6],
			Tags:               []string{"a", "b", "c"},
		})
	}
	rep := Audit("bias_test_1.js", qs2test(qs), DefaultTargets())
	if len(rep.CriticalFlags) == 0 {
		t.Fatalf("100%% length-biased test must produce a critical flag, got none")
	}
	if rep.DistributionAudit.AnswerLengthBias.CorrectIsLongestByCharCount != 100 {
		t.Errorf(
			"expected 100 char-longest, got %d",
			rep.DistributionAudit.AnswerLengthBias.CorrectIsLongestByCharCount,
		)
	}
	found := false
	for _, c := range rep.CriticalFlags {
		if strings.Contains(c, "consistently longer") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the severe >35 critical flag, got: %v", rep.CriticalFlags)
	}
}

func TestAudit_DoesNotFlagExamTipLength(t *testing.T) {
	// Owner decision: exam-tip sentence count is NOT audited (too noisy,
	// no value). Multi-sentence tips must produce zero examTipFlags.
	var qs []parse.Question
	for i := 0; i < 10; i++ {
		qs = append(qs, parse.Question{
			ID: i + 1, Question: "Which control enforces least privilege?",
			Options:            []string{"RBAC", "Discretionary access control lists", "Mandatory access labels", "Attribute based policies"},
			CorrectAnswerIndex: 0,
			Explanation:        "Role based access control maps permissions to roles unlike the others.",
			ExamTip:            "First sentence here. Second sentence too. Even a third one.",
			Domain:             "1. Security and Risk Management",
			QuestionType:       "recall", StemLength: "short", TrapType: trapTypes[i%6],
			Tags: []string{"a", "b", "c"},
		})
	}
	rep := Audit("tip_test_1.js", qs2test(qs), DefaultTargets())
	if len(rep.ExamTipFlags) != 0 {
		t.Fatalf("exam-tip length must not be audited, got %d flags: %+v",
			len(rep.ExamTipFlags), rep.ExamTipFlags)
	}
}

// biasedSet builds n questions where the first k have the correct option
// strictly longest by chars; the rest are clean.
func biasedSet(n, k int) []parse.Question {
	var qs []parse.Question
	for i := 0; i < n; i++ {
		correctLong := i < k
		opts := []string{"Short A", "Short B", "Short C", "Short D"}
		ci := 0
		if correctLong {
			opts = []string{
				"This correct option is deliberately by far the single longest one here",
				"Short B", "Short C", "Short D",
			}
		}
		qs = append(qs, parse.Question{
			ID: i + 1, Question: "Which control is best here?", Options: opts,
			CorrectAnswerIndex: ci,
			Explanation:        "The correct choice fits better than the others.",
			ExamTip:            "Match control to requirement.",
			Domain:             "1. Security and Risk Management",
			QuestionType:       "application", StemLength: "short",
			TrapType: trapTypes[i%6], Tags: []string{"a", "b", "c"},
		})
	}
	return qs
}

func TestAudit_LengthBiasScalesWithTestSize(t *testing.T) {
	cases := []struct {
		name         string
		n, k         int
		wantCritical bool
		wantSevere   bool
	}{
		{"25q 40% biased -> severe critical", 25, 10, true, true},
		{"25q 24% biased -> no critical", 25, 6, false, false},
		{"100q 35 -> warn critical not severe (boundary)", 100, 35, true, false},
		{"100q 36 -> severe (boundary)", 100, 36, true, true},
		{"100q 25 -> no critical (boundary)", 100, 25, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rep := Audit("x.js", qs2test(biasedSet(c.n, c.k)), DefaultTargets())
			if (len(rep.CriticalFlags) > 0) != c.wantCritical {
				t.Errorf("criticalFlags=%v want non-empty=%v", rep.CriticalFlags, c.wantCritical)
			}
			severe := false
			for _, f := range rep.CriticalFlags {
				if strings.Contains(f, "consistently longer") {
					severe = true
				}
			}
			if severe != c.wantSevere {
				t.Errorf("severe=%v want %v (critical=%v)", severe, c.wantSevere, rep.CriticalFlags)
			}
		})
	}
}

func TestAudit_TrapMinScalesDown_25q(t *testing.T) {
	// 25 questions, all 6 trap types present (~4 each). Must NOT flag
	// "below minimum" — the absolute 5 doesn't apply at N=25.
	qs := biasedSet(25, 0)
	rep := Audit("x.js", qs2test(qs), DefaultTargets())
	for _, f := range rep.DistributionAudit.TrapType.Flags {
		if strings.Contains(f, "minimum") {
			t.Errorf("trap min should scale down for 25q, got: %q", f)
		}
	}
}

func qs2test(qs []parse.Question) parse.Test {
	return parse.Test{
		Category:     "clean",
		TestID:       1,
		TestName:     "Clean #1",
		XPPerCorrect: 10,
		Questions:    qs,
	}
}
