// ===================
// © AngelaMos | 2026
// stemlen_test.go
// ===================

package audit

import "testing"

func TestComputeStemLength(t *testing.T) {
	cases := []struct {
		name    string
		stem    string
		options []string
		want    string
	}{
		{
			"one sentence short options -> short",
			"Which governance concept requires leaders to act with reasonable care?",
			[]string{
				"Due diligence here",
				"Due care applying protections",
				"Risk avoidance",
				"Security policy",
			},
			"short",
		},
		{
			"one sentence numeric options -> short",
			"If a flood destroys 60% of $2,000,000, what is the SLE?",
			[]string{"$200,000", "$600,000", "$1,200,000", "$2,000,000"},
			"short",
		},
		{
			"one sentence but a long option -> medium",
			"Which option best describes defense in depth?",
			[]string{
				"A layered strategy combining multiple independent security controls across people process and technology",
				"A firewall",
				"An antivirus",
				"A password",
			},
			"medium",
		},
		{
			"two sentences -> medium",
			"A firm aligns its program with COBIT for governance. What does COBIT primarily provide?",
			[]string{"A", "B", "C", "D"},
			"medium",
		},
		{
			"three sentences -> medium",
			"A team finds a leak. Legal demands action now. Which risk category applies?",
			[]string{"A", "B", "C", "D"},
			"medium",
		},
		{
			"four sentences -> long",
			"A breach occurs. Logs are reviewed. Root cause is found. What is next?",
			[]string{"A", "B", "C", "D"},
			"long",
		},
		{
			"exactly eight word option still short",
			"What is the best control here?",
			[]string{"one two three four five six seven eight", "a", "b", "c"},
			"short",
		},
		{
			"nine word option breaks short -> medium",
			"What is the best control here?",
			[]string{"one two three four five six seven eight nine", "a", "b", "c"},
			"medium",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ComputeStemLength(c.stem, c.options)
			if got != c.want {
				t.Fatalf(
					"ComputeStemLength(%q, %v) = %q, want %q",
					c.stem,
					c.options,
					got,
					c.want,
				)
			}
		})
	}
}

func TestWordCount(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"   ", 0},
		{"one", 1},
		{"one two three", 3},
		{"  spaced   out  words ", 3},
		{"$1,200,000", 1},
		{"line\nbreak\ttab", 3},
	}
	for _, c := range cases {
		if got := wordCount(c.in); got != c.want {
			t.Errorf("wordCount(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
