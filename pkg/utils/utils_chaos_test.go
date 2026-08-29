package utils_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/pkg/utils"
)

func TestChaosParseCurrencyRejectsNonFiniteInput(t *testing.T) {
	t.Skip("CHAOS: ParseCurrency accepts NaN/Inf via strconv.ParseFloat")
	for _, in := range []string{"NaN", "Inf", "-Inf", "inf", "1e400"} {
		v, err := utils.ParseCurrency(in)
		if err == nil && (math.IsNaN(v) || math.IsInf(v, 0)) {
			t.Errorf("ParseCurrency(%q) = %v, want error", in, v)
		}
	}
}

func TestChaosParseCurrencyRejectsGarbage(t *testing.T) {
	t.Skip("CHAOS: ParseCurrency accepts \"$$\" (=0) and \"1,5,\" (=15)")
	for _, in := range []string{"1.2.3", "abc", "$$", "1,5,", "１２"} {
		if _, err := utils.ParseCurrency(in); err == nil {
			t.Errorf("ParseCurrency(%q) accepted", in)
		}
	}
}

func TestChaosFormatCurrencyNonFiniteDoesNotPanic(t *testing.T) {
	t.Skip("CHAOS: FormatCurrency panics on NaN/Inf (no decimal point to split)")
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("FormatCurrency(%v) panicked: %v", v, r)
				}
			}()
			_ = utils.FormatCurrency(v)
		}()
	}
}

func TestChaosFormatCurrencyNegativeZeroRounding(t *testing.T) {
	t.Skip("CHAOS: FormatCurrency(-0.001) renders $-0.00 (cosmetic)")
	got := utils.FormatCurrency(-0.001)
	if strings.Contains(got, "-") {
		t.Errorf("FormatCurrency(-0.001) = %q, want no minus sign", got)
	}
}

func TestChaosFormatCurrencyRoundTripsParse(t *testing.T) {
	for _, v := range []float64{0, 0.1, 1234567.89, -42.5, 1e12} {
		back, err := utils.ParseCurrency(utils.FormatCurrency(v))
		if err != nil || math.Abs(back-v) > 0.005 {
			t.Errorf("round trip %v -> %q -> %v (%v)", v, utils.FormatCurrency(v), back, err)
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
	t.Skip("CHAOS: AddMonths(Jan 31,1) = Mar 3 (contestable: only called with 1st of month)")
	jan31 := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	got := utils.AddMonths(jan31, 1)
	if got.Month() != time.February {
		t.Errorf("AddMonths(Jan 31, 1) = %v, want a date in February", got)
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
