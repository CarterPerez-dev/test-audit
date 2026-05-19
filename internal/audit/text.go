// ===================
// © AngelaMos | 2026
// text.go
// ===================

package audit

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

var optionPosRE = regexp.MustCompile(
	`(?i)\boption [a-d]\b|\bchoice [a-d]\b|\b(?:the )?(?:first|second|third|fourth|1st|2nd|3rd|4th) (?:option|choice|answer)\b|\boption (?:one|two|three|four|1|2|3|4)\b`,
)

// ExplanationFlags flags explanations that reference an option by position.
// Options shuffle at runtime, so positional references become wrong text for
// the user — auditor.md treats this as a critical flag.
func ExplanationFlags(qs []parse.Question) map[int][]string {
	out := make(map[int][]string)
	for _, q := range qs {
		if m := optionPosRE.FindString(q.Explanation); m != "" {
			out[q.ID] = append(
				out[q.ID],
				fmt.Sprintf(
					"explanation references option position (%q) — options shuffle at runtime",
					m,
				),
			)
		}
	}
	return out
}

// QuestionTextFlags flags NOT/EXCEPT stem framing and all/none-of-the-above
// options (anti-giveaway checks).
func QuestionTextFlags(qs []parse.Question) map[int][]string {
	out := make(map[int][]string)
	notRE := regexp.MustCompile(`\bNOT\b|\bEXCEPT\b`)
	for _, q := range qs {
		if notRE.MatchString(q.Question) {
			out[q.ID] = append(out[q.ID],
				`stem uses "NOT"/"EXCEPT" framing — prefer LEAST/WORST phrasing`)
		}
		for _, o := range q.Options {
			lo := strings.ToLower(o)
			if strings.Contains(lo, "all of the above") ||
				strings.Contains(lo, "none of the above") {
				out[q.ID] = append(out[q.ID],
					fmt.Sprintf("option contains %q (not allowed)", o))
			}
		}
	}
	return out
}

// Exam-tip sentence count is intentionally NOT audited (owner decision:
// a 2-3 sentence tip is fine, flagging it is pure noise / busywork).
