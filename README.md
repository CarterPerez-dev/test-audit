# test-audit

> Deterministic certification practice-test auditor. One Go binary, zero dependencies.

Replaces the slow, expensive, non-reproducible LLM audit pass with a fast
deterministic engine for everything that is *mechanically verifiable*. It
faithfully implements `defaults/auditor/auditor.md` and emits the exact
audit-report JSON the `fixer.md` step already consumes — drop-in compatible.

```bash
test-audit cissp_test_6.js          # writes cissp_test_6_audit.json next to it
test-audit *.js -o ./audit/         # batch
test-audit cissp_test_6.js --stdout | jq .overallPass
```

## What it checks (deterministic — implemented)

- **#1 priority — answer-length bias.** Per question: correct option longest by
  characters, longest by words, or the *only* option with a parenthetical /
  "because" clause / second sentence / hedge word. Aggregate thresholds:
  ≤25 ok · 26–35 warning · >35 critical (`overallPass: false`).
- **stemLength accuracy** — declared vs computed from a conservative
  sentence counter (abbreviation/decimal/currency aware; under-counts on
  ambiguity so it never manufactures a false mismatch).
- **tag count** (exactly 3 or 4), **structural/enum validity**.
- **Distributions** — stemLength (computed) 50/35/15, questionType
  (declared) 10/15/40/20/15, trapType balance (all six present, none >25%,
  each ≥5). Targets are policy, overridable with `--targets file.json`.
- **Explanation option-position references** (critical — options shuffle at
  runtime), **NOT/EXCEPT framing**, **all/none-of-the-above**, **exam-tip
  sentence count**.

## What it does NOT check (semantic — intentionally out of scope)

Answer correctness, distractor plausibility, questionType/trapType *semantic*
accuracy, explanation pedagogical quality, wording. These need human/LLM
judgment and are disclosed in every report's `summary`. Answer-index/position
distribution is intentionally not audited (UI shuffles options every render).

`auditor.md` overrides the stale audit-output section of `schema.md` where
they conflict (no `correctAnswerPosition`; `answerLengthBias` is the #1 flag).

## Develop

```bash
just test     # go test -race ./...
just build    # bin/test-audit
just ci       # lint + test
```

Zero runtime dependencies. Go 1.24+. AGPL-3.0.
