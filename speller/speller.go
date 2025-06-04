//go:build !solution

package speller

import (
	"fmt"
	"slices"
	"strings"
)

var DIGITS = map[int64]string{
	0: "zero",
	1: "one",
	2: "two",
	3: "three",
	4: "four",
	5: "five",
	6: "six",
	7: "seven",
	8: "eight",
	9: "nine",
}

var TEENS = map[int64]string{
	11: "eleven",
	12: "twelve",
	13: "thirteen",
	14: "fourteen",
	15: "fifteen",
	16: "sixteen",
	17: "seventeen",
	18: "eighteen",
	19: "nineteen",
}

var TENS = map[int64]string{
	10: "ten",
	20: "twenty",
	30: "thirty",
	40: "forty",
	50: "fifty",
	60: "sixty",
	70: "seventy",
	80: "eighty",
	90: "ninety",
}

var THOUSANDS = map[int]string{
	1: "thousand",
	2: "million",
	3: "billion",
}

func Spell(n int64) string {
	if n == 0 {
		return DIGITS[n]
	}

	res := make([]string, 0, 5)

	isNegative := n < 0
	if isNegative {
		n *= -1
	}

	cntThousands := 0
	for n > 0 {
		curBlock := n % 1000
		curBlockStringParts := make([]string, 0, 3)

		if hundreds := curBlock / 100; hundreds > 0 {
			curBlockStringParts = append(curBlockStringParts, fmt.Sprintf("%s hundred", DIGITS[hundreds]))
		}

		tenBlock := curBlock % 100
		switch {
		case tenBlock >= 20 && tenBlock%10 == 0:
			curBlockStringParts = append(curBlockStringParts, TENS[tenBlock])
		case tenBlock >= 20 && tenBlock%10 > 0:
			curBlockStringParts = append(curBlockStringParts, fmt.Sprintf("%s-%s", TENS[tenBlock/10*10], DIGITS[tenBlock%10]))
		case tenBlock > 10:
			curBlockStringParts = append(curBlockStringParts, TEENS[tenBlock])
		case tenBlock == 10:
			curBlockStringParts = append(curBlockStringParts, TENS[tenBlock])
		case tenBlock != 0:
			curBlockStringParts = append(curBlockStringParts, DIGITS[tenBlock])
		}

		if len(curBlockStringParts) > 0 && cntThousands > 0 {
			curBlockStringParts = append(curBlockStringParts, THOUSANDS[cntThousands])
		}

		if len(curBlockStringParts) > 0 {
			res = append(res, strings.Join(curBlockStringParts, " "))
		}

		n /= 1000
		cntThousands++
	}

	if isNegative {
		res = append(res, "minus")
	}

	slices.Reverse(res)
	return strings.Join(res, " ")
}
