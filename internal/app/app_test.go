// ===================
// © AngelaMos | 2026
// app_test.go
// ===================

package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleJS = `// © AngelaMos | 2026
// sample_test_1.js
db.tests.insertOne({
  "category": "sample",
  "testId": 1,
  "testName": "Sample Practice Test #1",
  "xpPerCorrect": 10,
  "questions": [
    {
      "id": 1,
      "question": "Which control enforces least privilege?",
      "options": ["RBAC", "Discretionary access lists", "Mandatory labels", "Attribute policies"],
      "correctAnswerIndex": 0,
      "explanation": "Role based access control maps permissions to roles unlike the others.",
      "examTip": "Roles bundle permissions.",
      "domain": "1. Security and Risk Management",
      "questionType": "recall",
      "stemLength": "short",
      "trapType": "adjacent concept",
      "tags": ["rbac", "access control", "least privilege"]
    }
  ]
})
`

func TestRun_WritesAuditFileNextToInput(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "sample_test_1.js")
	if err := os.WriteFile(in, []byte(sampleJS), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{in}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr.String())
	}

	out := filepath.Join(dir, "sample_test_1_audit.json")
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("audit file not written: %v", err)
	}
	var rep map[string]any
	if err := json.Unmarshal(b, &rep); err != nil {
		t.Fatalf("audit json invalid: %v", err)
	}
	if rep["testFile"] != "sample_test_1.js" {
		t.Errorf("testFile = %v", rep["testFile"])
	}
	if _, ok := rep["answerLengthBias"]; ok {
		t.Errorf(
			"answerLengthBias should be nested under distributionAudit, not top-level",
		)
	}
	if _, ok := rep["overallPass"]; ok {
		t.Errorf("overallPass must be absent — there is no pass/fail verdict")
	}
	if _, ok := rep["criticalFlags"]; !ok {
		t.Errorf("missing criticalFlags")
	}
}

func TestRun_StdoutFlag(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "s_test_1.js")
	os.WriteFile(in, []byte(sampleJS), 0o644)

	var stdout, stderr bytes.Buffer
	code := Run([]string{in, "--stdout"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"testFile"`) {
		t.Errorf("stdout missing report json: %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "s_test_1_audit.json")); err == nil {
		t.Errorf("--stdout must not write a file")
	}
}

func TestRun_NoArgsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(nil, strings.NewReader(""), &stdout, &stderr); code != 2 {
		t.Errorf("no args exit = %d, want 2", code)
	}
}

func TestRun_HelpExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(
		[]string{"-h"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	); code != 0 {
		t.Errorf("-h exit = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "test-audit") {
		t.Errorf("help text missing")
	}
}

func TestRun_MissingFileFails(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(
		[]string{"/nope/does_not_exist.js"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	); code != 1 {
		t.Errorf("missing file exit = %d, want 1", code)
	}
}
