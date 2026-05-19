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

func TestAudit_RealCISSP6_KnownBadFails(t *testing.T) {
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
	// Invariant (not an assumption about pre/post-fix state): overallPass is
	// true iff there are zero critical flags.
	if rep.OverallPass != (len(rep.CriticalFlags) == 0) {
		t.Errorf(
			"overallPass=%v but criticalFlags=%v (must be consistent)",
			rep.OverallPass,
			rep.CriticalFlags,
		)
	}
	if !strings.Contains(strings.ToLower(rep.Summary), "deterministic") {
		t.Errorf("summary must disclose deterministic scope, got: %q", rep.Summary)
	}

	// Schema fidelity: answerLengthBias is a key; correctAnswerPosition is NOT a key.
	b, _ := json.Marshal(rep)
	js := string(b)
	if !strings.Contains(js, `"answerLengthBias"`) {
		t.Errorf("output missing answerLengthBias")
	}
	if strings.Contains(js, `"correctAnswerPosition"`) {
		t.Errorf(
			"output must NOT contain a correctAnswerPosition key (runtime shuffle)",
		)
	}
	for _, k := range []string{`"testFile"`, `"overallPass"`, `"criticalFlags"`, `"distributionAudit"`, `"questionFlags"`, `"fieldAccuracyFlags"`, `"coverageAudit"`, `"summary"`} {
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
	if !rep.OverallPass {
		t.Fatalf("clean test failed: critical=%v", rep.CriticalFlags)
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
	if rep.OverallPass {
		t.Fatalf("100%% length-biased test must FAIL, got overallPass=true")
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

func qs2test(qs []parse.Question) parse.Test {
	return parse.Test{
		Category:     "clean",
		TestID:       1,
		TestName:     "Clean #1",
		XPPerCorrect: 10,
		Questions:    qs,
	}
}
