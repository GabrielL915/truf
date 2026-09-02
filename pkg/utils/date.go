package utils

import (
	"fmt"
	"strings"
	"time"
)

var englishMonths = map[time.Month]string{
	time.January:   "January",
	time.February:  "February",
	time.March:     "March",
	time.April:     "April",
	time.May:       "May",
	time.June:      "June",
	time.July:      "July",
	time.August:    "August",
	time.September: "September",
	time.October:   "October",
	time.November:  "November",
	time.December:  "December",
}

var shortMonths = map[time.Month]string{
	time.January:   "Jan",
	time.February:  "Feb",
	time.March:     "Mar",
	time.April:     "Apr",
	time.May:       "May",
	time.June:      "Jun",
	time.July:      "Jul",
	time.August:    "Aug",
	time.September: "Sep",
	time.October:   "Oct",
	time.November:  "Nov",
	time.December:  "Dec",
}

func FormatDate(t time.Time) string {
	return t.Format("01/02/2006")
}

func FormatDateShort(t time.Time) string {
	return t.Format("01/02")
}

func FormatMonthYear(t time.Time) string {
	monthName := shortMonths[t.Month()]
	year := t.Format("06")
	return fmt.Sprintf("%s %s", monthName, year)
}

func FormatMonthYearFull(t time.Time) string {
	monthName := englishMonths[t.Month()]
	year := t.Format("2006")
	return fmt.Sprintf("%s %s", monthName, year)
}

func ParseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	t, err := time.Parse("01/02/2006", s)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("01/02/06", s)
	if err == nil {
		return t, nil
	}

	if len(s) == 5 && strings.Count(s, "/") == 1 {
		currentYear := time.Now().Year()
		fullDate := fmt.Sprintf("%s/%d", s, currentYear)
		t, err = time.Parse("01/02/2006", fullDate)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s (expected MM/DD/YYYY)", s)
}

func ValidateDate(s string) bool {
	_, err := ParseDate(s)
	return err == nil
}

func FormatDateInput(s string) string {
	var cleaned strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			cleaned.WriteRune(r)
		}
	}

	digits := cleaned.String()
	if len(digits) == 0 {
		return ""
	}

	var result strings.Builder
	for i, r := range digits {
		if i == 2 || i == 4 {
			result.WriteRune('/')
		}
		if i < 8 {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func GetMonthName(month time.Month) string {
	return englishMonths[month]
}

func GetMonthShort(month time.Month) string {
	return shortMonths[month]
}

func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func FirstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func AddMonths(t time.Time, months int) time.Time {
	first := time.Date(t.Year(), t.Month()+time.Month(months), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	lastDay := first.AddDate(0, 1, -1).Day()
	day := min(t.Day(), lastDay)
	return first.AddDate(0, 0, day-1)
}

func PreviousMonth(t time.Time) time.Time {
	return AddMonths(FirstOfMonth(t), -1)
}

func NextMonth(t time.Time) time.Time {
	return AddMonths(FirstOfMonth(t), 1)
}
