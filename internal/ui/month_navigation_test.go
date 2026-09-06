package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
)

func monthModel(t *testing.T) *Model {
	t.Helper()
	m, _ := chaosModel(t)
	feb := time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)
	for _, e := range []ledger.Entry{
		{Kind: ledger.Income, Date: chaosNow, Description: "march salary", Amount: 100},
		{Kind: ledger.Income, Date: feb, Description: "feb salary", Amount: 200},
		{Kind: ledger.Income, Date: feb, Description: "feb bonus", Amount: 300},
		{Kind: ledger.Expense, Date: feb, Description: "feb rent", Amount: 400},
	} {
		if _, err := m.ledger.Add(e); err != nil {
			t.Fatal(err)
		}
	}
	m.refreshTables()
	return m
}

func openIncome(m *Model) {
	key(m, "down")
	key(m, "enter")
}

func TestMonthNavigationPreviousShowsOlderEntries(t *testing.T) {
	m := monthModel(t)
	openIncome(m)
	key(m, "down")
	key(m, "[")

	if got := m.month; !got.Equal(time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("month = %v, want Feb 2026", got)
	}
	if n := len(m.incomeTable.Entries); n != 2 {
		t.Errorf("income rows = %d, want 2", n)
	}
	if n := len(m.expenseTable.Entries); n != 1 {
		t.Errorf("expense rows = %d, want 1", n)
	}
	if m.incomeTable.Cursor != 0 {
		t.Errorf("cursor = %d, want 0 after month change", m.incomeTable.Cursor)
	}
	view := m.View()
	if !strings.Contains(view, "Income — Feb 2026") {
		t.Errorf("view lacks month in title:\n%s", view)
	}
	if !strings.HasSuffix(m.statusBar.TimeRange, "Feb 26") {
		t.Errorf("chart range = %q, want it to end in Feb 26", m.statusBar.TimeRange)
	}
}

func TestMonthNavigationNextAndNewEntryDate(t *testing.T) {
	m := monthModel(t)
	openIncome(m)
	key(m, "]")

	if n := len(m.incomeTable.Entries); n != 0 {
		t.Fatalf("April should be empty, got %d rows", n)
	}
	key(m, "n")
	current, ok := m.incomeTable.Current()
	if !ok {
		t.Fatal("expected a new row")
	}
	want := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	if !current.Date.Equal(want) {
		t.Errorf("new entry date = %v, want %v", current.Date, want)
	}
	key(m, "esc")

	key(m, "h")
	key(m, "h")
	if !strings.Contains(m.View(), "Income — Feb 2026") {
		t.Errorf("h should move back two months to Feb 2026")
	}
	key(m, "l")
	if !strings.Contains(m.View(), "Income — Mar 2026") {
		t.Errorf("l should move forward to Mar 2026")
	}
}

func TestMonthNavigationIgnoredOutsideTables(t *testing.T) {
	m := monthModel(t)
	start := m.month

	key(m, "[")
	if !m.month.Equal(start) {
		t.Errorf("[ on menu changed month to %v", m.month)
	}

	key(m, "enter")
	key(m, "]")
	if !m.month.Equal(start) {
		t.Errorf("] on overview chart changed month to %v", m.month)
	}

	key(m, "tab")
	openIncome(m)
	key(m, "enter")
	key(m, "tab")
	typeText(m, "[")
	if !m.month.Equal(start) {
		t.Errorf("[ while editing changed month to %v", m.month)
	}
	if !strings.Contains(m.incomeTable.EditBuffer, "[") {
		t.Errorf("[ while editing should be typed, buffer=%q", m.incomeTable.EditBuffer)
	}
}
