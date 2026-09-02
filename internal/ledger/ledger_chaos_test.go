package ledger_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
)

func TestChaosSummaryBalanceIsExactForCents(t *testing.T) {
	l, _ := newLedger(t)
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 10})
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 20})
	mustAdd(t, l, ledger.Entry{Kind: ledger.Expense, Amount: 30})

	if b := l.Summary(fixedNow).Balance; b != 0 {
		t.Errorf("10 + 20 - 30 balance = %v, want exactly 0", b)
	}
}

func TestChaosTotalBalanceAfterManyCents(t *testing.T) {
	l, _ := newLedger(t)
	for range 1000 {
		mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 1})
	}
	mustAdd(t, l, ledger.Entry{Kind: ledger.Expense, Amount: 1000})
	if b := l.TotalBalance(); b != 0 {
		t.Errorf("1000 × 1 - 1000 = %v, want exactly 0", b)
	}
}

func TestChaosAddRejectsNegativeAmount(t *testing.T) {
	l, _ := newLedger(t)
	if _, err := l.Add(ledger.Entry{Kind: ledger.Expense, Amount: -50}); err == nil {
		t.Errorf("negative expense accepted; balance %v", l.TotalBalance())
	}
	if l.TotalBalance() != 0 {
		t.Errorf("rejected entry still changed balance: %v", l.TotalBalance())
	}
}

func TestChaosUpdateRejectsNegativeAmount(t *testing.T) {
	l, _ := newLedger(t)
	e := mustAdd(t, l, ledger.Entry{Kind: ledger.Expense, Amount: 50})
	e.Amount = -50
	if err := l.Update(e); err == nil {
		t.Error("Update to negative amount accepted")
	}
	if l.TotalBalance() != -50 {
		t.Errorf("balance after rejected update = %v, want -50", l.TotalBalance())
	}
}

func TestChaosAddRejectsUnknownKind(t *testing.T) {
	l, _ := newLedger(t)
	if _, err := l.Add(ledger.Entry{Kind: ledger.Kind("refund"), Amount: 10}); err == nil {
		s := l.Summary(fixedNow)
		t.Errorf("unknown kind accepted; counted as expenses=%v income=%v", s.TotalExpenses, s.TotalIncome)
	}
}

func TestChaosAddWithDuplicateIDIsRejected(t *testing.T) {
	l, _ := newLedger(t)
	first := mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 10})
	_, err := l.Add(ledger.Entry{ID: first.ID, Kind: ledger.Income, Amount: 20})
	if err == nil {
		t.Fatalf("duplicate ID accepted; total=%v", l.TotalBalance())
	}
	if l.TotalBalance() != 10 {
		t.Errorf("balance after rejected duplicate = %v, want 10", l.TotalBalance())
	}
}

func TestChaosChartSeriesHostileN(t *testing.T) {
	l, _ := newLedger(t)
	for _, n := range []int{0, -1, math.MinInt32} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ChartSeries(n=%d) panicked: %v", n, r)
				}
			}()
			d := l.ChartSeries(fixedNow, n)
			if len(d.Months) != 0 {
				t.Errorf("n=%d produced %d months", n, len(d.Months))
			}
		}()
	}
	d := l.ChartSeries(fixedNow, 1200)
	if len(d.Months) != 1200 {
		t.Errorf("n=1200 produced %d months", len(d.Months))
	}
}

func TestChaosChartSeriesFromMonthEndAnchor(t *testing.T) {
	l, _ := newLedger(t)
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 100, Date: time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)})
	mar31 := time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)
	d := l.ChartSeries(mar31, 2)
	if d.Months[0] != "Feb 26" || d.Months[1] != "Mar 26" {
		t.Errorf("months = %v, want [Feb 26 Mar 26]", d.Months)
	}
	if d.Income[0] != 100 {
		t.Errorf("Feb income = %v", d.Income[0])
	}
}

func TestChaosTimezoneMonthBoundary(t *testing.T) {
	l, _ := newLedger(t)
	loc := time.FixedZone("UTC+14", 14*3600)
	local := time.Date(2026, time.March, 1, 0, 30, 0, 0, loc)
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 5, Date: local})

	if n := len(l.Entries(month(2026, time.March), ledger.Income)); n != 1 {
		t.Errorf("entry dated Mar 1 (local) listed %d times in March", n)
	}
	if n := len(l.Entries(month(2026, time.February), ledger.Income)); n != 0 {
		t.Errorf("entry dated Mar 1 (local) leaked into February %d times", n)
	}
}

func TestChaosHostileDescriptionSurvivesLedger(t *testing.T) {
	l, store := newLedger(t)
	hostile := []string{
		strings.Repeat("A", 10000),
		"'; DROP TABLE entries; --",
		"😀​‮",
		"line1\nline2\r\n",
		"nul\x00byte",
	}
	for _, d := range hostile {
		e := mustAdd(t, l, ledger.Entry{Kind: ledger.Expense, Amount: 1, Description: d})
		got := l.Entries(fixedNow, ledger.Expense)
		if got[len(got)-1].Description != d {
			t.Errorf("description %.20q altered in memory", d)
		}
		saved, _ := store.LastSave()
		if saved.Entries[len(saved.Entries)-1].ID != e.ID {
			t.Errorf("last save does not contain %.20q", d)
		}
	}
}

func TestChaosEntriesReturnsCopyNotAlias(t *testing.T) {
	l, _ := newLedger(t)
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 10})
	got := l.Entries(fixedNow, ledger.Income)
	got[0].Amount = 999
	if l.TotalBalance() != 10 {
		t.Errorf("mutating Entries() result changed ledger: %v", l.TotalBalance())
	}
}
