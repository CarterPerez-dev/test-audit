// ===================
// © AngelaMos | 2026
// distribution.go
// ===================

package audit

import (
	"fmt"
	"math"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}

func deviates(actualPct, targetPct, tolPct float64) bool {
	return math.Abs(actualPct-targetPct) > tolPct
}

// DistStemLength counts the COMPUTED stem length (auditor.md: "count actual
// percentages") — not the declared field — and flags >tol deviation.
func DistStemLength(qs []parse.Question, tg Targets) (map[string]int, []string) {
	counts := map[string]int{"short": 0, "medium": 0, "long": 0}
	for _, q := range qs {
		counts[ComputeStemLength(q.Question, q.Options)]++
	}
	total := len(qs)
	var flags []string
	for _, c := range []struct {
		name   string
		target float64
	}{
		{"short", tg.StemShortPct}, {"medium", tg.StemMediumPct}, {"long", tg.StemLongPct},
	} {
		if p := pct(counts[c.name], total); deviates(p, c.target, tg.StemTolPct) {
			flags = append(flags, fmt.Sprintf(
				"stemLength %s at %.0f%% (%d) deviates from target %.0f%% (tolerance ±%.0f)",
				c.name,
				p,
				counts[c.name],
				c.target,
				tg.StemTolPct,
			))
		}
	}
	return counts, flags
}

// DistQuestionType counts the DECLARED questionType (cognitive demand can't be
// computed deterministically) and flags >tol deviation from target.
func DistQuestionType(qs []parse.Question, tg Targets) (map[string]int, []string) {
	counts := map[string]int{}
	for _, t := range questionTypes {
		counts[t] = 0
	}
	for _, q := range qs {
		counts[q.QuestionType]++
	}
	total := len(qs)
	var flags []string
	for _, c := range []struct {
		name   string
		target float64
	}{
		{"recall", tg.TypeRecallPct},
		{"comprehension", tg.TypeComprehensionPct},
		{"application", tg.TypeApplicationPct},
		{"analysis", tg.TypeAnalysisPct},
		{"evaluation", tg.TypeEvaluationPct},
	} {
		if p := pct(counts[c.name], total); deviates(p, c.target, tg.TypeTolPct) {
			flags = append(flags, fmt.Sprintf(
				"questionType %s at %.0f%% (%d) deviates from target %.0f%% (tolerance ±%.0f)",
				c.name,
				p,
				counts[c.name],
				c.target,
				tg.TypeTolPct,
			))
		}
	}
	return counts, flags
}

// DistTrapType counts the DECLARED trapType and flags: any type missing, any
// type below the minimum, any type over the cap.
func DistTrapType(qs []parse.Question) (map[string]int, []string) {
	tg := DefaultTargets()
	counts := map[string]int{}
	for _, t := range trapTypes {
		counts[t] = 0
	}
	for _, q := range qs {
		counts[q.TrapType]++
	}
	total := len(qs)
	var flags []string
	for _, t := range trapTypes {
		n := counts[t]
		switch {
		case n == 0:
			flags = append(
				flags,
				fmt.Sprintf("trap type %q is missing (all six must appear)", t),
			)
		case n < tg.TrapMinN:
			flags = append(
				flags,
				fmt.Sprintf(
					"trap type %q appears only %d time(s) (minimum %d)",
					t,
					n,
					tg.TrapMinN,
				),
			)
		}
		if p := pct(n, total); p > tg.TrapMaxPct {
			flags = append(flags, fmt.Sprintf(
				"trap type %q at %d (%.0f%%) exceeds the %.0f%% maximum",
				t,
				n,
				p,
				tg.TrapMaxPct,
			))
		}
	}
	return counts, flags
}
