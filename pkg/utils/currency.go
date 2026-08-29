package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatCurrency(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	intPart := addThousandSeparators(strconv.FormatInt(cents/100, 10))
	out := fmt.Sprintf("$%s,%02d", intPart, cents%100)
	if negative {
		return "-" + out
	}
	return out
}

func ParseCurrency(s string) (int64, error) {
	raw := s
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "R$")
	s = strings.TrimPrefix(s, "$")
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, fmt.Errorf("invalid amount: %q", raw)
	}
	if strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("amount must not be negative: %q", raw)
	}

	whole, frac, ok := splitDecimal(s)
	if !ok {
		return 0, fmt.Errorf("invalid amount: %q", raw)
	}

	units, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %q", raw)
	}
	if units > (1<<62)/100 {
		return 0, fmt.Errorf("amount too large: %q", raw)
	}

	switch len(frac) {
	case 0:
		frac = "00"
	case 1:
		frac += "0"
	}
	centsPart, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %q", raw)
	}

	return units*100 + centsPart, nil
}

func splitDecimal(s string) (whole, frac string, ok bool) {
	lastComma := strings.LastIndex(s, ",")
	lastDot := strings.LastIndex(s, ".")

	sep := -1
	switch {
	case lastComma >= 0 && lastDot >= 0:
		if lastComma > lastDot {
			sep = lastComma
		} else {
			sep = lastDot
		}
	case lastComma >= 0:
		sep = lastComma
	case lastDot >= 0:
		if strings.Count(s, ".") == 1 && len(s)-lastDot-1 <= 2 {
			sep = lastDot
		}
	}

	if sep >= 0 {
		whole, frac = s[:sep], s[sep+1:]
	} else {
		whole = s
	}

	if len(frac) > 2 || frac != "" && !allDigits(frac) {
		return "", "", false
	}
	if strings.Contains(frac, ",") || strings.Contains(frac, ".") {
		return "", "", false
	}

	if !validGrouping(whole) {
		return "", "", false
	}
	whole = strings.NewReplacer(".", "", ",", "").Replace(whole)
	if whole == "" {
		whole = "0"
	}
	return whole, frac, true
}

func validGrouping(whole string) bool {
	if whole == "" {
		return true
	}
	groups := strings.FieldsFunc(whole, func(r rune) bool { return r == '.' || r == ',' })
	if len(groups) == 1 {
		return allDigits(groups[0]) && !strings.ContainsAny(whole, ".,")
	}
	for i, g := range groups {
		if !allDigits(g) || len(g) > 3 || len(g) == 0 || (i > 0 && len(g) != 3) {
			return false
		}
	}
	return len(groups) == strings.Count(whole, ".")+strings.Count(whole, ",")+1
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func addThousandSeparators(s string) string {
	var result []rune
	for i, digit := range reverse(s) {
		if i > 0 && i%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, digit)
	}
	return reverse(string(result))
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
