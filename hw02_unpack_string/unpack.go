package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {
	if str == "" {
		return "", nil
	}
	runes := []rune(str)

	var strBuilder strings.Builder

	for i := 0; i < len(runes); i++ {
		curr := runes[i]
		if curr == '\\' {
			if i+1 > len(runes) {
				return "", ErrInvalidString
			}
			next := runes[i+1]
			if next != '\\' && !unicode.IsDigit(next) {
				return "", ErrInvalidString
			}
			strBuilder.WriteRune(next)
			i++
			if i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
				count := int(runes[i+1] - '0')
				symbol := next
				strBuilder.WriteString(strings.Repeat(string(symbol), count-1))
				i++
			}
			continue
		}
		if unicode.IsDigit(curr) {
			return "", ErrInvalidString
		}
		count := 1
		if i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
			count = int(runes[i+1] - '0')
			i++
		}
		strBuilder.WriteString(strings.Repeat(string(curr), count))
	}

	return strBuilder.String(), nil
}
