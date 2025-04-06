package utils

import (
	"strings"
	"unicode"
)

func CamelCaseToSpaceCase(s string) string {
	var result strings.Builder
	previousCapital := false
	for _, r := range s {
		if unicode.IsUpper(r) && result.Len() > 0 && !previousCapital {
			result.WriteRune(' ')
		}
		result.WriteRune(r)

		previousCapital = unicode.IsUpper(r)
	}
	return result.String()
}
