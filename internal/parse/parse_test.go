// ===================
// © AngelaMos | 2026
// parse_test.go
// ===================

package parse

import (
	"os"
	"testing"
)

const minimalJS = `// © AngelaMos | 2026
// sample_test_1.js
db.tests.insertOne({
  "category": "sample",
  "testId": 1,
  "testName": "Sample Practice Test #1",
  "xpPerCorrect": 10,
  "questions": [
    {
      "id": 1,
      "question": "What is 2+2?",
      "options": ["Three", "Four", "Five", "Six"],
      "correctAnswerIndex": 1,
      "explanation": "Two plus two is four.",
      "examTip": "Add carefully.",
      "domain": "1. Arithmetic",
      "questionType": "recall",
      "stemLength": "short",
      "trapType": "adjacent concept",
      "tags": ["math", "addition", "basics"]
    }
  ]
})
`

func TestParseFile_Minimal(t *testing.T) {
	got, err := ParseFile([]byte(minimalJS))
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if got.Category != "sample" {
		t.Errorf("Category = %q, want %q", got.Category, "sample")
	}
	if got.TestID != 1 {
		t.Errorf("TestID = %d, want 1", got.TestID)
	}
	if got.TestName != "Sample Practice Test #1" {
		t.Errorf("TestName = %q", got.TestName)
	}
	if got.XPPerCorrect != 10 {
		t.Errorf("XPPerCorrect = %d, want 10", got.XPPerCorrect)
	}
	if len(got.Questions) != 1 {
		t.Fatalf("len(Questions) = %d, want 1", len(got.Questions))
	}
	q := got.Questions[0]
	if q.ID != 1 {
		t.Errorf("q.ID = %d, want 1", q.ID)
	}
	if q.Question != "What is 2+2?" {
		t.Errorf("q.Question = %q", q.Question)
	}
	if len(q.Options) != 4 || q.Options[1] != "Four" {
		t.Errorf("q.Options = %v", q.Options)
	}
	if q.CorrectAnswerIndex != 1 {
		t.Errorf("q.CorrectAnswerIndex = %d, want 1", q.CorrectAnswerIndex)
	}
	if q.QuestionType != "recall" || q.StemLength != "short" ||
		q.TrapType != "adjacent concept" {
		t.Errorf("enum fields wrong: %+v", q)
	}
	if len(q.Tags) != 3 {
		t.Errorf("q.Tags = %v, want 3", q.Tags)
	}
}

func TestParseFile_BraceAndEscapeInsideStrings(t *testing.T) {
	js := `// © AngelaMos | 2026
// edge_test_1.js
db.tests.insertOne({
  "category": "edge",
  "testId": 2,
  "testName": "Edge #2",
  "xpPerCorrect": 10,
  "questions": [
    {
      "id": 1,
      "question": "Which config block is valid: { \"a\": 1 } or }{ ?",
      "options": ["{ \"a\": 1 }", "}{ broken", "He said \"hi\" loudly", "None"],
      "correctAnswerIndex": 0,
      "explanation": "A JSON object uses { } braces and \"quoted\" keys.",
      "examTip": "Braces matter.",
      "domain": "1. Syntax",
      "questionType": "comprehension",
      "stemLength": "short",
      "trapType": "reversal",
      "tags": ["json", "syntax", "braces"]
    }
  ]
})
`
	got, err := ParseFile([]byte(js))
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if len(got.Questions) != 1 {
		t.Fatalf("len(Questions) = %d, want 1", len(got.Questions))
	}
	q := got.Questions[0]
	if q.Options[0] != `{ "a": 1 }` {
		t.Errorf("option with braces mangled: %q", q.Options[0])
	}
	if q.Options[2] != `He said "hi" loudly` {
		t.Errorf("escaped quotes mangled: %q", q.Options[2])
	}
}

func TestParseFile_RejectsGarbage(t *testing.T) {
	if _, err := ParseFile([]byte("not a test file at all")); err == nil {
		t.Fatal("expected error for non-test input, got nil")
	}
	if _, err := ParseFile([]byte("db.tests.insertOne({ unbalanced ")); err == nil {
		t.Fatal("expected error for unbalanced braces, got nil")
	}
}

func TestParseFile_RealCISSP6(t *testing.T) {
	path := "/home/yoshi/AngelaMos-LLC/CertGames-Content/cissp/tests/cissp_test_6.js"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("real fixture not available: %v", err)
	}
	got, err := ParseFile(raw)
	if err != nil {
		t.Fatalf("ParseFile(real cissp_test_6) error: %v", err)
	}
	if got.Category != "cissp" {
		t.Errorf("Category = %q, want cissp", got.Category)
	}
	if got.TestID != 6 {
		t.Errorf("TestID = %d, want 6", got.TestID)
	}
	if len(got.Questions) != 100 {
		t.Fatalf("len(Questions) = %d, want 100", len(got.Questions))
	}
	q1 := got.Questions[0]
	if q1.ID != 1 || q1.CorrectAnswerIndex != 1 {
		t.Errorf("q1 id/idx wrong: id=%d idx=%d", q1.ID, q1.CorrectAnswerIndex)
	}
	if len(q1.Options) != 4 {
		t.Errorf("q1 options = %d, want 4", len(q1.Options))
	}
	if got.Questions[99].ID != 100 {
		t.Errorf("last question id = %d, want 100", got.Questions[99].ID)
	}
}
