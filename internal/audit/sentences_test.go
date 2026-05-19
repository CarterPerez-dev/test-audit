// ===================
// © AngelaMos | 2026
// sentences_test.go
// ===================

package audit

import "testing"

func TestCountSentences(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		// Real CISSP test 6 stems with hand-verified counts.
		{
			"q1 single question",
			"Which security governance concept requires organizational leaders to act with reasonable care to protect information assets from foreseeable harm?",
			1,
		},
		{
			"q2 three sentences",
			"An organization's security team discovers that a recently deployed AI chatbot generates responses that inadvertently reveal confidential contract terms stored in its training corpus. The legal department demands immediate action. Which risk category does this exposure represent?",
			3,
		},
		{
			"q3 two sentences",
			"A multinational financial firm aligns its security program with COBIT to ensure IT governance supports business objectives. What does COBIT primarily provide?",
			2,
		},
		{
			"q5 two sentences",
			"An industry standards investigation determines that a payment processor failed to protect cardholder data as required. Which investigation type applies?",
			2,
		},

		// Adversarial: numbers, currency, percent must NOT split.
		{
			"currency + percent one sentence",
			"If a flood would destroy 60% of $2,000,000 in assets, what is the SLE?",
			1,
		},
		{"decimal number no split", "He paid $1,234.56 today and left.", 1},
		{
			"decimal then new sentence",
			"Costs rose 3.5 percent. Why did that happen?",
			2,
		},
		{
			"version no split",
			"Which process does NIST SP 800-37 define for managing information system risk?",
			1,
		},

		// Adversarial: abbreviations must NOT split.
		{"US abbrev", "The U.S. signed the treaty quickly.", 1},
		{"eg abbrev", "Use a cipher, e.g. AES, for this data.", 1},
		{
			"ie abbrev then sentence",
			"It is symmetric, i.e. one key. Which mode applies?",
			2,
		},
		{"vs abbrev", "Compare TCP vs. UDP for this workload.", 1},

		// Terminator handling.
		{"exclaim then question", "Stop! What should the analyst do next?", 2},
		{"double terminator collapses", "Really?? Yes indeed.", 2},
		{"ellipsis single", "The handshake stalls here...", 1},
		{"no terminator still one", "Pick the best control", 1},
		{"empty is zero", "", 0},
		{"whitespace only is zero", "   \n  ", 0},
		{
			"four sentences long",
			"A breach occurs. Logs are reviewed. Root cause is found. What is the next step?",
			4,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CountSentences(c.in)
			if got != c.want {
				t.Fatalf("CountSentences(%q) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}
