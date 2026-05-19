// ===================
// © AngelaMos | 2026
// text_test.go
// ===================

package audit

import (
	"strings"
	"testing"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

func TestExplanationFlags_OptionPosition(t *testing.T) {
	qs := []parse.Question{
		{
			ID:          1,
			Explanation: "Option A is correct because it scales. The third choice is a distractor.",
		},
		{
			ID:          2,
			Explanation: "Due care means acting reasonably; due diligence is investigation.",
		},
		{ID: 3, Explanation: "The second option fails because the key is static."},
	}
	f := ExplanationFlags(qs)
	if len(f[1]) == 0 {
		t.Errorf("q1 should flag 'Option A' / 'third choice'")
	}
	if len(f[2]) != 0 {
		t.Errorf("q2 is clean prose, should not flag: %v", f[2])
	}
	if len(f[3]) == 0 {
		t.Errorf("q3 should flag 'the second option'")
	}
}

func TestQuestionTextFlags_NotFramingAndAboveOptions(t *testing.T) {
	qs := []parse.Question{
		{
			ID: 1, Question: "Which of the following is NOT a symmetric cipher?",
			Options: []string{"AES", "DES", "RSA", "3DES"},
		},
		{
			ID: 2, Question: "Which control is best here?",
			Options: []string{"RBAC", "All of the above", "MAC", "DAC"},
		},
		{
			ID: 3, Question: "Which protocol secures web traffic?",
			Options: []string{"TLS", "FTP", "Telnet", "SMTP"},
		},
	}
	f := QuestionTextFlags(qs)
	if !strings.Contains(strings.Join(f[1], " "), "NOT") {
		t.Errorf("q1 should flag NOT-framing, got %v", f[1])
	}
	if len(f[2]) == 0 {
		t.Errorf("q2 should flag 'All of the above'")
	}
	if len(f[3]) != 0 {
		t.Errorf("q3 is clean, should not flag: %v", f[3])
	}
}
