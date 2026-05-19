// ===================
// © AngelaMos | 2026
// targets.go
// ===================

package audit

// Targets holds the distribution policy. Mechanism (counting) is separate
// from policy (these numbers) so a wrong target can never silently corrupt
// every audit — override via the loader, defaults below are auditor.md's.
type Targets struct {
	StemShortPct  float64
	StemMediumPct float64
	StemLongPct   float64
	StemTolPct    float64 // flag if |actual% - target%| exceeds this (points)

	TypeRecallPct        float64
	TypeComprehensionPct float64
	TypeApplicationPct   float64
	TypeAnalysisPct      float64
	TypeEvaluationPct    float64
	TypeTolPct           float64

	TrapMaxPct float64 // no single trap type may exceed this % of questions
	TrapMinN   int     // every trap type must appear at least this many times
}

// DefaultTargets is auditor.md's spec: stem 50/35/15 (±10pt), questionType
// 10/15/40/20/15 (±5pt), trap none >25%, each ≥5.
func DefaultTargets() Targets {
	return Targets{
		StemShortPct: 50, StemMediumPct: 35, StemLongPct: 15, StemTolPct: 10,
		TypeRecallPct: 10, TypeComprehensionPct: 15, TypeApplicationPct: 40,
		TypeAnalysisPct: 20, TypeEvaluationPct: 15, TypeTolPct: 5,
		TrapMaxPct: 25, TrapMinN: 5,
	}
}

var trapTypes = []string{
	"reversal", "adjacent concept", "overgeneralization",
	"partial knowledge", "plausible but unrelated", "correct-in-different-scenario",
}

var questionTypes = []string{
	"recall",
	"comprehension",
	"application",
	"analysis",
	"evaluation",
}

var stemLengths = []string{"short", "medium", "long"}
