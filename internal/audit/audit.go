// ===================
// © AngelaMos | 2026
// audit.go
// ===================

package audit

import (
	"fmt"
	"sort"

	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

type QFlag struct {
	QuestionID int      `json:"questionId"`
	Issues     []string `json:"issues"`
}

type StemLengthDist struct {
	Short  int      `json:"short"`
	Medium int      `json:"medium"`
	Long   int      `json:"long"`
	Flags  []string `json:"flags"`
}

type QuestionTypeDist struct {
	Recall        int      `json:"recall"`
	Comprehension int      `json:"comprehension"`
	Application   int      `json:"application"`
	Analysis      int      `json:"analysis"`
	Evaluation    int      `json:"evaluation"`
	Flags         []string `json:"flags"`
}

type TrapTypeDist struct {
	Counts map[string]int `json:"counts"`
	Flags  []string       `json:"flags"`
}

type AnswerLengthBias struct {
	CorrectIsLongestByCharCount int      `json:"correctIsLongestByCharCount"`
	CorrectIsLongestByWordCount int      `json:"correctIsLongestByWordCount"`
	QuestionsFlagged            []int    `json:"questionsFlagged"`
	Flags                       []string `json:"flags"`
}

type DistributionAudit struct {
	StemLength       StemLengthDist   `json:"stemLength"`
	QuestionType     QuestionTypeDist `json:"questionType"`
	TrapType         TrapTypeDist     `json:"trapType"`
	AnswerLengthBias AnswerLengthBias `json:"answerLengthBias"`
}

type CoverageAudit struct {
	DomainCoverage map[string]int `json:"domainCoverage"`
	Flags          []string       `json:"flags"`
}

type Report struct {
	TestFile           string            `json:"testFile"`
	CriticalFlags      []string          `json:"criticalFlags"`
	DistributionAudit  DistributionAudit `json:"distributionAudit"`
	QuestionFlags      []QFlag           `json:"questionFlags"`
	ExplanationFlags   []QFlag           `json:"explanationFlags"`
	ExamTipFlags       []QFlag           `json:"examTipFlags"`
	FieldAccuracyFlags []QFlag           `json:"fieldAccuracyFlags"`
	CoverageAudit      CoverageAudit     `json:"coverageAudit"`
	Summary            string            `json:"summary"`
}

const summaryScope = "This is a DETERMINISTIC audit. Implemented (mechanically verifiable): " +
	"answer-length bias (#1: char/word longest, sole parenthetical/because/2nd-sentence/hedge), " +
	"stemLength accuracy vs computed sentence count, tag count, stemLength/questionType/trapType " +
	"distributions, trap-type balance, explanation option-position references, NOT/EXCEPT framing, " +
	"all/none-of-the-above options, structural/enum validity. " +
	"NOT checked (intentionally out of scope): exam-tip sentence count (a 2-3 sentence tip is " +
	"fine — owner decision), semantic correctness of answers, distractor plausibility, " +
	"questionType/trapType semantic accuracy, explanation pedagogical quality, wording quality. " +
	"Answer-index/answer-position distribution is not audited (options shuffle at runtime)."

// Audit runs every deterministic rule and assembles the auditor.md report.
// overallPass is true only when there are zero critical flags.
func Audit(testFile string, t parse.Test, tg Targets) Report {
	qs := t.Questions

	slCounts, slFlags := DistStemLength(qs, tg)
	qtCounts, qtFlags := DistQuestionType(qs, tg)
	trCounts, trFlags := DistTrapType(qs)
	lb := AnalyzeLengthBias(qs)
	fa, structural := FieldFlags(qs)
	explFlags := ExplanationFlags(qs)
	qTextFlags := QuestionTextFlags(qs)

	var critical []string
	var albFlags []string

	// Length-bias bands are RELATIVE to test size so they hold for 25q,
	// 100q, or any N. ratio ≤25% ok · >25%–35% warning · >35% severe.
	// At N=100 this is byte-identical to the old absolute 25/26-35/>35.
	if n, total := lb.CorrectIsLongestByCharCount, len(qs); total > 0 {
		ratio := float64(n) / float64(total)
		switch {
		case ratio > 0.35:
			msg := fmt.Sprintf(
				"correct answer consistently longer than distractors — %d/%d (%.0f%%) questions where the correct answer is the longest by character count (severe; >35%%)",
				n, total, ratio*100)
			critical = append(critical, msg)
			albFlags = append(albFlags, msg)
		case ratio > 0.25:
			msg := fmt.Sprintf(
				"lengthBias warning — %d/%d (%.0f%%) questions where the correct answer is the longest by character count (concerning; 25-35%% band)",
				n, total, ratio*100)
			critical = append(critical, msg)
			albFlags = append(albFlags, msg)
		}
	}

	if len(explFlags) > 0 {
		critical = append(
			critical,
			fmt.Sprintf(
				"option position referenced in explanations (%d question(s))",
				len(explFlags),
			),
		)
	}

	// Per-question questionFlags: length bias + structural + question-text.
	qIssues := map[int][]string{}
	for id, v := range lb.PerQuestion {
		qIssues[id] = append(qIssues[id], v...)
	}
	for id, v := range structural {
		qIssues[id] = append(qIssues[id], v...)
	}
	for id, v := range qTextFlags {
		qIssues[id] = append(qIssues[id], v...)
	}

	rep := Report{
		TestFile:      testFile,
		CriticalFlags: nonNil(critical),
		DistributionAudit: DistributionAudit{
			StemLength: StemLengthDist{
				Short:  slCounts["short"],
				Medium: slCounts["medium"],
				Long:   slCounts["long"],
				Flags:  nonNil(slFlags),
			},
			QuestionType: QuestionTypeDist{
				Recall: qtCounts["recall"], Comprehension: qtCounts["comprehension"],
				Application: qtCounts["application"], Analysis: qtCounts["analysis"],
				Evaluation: qtCounts["evaluation"], Flags: nonNil(qtFlags),
			},
			TrapType: TrapTypeDist{Counts: trCounts, Flags: nonNil(trFlags)},
			AnswerLengthBias: AnswerLengthBias{
				CorrectIsLongestByCharCount: lb.CorrectIsLongestByCharCount,
				CorrectIsLongestByWordCount: lb.CorrectIsLongestByWordCount,
				QuestionsFlagged:            nonNilInts(lb.QuestionsFlagged),
				Flags:                       nonNil(albFlags),
			},
		},
		QuestionFlags:      toQFlags(qIssues),
		ExplanationFlags:   toQFlags(explFlags),
		ExamTipFlags:       []QFlag{},
		FieldAccuracyFlags: toQFlags(fa),
		CoverageAudit:      domainCoverage(qs),
		Summary:            "",
	}
	// No pass/fail verdict by design: this is a workflow step, not a gate.
	// Every test has flags and every flag gets fixed downstream; a boolean
	// "pass" is meaningless and risks the fixer deprioritizing real issues.
	// criticalFlags remains as the highest-priority "fix these first" bucket.
	rep.Summary = fmt.Sprintf(
		"%d questions audited — %d high-priority flag(s), %d flagged question(s); correct-is-longest by char %d / by word %d. Fix every flagged item; there is no pass/fail. %s",
		len(qs),
		len(critical),
		len(rep.QuestionFlags),
		lb.CorrectIsLongestByCharCount,
		lb.CorrectIsLongestByWordCount,
		summaryScope,
	)
	return rep
}

func domainCoverage(qs []parse.Question) CoverageAudit {
	counts := map[string]int{}
	for _, q := range qs {
		counts[q.Domain]++
	}
	return CoverageAudit{
		DomainCoverage: counts,
		Flags: []string{
			"coverage audit skipped — no exam objectives provided (pass --objectives to enable weighting checks)",
		},
	}
}

func toQFlags(m map[int][]string) []QFlag {
	ids := make([]int, 0, len(m))
	for id := range m {
		if len(m[id]) > 0 {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	out := make([]QFlag, 0, len(ids))
	for _, id := range ids {
		out = append(out, QFlag{QuestionID: id, Issues: m[id]})
	}
	return out
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func nonNilInts(s []int) []int {
	if s == nil {
		return []int{}
	}
	return s
}
