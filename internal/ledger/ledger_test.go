package ledger_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/storage"
)

var fixedNow = time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedNow }

func month(year int, m time.Month) time.Time {
	return time.Date(year, m, 1, 0, 0, 0, 0, time.UTC)
}

func newLedger(t *testing.T) (*ledger.Ledger, *storage.MemoryStorage) {
	t.Helper()
	store := storage.NewMemoryStorage()
	l, err := ledger.New(store, fixedClock)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return l, store
}

func mustAdd(t *testing.T, l *ledger.Ledger, e ledger.Entry) ledger.Entry {
	t.Helper()
	created, err := l.Add(e)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	return created
}

func TestAddStampsIDAndDateAndPersists(t *testing.T) {
	l, store := newLedger(t)

	created := mustAdd(t, l, ledger.Entry{Description: "Salary", Amount: 100, Kind: ledger.Income})

	if created.ID == "" {
		t.Fatal("expected Add to stamp an ID")
	}
	if !created.Date.Equal(fixedNow) {
		t.Fatalf("expected date %v, got %v", fixedNow, created.Date)
	}

	saved, ok := store.LastSave()
	if !ok {
		t.Fatal("expected Add to persist")
	}
	if len(saved.Entries) != 1 || saved.Entries[0].ID != created.ID {
		t.Fatalf("unexpected saved entries: %+v", saved.Entries)
	}

	if got := len(l.Entries(month(2026, time.March), ledger.Income)); got != 1 {
		t.Fatalf("expected 1 entry in March, got %d", got)
	}
}

func TestAddDefaultsCategory(t *testing.T) {
	l, _ := newLedger(t)

	income := mustAdd(t, l, ledger.Entry{Description: "Salary", Amount: 100, Kind: ledger.Income})
	if want := l.Categories(ledger.Income)[0].Name; income.Category != want {
		t.Fatalf("expected category %q, got %q", want, income.Category)
	}

	expense := mustAdd(t, l, ledger.Entry{Description: "Rent", Amount: 50, Kind: ledger.Expense})
	if want := l.Categories(ledger.Expense)[0].Name; expense.Category != want {
		t.Fatalf("expected category %q, got %q", want, expense.Category)
	}
}

func TestAddKeepsFreeTextCategory(t *testing.T) {
	l, _ := newLedger(t)

	created := mustAdd(t, l, ledger.Entry{Description: "Gift", Category: "Anything goes", Amount: 10, Kind: ledger.Income})
	if created.Category != "Anything goes" {
		t.Fatalf("expected free-text category to be kept, got %q", created.Category)
	}
}

func TestUpdateMovesEntryBetweenMonths(t *testing.T) {
	l, store := newLedger(t)

	created := mustAdd(t, l, ledger.Entry{Description: "Rent", Amount: 1200, Kind: ledger.Expense})

	created.Date = time.Date(2026, time.April, 3, 0, 0, 0, 0, time.UTC)
	if err := l.Update(created); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got := len(l.Entries(month(2026, time.March), ledger.Expense)); got != 0 {
		t.Fatalf("expected March to be empty, got %d entries", got)
	}
	april := l.Entries(month(2026, time.April), ledger.Expense)
	if len(april) != 1 || april[0].ID != created.ID {
		t.Fatalf("expected entry in April, got %+v", april)
	}

	saved, _ := store.LastSave()
	if len(saved.Entries) != 1 || !saved.Entries[0].Date.Equal(created.Date) {
		t.Fatalf("expected update to persist new date, got %+v", saved.Entries)
	}
}

func TestUpdateUnknownIDErrors(t *testing.T) {
	l, _ := newLedger(t)

	if err := l.Update(ledger.Entry{ID: "missing", Kind: ledger.Income}); err == nil {
		t.Fatal("expected error updating unknown entry")
	}
}

func TestRemovePersistsAndUnknownIDErrors(t *testing.T) {
	l, store := newLedger(t)

	created := mustAdd(t, l, ledger.Entry{Description: "Coffee", Amount: 5, Kind: ledger.Expense})

	if err := l.Remove(created.ID); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if got := len(l.Entries(month(2026, time.March), ledger.Expense)); got != 0 {
		t.Fatalf("expected entry removed, got %d", got)
	}

	saved, _ := store.LastSave()
	if len(saved.Entries) != 0 {
		t.Fatalf("expected removal to persist, got %+v", saved.Entries)
	}

	if err := l.Remove("missing"); err == nil {
		t.Fatal("expected error removing unknown entry")
	}
}

func TestSummaryTotalsPerMonth(t *testing.T) {
	l, _ := newLedger(t)

	march := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)
	april := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)

	mustAdd(t, l, ledger.Entry{Date: march, Category: "Salary", Amount: 1000, Kind: ledger.Income})
	mustAdd(t, l, ledger.Entry{Date: march, Category: "Food", Amount: 300, Kind: ledger.Expense})
	mustAdd(t, l, ledger.Entry{Date: march, Category: "Food", Amount: 200, Kind: ledger.Expense})
	mustAdd(t, l, ledger.Entry{Date: april, Category: "Salary", Amount: 900, Kind: ledger.Income})

	s := l.Summary(month(2026, time.March))
	if s.TotalIncome != 1000 || s.TotalExpenses != 500 || s.Balance != 500 {
		t.Fatalf("unexpected March summary: %+v", s)
	}
	if s.CategoryTotals["Food"] != 500 {
		t.Fatalf("expected Food total 500, got %v", s.CategoryTotals["Food"])
	}

	if got := l.Summary(month(2026, time.April)).Balance; got != 900 {
		t.Fatalf("expected April balance 900, got %v", got)
	}
}

func TestTotalBalanceAcrossMonths(t *testing.T) {
	l, _ := newLedger(t)

	mustAdd(t, l, ledger.Entry{Date: time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC), Amount: 1000, Kind: ledger.Income})
	mustAdd(t, l, ledger.Entry{Date: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC), Amount: 400, Kind: ledger.Expense})

	if got := l.TotalBalance(); got != 600 {
		t.Fatalf("expected 600, got %v", got)
	}
}

func TestChartSeriesRunningBalance(t *testing.T) {
	l, _ := newLedger(t)

	mustAdd(t, l, ledger.Entry{Date: time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC), Amount: 1000, Kind: ledger.Income})
	mustAdd(t, l, ledger.Entry{Date: time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC), Amount: 400, Kind: ledger.Expense})
	mustAdd(t, l, ledger.Entry{Date: time.Date(2026, time.March, 5, 0, 0, 0, 0, time.UTC), Amount: 200, Kind: ledger.Income})

	data := l.ChartSeries(month(2026, time.March), 3)

	if len(data.Months) != 3 {
		t.Fatalf("expected 3 months, got %d", len(data.Months))
	}
	for i, want := range []string{"Jan 26", "Feb 26", "Mar 26"} {
		if data.Months[i] != want {
			t.Fatalf("month %d: expected %q, got %q", i, want, data.Months[i])
		}
	}
	for i, want := range []int64{1000, 600, 800} {
		if data.Balance[i] != want {
			t.Fatalf("balance %d: expected %v, got %v", i, want, data.Balance[i])
		}
	}
	if data.Income[0] != 1000 || data.Expenses[1] != 400 || data.Income[2] != 200 {
		t.Fatalf("unexpected series: %+v", data)
	}
}

func TestChartSeriesEmptyLedgerEndsAtClockMonth(t *testing.T) {
	l, _ := newLedger(t)

	data := l.ChartSeries(l.Now(), 4)

	if len(data.Months) != 4 {
		t.Fatalf("expected 4 months, got %d", len(data.Months))
	}
	if data.Months[3] != "Mar 26" {
		t.Fatalf("expected series to end at Mar 26, got %q", data.Months[3])
	}
	for i := range data.Months {
		if data.Income[i] != 0 || data.Expenses[i] != 0 || data.Balance[i] != 0 {
			t.Fatalf("expected zero-valued month at %d, got %+v", i, data)
		}
	}
}

func TestSaveFailureIsReturned(t *testing.T) {
	store := storage.NewMemoryStorage()
	store.SaveErr = errors.New("disk full")

	l, err := ledger.New(store, fixedClock)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	created, err := l.Add(ledger.Entry{Description: "Salary", Amount: 100, Kind: ledger.Income})
	if err == nil {
		t.Fatal("expected Add to return the storage error")
	}
	if created.ID == "" {
		t.Fatal("expected the entry to still be stamped")
	}
	if got := len(l.Entries(month(2026, time.March), ledger.Income)); got != 1 {
		t.Fatalf("expected in-memory state to keep the entry, got %d", got)
	}
}

func TestNewReturnsLoadError(t *testing.T) {
	store := storage.NewMemoryStorage()
	store.LoadErr = errors.New("corrupt row")

	if _, err := ledger.New(store, fixedClock); err == nil {
		t.Fatal("expected New to return the load error")
	}
}
