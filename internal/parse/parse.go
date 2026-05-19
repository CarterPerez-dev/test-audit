// ===================
// © AngelaMos | 2026
// parse.go
// ===================

package parse

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Question struct {
	ID                 int      `json:"id"`
	Question           string   `json:"question"`
	Options            []string `json:"options"`
	CorrectAnswerIndex int      `json:"correctAnswerIndex"`
	Explanation        string   `json:"explanation"`
	ExamTip            string   `json:"examTip"`
	Domain             string   `json:"domain"`
	QuestionType       string   `json:"questionType"`
	StemLength         string   `json:"stemLength"`
	TrapType           string   `json:"trapType"`
	Tags               []string `json:"tags"`
}

type Test struct {
	Category     string     `json:"category"`
	TestID       int        `json:"testId"`
	TestName     string     `json:"testName"`
	XPPerCorrect int        `json:"xpPerCorrect"`
	Questions    []Question `json:"questions"`
}

const marker = "db.tests.insertOne("

// ParseFile extracts the JSON object from a `db.tests.insertOne({...})`
// JavaScript test file and unmarshals it into a Test. The two-line
// `// © AngelaMos | 2026` header and any surrounding whitespace are ignored.
func ParseFile(raw []byte) (Test, error) {
	s := string(raw)

	m := strings.Index(s, marker)
	if m < 0 {
		return Test{}, fmt.Errorf(
			"not a test file: %q wrapper not found",
			strings.TrimSuffix(marker, "("),
		)
	}

	obj, err := extractObject(s[m+len(marker):])
	if err != nil {
		return Test{}, err
	}

	var t Test
	if err := json.Unmarshal([]byte(obj), &t); err != nil {
		return Test{}, fmt.Errorf("parse test JSON: %w", err)
	}
	return t, nil
}

// extractObject returns the first balanced {...} object in s, scanning in a
// string-aware way so that braces, quotes, and escapes inside JSON string
// values never throw off the brace depth count.
func extractObject(s string) (string, error) {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return "", fmt.Errorf("no opening brace after insertOne(")
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		c := s[i]

		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}

		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], nil
			}
		}
	}

	return "", fmt.Errorf("unbalanced braces: object never closes")
}
