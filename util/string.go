package util

import (
	"strings"
	"unicode"
)

func Truncate(in string, length int) string {
	parts := strings.Split(in, "\n")
	in = parts[0]
	if len(in) <= length {
		if len(parts) > 1 {
			return in + "…"
		} else {
			return in
		}
	}

	runes := []rune(in)

	for i := len(runes) - 1; i > 1; i-- {
		if unicode.IsSpace(runes[i]) && len(string(runes[:i])) < length {
			return string(runes[:i]) + "…"
		}
	}

	return ""
}
