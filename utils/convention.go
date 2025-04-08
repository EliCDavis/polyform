package utils

import (
	"strings"
	"unicode"
)

func CamelCaseToSpaceCase(s string) string {
	var result strings.Builder
	dontSpaceNextRune := false
	for _, r := range s {
		if unicode.IsUpper(r) && result.Len() > 0 && !dontSpaceNextRune {
			result.WriteRune(' ')
		} else if unicode.IsDigit(r) && !dontSpaceNextRune {
			result.WriteRune(' ')
		}

		result.WriteRune(r)
		dontSpaceNextRune = unicode.IsUpper(r) || unicode.IsDigit(r)
	}
	return result.String()
}
