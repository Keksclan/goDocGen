package util

import (
	"regexp"
)

var puncRegex = regexp.MustCompile(`(^|.)([,?!:;.]+)([^ \n\r\t])`)

// FixPunctuationSpacing stellt sicher, dass nach Satzzeichen ein Leerzeichen folgt.
func FixPunctuationSpacing(text string) string {
	if text == "" {
		return ""
	}

	return puncRegex.ReplaceAllStringFunc(text, func(match string) string {
		submatches := puncRegex.FindStringSubmatch(match)
		if len(submatches) < 4 {
			return match
		}
		prev := submatches[1]
		punc := submatches[2]
		next := submatches[3]

		// AUSNAHMEN:
		// 1. Dezimalzahlen (z.B. 1.5)
		if punc == "." && next >= "0" && next <= "9" {
			return match
		}

		// 2. URLs und Pfade (z.B. https:// oder C:\)
		if punc == ":" && (next == "/" || next == "\\") {
			return match
		}

		// 3. Abkürzungen und Domains schützen (z.B. test.de oder i.V.)
		// Wenn vor und nach dem Punkt ein kleiner Buchstabe steht, ignorieren wir es.
		if punc == "." && len(prev) > 0 && len(next) > 0 {
			rPrev := rune(prev[0])
			rNext := rune(next[0])
			if rPrev >= 'a' && rPrev <= 'z' && rNext >= 'a' && rNext <= 'z' {
				return match
			}
		}

		return prev + punc + " " + next
	})
}
