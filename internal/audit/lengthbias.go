// ===================
// © AngelaMos | 2026
// lengthbias.go
// ===================

package audit

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

// LengthBias is the aggregate + per-question result of the #1 audit check:
// the correct answer must not be the longest / most-qualified option.
type LengthBias struct {
	CorrectIsLongestByCharCount int
	CorrectIsLongestByWordCount int
	QuestionsFlagged            []int // IDs where correct is the strict char-longest
	PerQuestion                 map[int][]string
}

// hedgeWords is exactly auditor.md's list — kept tight to minimize false
// positives (every false flag becomes a wrong fix downstream).
var hedgeWords = map[string]bool{
	"may": true, "typically": true, "generally": true, "often": true,
}

var parenRE = regexp.MustCompile(`\([^)]*\)`)

func AnalyzeLengthBias(qs []parse.Question) LengthBias {
	lb := LengthBias{PerQuestion: make(map[int][]string)}

	for _, qq := range qs {
		if len(qq.Options) == 0 || qq.CorrectAnswerIndex < 0 ||
			qq.CorrectAnswerIndex >= len(qq.Options) {
			continue
		}
		ci := qq.CorrectAnswerIndex
		correct := qq.Options[ci]

		chars := make([]int, len(qq.Options))
		words := make([]int, len(qq.Options))
		for i, o := range qq.Options {
			chars[i] = utf8.RuneCountInString(o)
			words[i] = wordCount(o)
		}

		var issues []string

		if strictMax(chars, ci) {
			lb.CorrectIsLongestByCharCount++
			lb.QuestionsFlagged = append(lb.QuestionsFlagged, qq.ID)
			issues = append(issues, fmt.Sprintf(
				"correct answer is the longest option (character count) — %d vs %s",
				chars[ci], othersList(chars, ci)))
		}
		if strictMax(words, ci) {
			lb.CorrectIsLongestByWordCount++
			issues = append(issues, fmt.Sprintf(
				"correct answer is the longest option (word count) — %d vs %s",
				words[ci], othersList(words, ci)))
		}

		if onlyCorrect(qq, ci, hasParen) {
			issues = append(
				issues,
				"correct answer is the only option with a parenthetical",
			)
		}
		if onlyCorrect(qq, ci, hasBecause) {
			issues = append(
				issues,
				"correct answer is the only option with a 'because' clause",
			)
		}
		if onlyCorrect(qq, ci, hasSecondSentence) {
			issues = append(
				issues,
				"correct answer is the only option with a second sentence",
			)
		}
		if h := soleHedge(qq, ci); h != "" {
			issues = append(
				issues,
				fmt.Sprintf(
					"correct answer is the only option with a hedge word (%q)",
					h,
				),
			)
		}

		if len(issues) > 0 {
			lb.PerQuestion[qq.ID] = issues
		}
		_ = correct
	}

	sort.Ints(lb.QuestionsFlagged)
	return lb
}

// strictMax reports whether vals[idx] is strictly greater than every other
// value (a tie is not "the longest" — picking longest wouldn't reveal it).
func strictMax(vals []int, idx int) bool {
	for i, v := range vals {
		if i != idx && v >= vals[idx] {
			return false
		}
	}
	return true
}

func othersList(vals []int, idx int) string {
	var parts []string
	for i, v := range vals {
		if i != idx {
			parts = append(parts, fmt.Sprintf("%d", v))
		}
	}
	return strings.Join(parts, "/")
}

// onlyCorrect reports whether pred holds for the correct option and for none
// of the distractors.
func onlyCorrect(qq parse.Question, ci int, pred func(string) bool) bool {
	if !pred(qq.Options[ci]) {
		return false
	}
	for i, o := range qq.Options {
		if i != ci && pred(o) {
			return false
		}
	}
	return true
}

func hasParen(s string) bool { return parenRE.MatchString(s) }
func hasBecause(s string) bool {
	return containsWord(strings.ToLower(s), "because")
}
func hasSecondSentence(s string) bool { return CountSentences(s) >= 2 }

func soleHedge(qq parse.Question, ci int) string {
	correctWords := wordSet(qq.Options[ci])
	for w := range hedgeWords {
		if !correctWords[w] {
			continue
		}
		inDistractor := false
		for i, o := range qq.Options {
			if i == ci {
				continue
			}
			if wordSet(o)[w] {
				inDistractor = true
				break
			}
		}
		if !inDistractor {
			return w
		}
	}
	return ""
}

func wordSet(s string) map[string]bool {
	set := make(map[string]bool)
	for _, f := range strings.Fields(strings.ToLower(s)) {
		set[strings.Trim(f, ".,;:!?\"'()[]")] = true
	}
	return set
}

func containsWord(lower, word string) bool {
	for _, f := range strings.Fields(lower) {
		if strings.Trim(f, ".,;:!?\"'()[]") == word {
			return true
		}
	}
	return false
}
