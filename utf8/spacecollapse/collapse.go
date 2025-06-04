//go:build !solution

package spacecollapse

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const SPACE rune = ' '

func CollapseSpaces(input string) string {
	sb := strings.Builder{}
	sb.Grow(len(input))
	prevIsSpace := false
	for len(input) > 0 {
		r, size := utf8.DecodeRuneInString(input)
		input = input[size:]
		curIsSpace := unicode.IsSpace(r)
		if prevIsSpace && curIsSpace {
			continue
		}
		if curIsSpace {
			sb.WriteRune(SPACE)
		} else {
			sb.WriteRune(r)
		}
		prevIsSpace = curIsSpace
	}
	return sb.String()
}
