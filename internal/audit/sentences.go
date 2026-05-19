// ===================
// © AngelaMos | 2026
// sentences.go
// ===================

package audit

import "strings"

// abbrev is the set of lowercased tokens (trailing dot stripped) that take a
// period without ending a sentence. Single-letter-segment patterns like
// "u.s" / "e.g" / "a.m" are also detected structurally in isAbbrev.
var abbrev = map[string]bool{
	"mr": true, "mrs": true, "ms": true, "dr": true, "sr": true, "jr": true,
	"st": true, "vs": true, "etc": true, "no": true, "inc": true, "ltd": true,
	"co": true, "corp": true, "dept": true, "fig": true, "al": true,
	"approx": true, "est": true, "gen": true, "min": true, "max": true,
	"sec": true, "std": true, "sp": true, "eg": true, "ie": true,
	"e.g": true, "i.e": true, "u.s": true, "e.u": true, "a.m": true,
	"p.m": true, "ph.d": true,
}

// CountSentences returns the number of sentences in s. It treats '.', '?',
// and '!' as terminators but ignores decimal points (digit.digit),
// abbreviation periods, and collapses runs of terminators. A terminator only
// ends a sentence when followed by a new sentence start (uppercase letter,
// digit, or opening quote/paren) or the end of the string. On ambiguity it
// under-counts, so it never manufactures a false stemLength mismatch.
func CountSentences(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}

	boundaries := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '.' && c != '?' && c != '!' {
			continue
		}

		// Decimal point: digit before AND digit after.
		if c == '.' && i > 0 && i+1 < len(s) && isDigit(s[i-1]) && isDigit(s[i+1]) {
			continue
		}

		// Abbreviation period.
		if c == '.' && isAbbrev(s, i) {
			continue
		}

		// Skip the rest of a terminator run (?!. ...) — one boundary max.
		j := i
		for j+1 < len(s) && isTerminator(s[j+1]) {
			j++
		}

		// Look past closing quotes/brackets and whitespace for the next char.
		k := j + 1
		for k < len(s) && (isSpace(s[k]) || s[k] == '"' || s[k] == '\'' || s[k] == ')' || s[k] == ']') {
			k++
		}

		if k >= len(s) {
			// Terminator closes the final sentence; counted by the +1 below.
			break
		}
		if startsNewSentence(s[k]) {
			boundaries++
		}
		i = j
	}

	return boundaries + 1
}

func isTerminator(b byte) bool { return b == '.' || b == '?' || b == '!' }
func isDigit(b byte) bool      { return b >= '0' && b <= '9' }
func isUpper(b byte) bool      { return b >= 'A' && b <= 'Z' }
func isAlpha(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}

func startsNewSentence(b byte) bool {
	return isUpper(b) || isDigit(b) || b == '"' || b == '\'' || b == '('
}

// isAbbrev reports whether the period at index i terminates an abbreviation
// token (e.g. "Dr.", "etc.", "U.S.", "e.g.") rather than a sentence.
func isAbbrev(s string, i int) bool {
	start := i
	for start > 0 && (isAlpha(s[start-1]) || s[start-1] == '.') {
		start--
	}
	tok := strings.ToLower(s[start : i+1]) // includes trailing '.'
	base := strings.TrimSuffix(tok, ".")
	if base == "" {
		return false
	}
	if abbrev[base] {
		return true
	}

	// A lone single letter before a period is an initial/abbreviation
	// ("U." in "U.S.", middle initials). Sentence-final single capitals
	// are vanishingly rare in exam stems; under-counting is the safe bias.
	if len(base) == 1 && isAlpha(base[0]) {
		return true
	}

	// Structural single-letter-segment pattern: every dot-separated segment
	// is a single letter (covers "u.s", "e.g", "a.m", "e.u", initials).
	if strings.Contains(base, ".") {
		for _, seg := range strings.Split(base, ".") {
			if len(seg) != 1 || !isAlpha(seg[0]) {
				return false
			}
		}
		return true
	}
	return false
}
