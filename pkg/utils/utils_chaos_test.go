package utils_test

import (
	"strings"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/pkg/utils"
)

func TestChaosParseCurrencyRejectsNonFiniteInput(t *testing.T) {
	for _, in := range []string{"NaN", "Inf", "-Inf", "inf", "1e400", "1e3"} {
		if v, err := utils.ParseCurrency(in); err == nil {
			t.Errorf("ParseCurrency(%q) = %v, want error", in, v)
		}
	}
}

func TestChaosParseCurrencyRejectsGarbage(t *testing.T) {
	for _, in := range []string{"", "1.2.3", "abc", "$$", "1,5,", "１２", "-5", "1,234", "1,2,3", "1.234.5", "12.34.56", "1..5", ",5,"} {
		if v, err := utils.ParseCurrency(in); err == nil {
			t.Errorf("ParseCurrency(%q) accepted as %d", in, v)
		}
	}
}

func TestChaosParseCurrencyBrazilianAndUSInputs(t *testing.T) {
	cases := map[string]int64{
		"1,5":          150,
		"1,50":         150,
		"1.234,56":     123456,
		"1234,56":      123456,
		"1.234.567,89": 123456789,
		"1234.56":      123456,
		"1,234.56":     123456,
		"1.234":        123400,
		"10":           1000,
		"0":            0,
		"0,07":         7,
		"$ 42":         4200,
		"R$ 42,10":     4210,
		"  7,3  ":      730,
	}
	for in, want := range cases {
		got, err := utils.ParseCurrency(in)
		if err != nil || got != want {
			t.Errorf("ParseCurrency(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
}

func TestChaosFormatCurrencyBrazilianStyle(t *testing.T) {
	cases := map[int64]string{
		0:          "$0,00",
		7:          "$0,07",
		150:        "$1,50",
		123456:     "$1.234,56",
		123456789:  "$1.234.567,89",
		-1:         "-$0,01",
		-123456:    "-$1.234,56",
		1 << 62:    "$46.116.860.184.273.879,04",
		-(1 << 62): "-$46.116.860.184.273.879,04",
	}
	for in, want := range cases {
		if got := utils.FormatCurrency(in); got != want {
			t.Errorf("FormatCurrency(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestChaosFormatCurrencyRoundTripsParse(t *testing.T) {
	for _, v := range []int64{0, 1, 10, 150, 123456789, 1e12} {
		back, err := utils.ParseCurrency(utils.FormatCurrency(v))
		if err != nil || back != v {
			t.Errorf("round trip %d -> %q -> %d (%v)", v, utils.FormatCurrency(v), back, err)
		}
	}
}

func TestChaosParseDateLeapDay(t *testing.T) {
	if _, err := utils.ParseDate("02/29/2024"); err != nil {
		t.Errorf("valid leap day rejected: %v", err)
	}
	if _, err := utils.ParseDate("02/29/2023"); err == nil {
		t.Error("invalid leap day accepted")
	}
	if _, err := utils.ParseDate("13/01/2026"); err == nil {
		t.Error("month 13 accepted")
	}
}

func TestChaosParseDateInvalidNeverPanics(t *testing.T) {
	for _, in := range []string{"", "/", "//", "1/2", strings.Repeat("9", 10000), "\x00", "😀/😀/😀"} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ParseDate(%q) panicked: %v", in, r)
				}
			}()
			_, _ = utils.ParseDate(in)
		}()
	}
}

func TestChaosAddMonthsFromMonthEnd(t *testing.T) {
	d := func(y int, m time.Month, day int) time.Time {
		return time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
	}
	cases := []struct {
		in     time.Time
		months int
		want   time.Time
	}{
		{d(2026, time.January, 31), 1, d(2026, time.February, 28)},
		{d(2024, time.January, 31), 1, d(2024, time.February, 29)},
		{d(2026, time.March, 31), -1, d(2026, time.February, 28)},
		{d(2026, time.January, 15), 1, d(2026, time.February, 15)},
		{d(2026, time.December, 31), 1, d(2027, time.January, 31)},
		{d(2026, time.January, 1), -1, d(2025, time.December, 1)},
		{d(2026, time.January, 31), 13, d(2027, time.February, 28)},
	}
	for _, c := range cases {
		if got := utils.AddMonths(c.in, c.months); !got.Equal(c.want) {
			t.Errorf("AddMonths(%v, %d) = %v, want %v", c.in.Format("2006-01-02"), c.months, got.Format("2006-01-02"), c.want.Format("2006-01-02"))
		}
	}
}

func TestChaosFormatDateInputHostile(t *testing.T) {
	for _, in := range []string{strings.Repeat("1", 5000), "😀😀😀", "\x00\x00"} {
		got := utils.FormatDateInput(in)
		if len(got) > 10 {
			t.Errorf("FormatDateInput(%.10q...) = %q, longer than MM/DD/YYYY", in, got)
		}
	}
}
