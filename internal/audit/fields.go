// ===================
// © AngelaMos | 2026
// fields.go
// ===================

package audit

import (
	"fmt"
	"strings"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

func inSet(v string, set []string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// FieldFlags returns (fieldAccuracy, structural) keyed by question id.
// fieldAccuracy = declared metadata that disagrees with the content
// (stemLength mismatch, tag count). structural = schema/validity breakage.
func FieldFlags(qs []parse.Question) (map[int][]string, map[int][]string) {
	fa := make(map[int][]string)
	st := make(map[int][]string)

	for _, q := range qs {
		// Structural / schema validity.
		if q.ID <= 0 {
			st[q.ID] = append(st[q.ID], "id is missing or invalid")
		}
		if len(q.Options) != 4 {
			st[q.ID] = append(
				st[q.ID],
				fmt.Sprintf("options has %d (must be exactly 4)", len(q.Options)),
			)
		}
		if q.CorrectAnswerIndex < 0 || q.CorrectAnswerIndex >= len(q.Options) {
			st[q.ID] = append(
				st[q.ID],
				fmt.Sprintf(
					"correctAnswerIndex %d out of range for %d options",
					q.CorrectAnswerIndex,
					len(q.Options),
				),
			)
		}
		if !inSet(q.QuestionType, questionTypes) {
			st[q.ID] = append(
				st[q.ID],
				fmt.Sprintf("questionType %q is not a valid enum", q.QuestionType),
			)
		}
		if !inSet(q.StemLength, stemLengths) {
			st[q.ID] = append(
				st[q.ID],
				fmt.Sprintf("stemLength %q is not a valid enum", q.StemLength),
			)
		}
		if !inSet(q.TrapType, trapTypes) {
			st[q.ID] = append(
				st[q.ID],
				fmt.Sprintf("trapType %q is not a valid enum", q.TrapType),
			)
		}
		for _, f := range []struct{ name, val string }{
			{"question", q.Question},
			{"explanation", q.Explanation},
			{"examTip", q.ExamTip},
			{"domain", q.Domain},
		} {
			if strings.TrimSpace(f.val) == "" {
				st[q.ID] = append(st[q.ID], f.name+" is empty")
			}
		}

		// Field accuracy: tags count.
		if n := len(q.Tags); n != 3 && n != 4 {
			fa[q.ID] = append(
				fa[q.ID],
				fmt.Sprintf("tags has %d items (must be exactly 3 or 4)", n),
			)
		}

		// Field accuracy: declared stemLength vs computed (only when the
		// declared value is itself a valid enum — invalid is structural).
		if inSet(q.StemLength, stemLengths) {
			computed := ComputeStemLength(q.Question, q.Options)
			if computed != q.StemLength {
				fa[q.ID] = append(fa[q.ID], fmt.Sprintf(
					"stemLength declared %q but content is %q (%d sentence(s))",
					q.StemLength, computed, CountSentences(q.Question)))
			}
		}
	}
	return fa, st
}
