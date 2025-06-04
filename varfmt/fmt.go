//go:build !solution

package varfmt

import (
	"fmt"
	"strconv"
	"strings"
)

func Sprintf(format string, args ...interface{}) string {
	// prepare string for each arg from args
	stringArgs := make([]string, len(args))
	for idx, arg := range args {
		stringArgs[idx] = fmt.Sprintf("%v", arg)
	}

	// result string builder
	sb := strings.Builder{}
	sb.Grow(len(format))

	cntBraces := 0
	isBrace := false
	curBrace := make([]rune, 256)
	curBraceWritePos := 0

	for _, ch := range format {
		if ch == '{' {
			isBrace = true
			continue
		}
		if ch == '}' {
			if curBraceWritePos > 0 {
				pos, err := strconv.Atoi(string(curBrace[:curBraceWritePos]))
				if err != nil {
					panic("wrong format string")
				}
				sb.WriteString(stringArgs[pos])
			} else {
				sb.WriteString(stringArgs[cntBraces])
			}

			cntBraces++
			isBrace = false
			curBraceWritePos = 0
			continue
		}
		if isBrace {
			curBrace[curBraceWritePos] = ch
			curBraceWritePos++
		} else {
			sb.WriteRune(ch)
		}
	}

	return sb.String()
}
