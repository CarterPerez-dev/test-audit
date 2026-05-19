// ===================
// © AngelaMos | 2026
// stemlen.go
// ===================

package audit

import "strings"

// shortMaxOptionWords is the per-option word ceiling for a "short" question
// (auditor.md: short = single-sentence stem, all options ≤8 words).
const shortMaxOptionWords = 8

// ComputeStemLength derives the true stemLength category from the stem text
// and options, independent of the declared field:
//
//	4+ sentences            -> long
//	2-3 sentences           -> medium
//	1 sentence, all options ≤8 words -> short
//	1 sentence, an option >8 words   -> medium (scenario-weight options)
func ComputeStemLength(stem string, options []string) string {
	n := CountSentences(stem)
	switch {
	case n >= 4:
		return "long"
	case n >= 2:
		return "medium"
	default:
		for _, o := range options {
			if wordCount(o) > shortMaxOptionWords {
				return "medium"
			}
		}
		return "short"
	}
}

func wordCount(s string) int {
	return len(strings.Fields(s))
}
