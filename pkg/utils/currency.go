package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatCurrency(amount float64) string {
	str := fmt.Sprintf("%.2f", amount)

	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]

	intPart = addThousandSeparators(intPart)

	return fmt.Sprintf("$%s.%s", intPart, decPart)
}

func ParseCurrency(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, nil
	}

	s = strings.ReplaceAll(s, ",", "")

	return strconv.ParseFloat(s, 64)
}

func addThousandSeparators(s string) string {
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	var result []rune
	for i, digit := range reverse(s) {
		if i > 0 && i%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, digit)
	}

	formatted := string(reverse(string(result)))
	if negative {
		formatted = "-" + formatted
	}

	return formatted
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func FormatCurrencyInput(s string) string {
	var cleaned strings.Builder
	hasDecimal := false

	for _, r := range s {
		if r >= '0' && r <= '9' {
			cleaned.WriteRune(r)
		} else if r == '.' && !hasDecimal {
			cleaned.WriteRune('.')
			hasDecimal = true
		}
	}

	return cleaned.String()
}
