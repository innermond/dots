package dots

import (
	"regexp"
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

func printable(suspects []*string) error {
	// only white spaces
	pattern := "^\\s+$"
	re := regexp.MustCompile(pattern)

	for _, suspect := range suspects {
		if suspect == nil {
			continue
		}

		match := re.MatchString(*suspect)
		if match {
			return Errorf(EINVALID, "input is empty")
		}

		if has := hasNonPrintable(*suspect); has {
			return Errorf(EINVALID, "input is not a text line")
		}

		*suspect = strings.Trim(*suspect, " ")
	}

	return nil
}
