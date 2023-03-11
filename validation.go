package dots

import (
	"strings"
	"unicode"
)

// https://gosamples.dev/remove-non-printable-characters/
func filterNonPrintables(text string) string {
	text = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, text)

	return text
}

func hasNonPrintable(text string) bool {
	pp := filterNonPrintables(text)
	return len(pp) != len(text)
}
