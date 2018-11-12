package util

import (
	"unicode"
)

func Truncate(in string, length int) string {
	if len(in) <= length {
		return in
	}

	runes := []rune(in)

	for i := len(runes) - 1; i > 1; i-- {
		if unicode.IsSpace(runes[i]) && len(string(runes[:i])) < length {
			return string(runes[:i]) + "…"
		}
	}

	return ""
}
