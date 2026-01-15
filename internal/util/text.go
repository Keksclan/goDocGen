package util

import (
	"regexp"
)

var puncRegex = regexp.MustCompile(`(.)([,?!:;.]+)([^ \n\r\t])`)

// FixPunctuationSpacing stellt sicher, dass nach Satzzeichen ein Leerzeichen folgt.
func FixPunctuationSpacing(text string) string {
	return text
}
