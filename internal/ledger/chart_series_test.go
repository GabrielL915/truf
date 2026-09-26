package ledger_test

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/internal/fixture"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/storage"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

const (
	chartSeriesBudgetEntries = 10_000
	chartSeriesBudgetRuns    = 50
	chartSeriesBudgetLax     = 3 * time.Millisecond
	chartSeriesBudgetTight   = 1 * time.Millisecond
)

func seriesFromSummaries(l *ledger.Ledger, endMonth time.Time, n int) ledger.ChartData {
	if n <= 0 {
		return ledger.ChartData{}
	}
	data := ledger.ChartData{
		Months:   make([]string, 0, n),
		Income:   make([]int64, 0, n),
		Expenses: make([]int64, 0, n),
		Balance:  make([]int64, 0, n),
	}
	var running int64
	for i := n - 1; i >= 0; i-- {
		month := utils.AddMonths(utils.FirstOfMonth(endMonth), -i)
		s := l.Summary(month)
		running += s.Balance
		data.Months = append(data.Months, utils.FormatMonthYear(month))
		data.Income = append(data.Income, s.TotalIncome)
		data.Expenses = append(data.Expenses, s.TotalExpenses)
		data.Balance = append(data.Balance, running)
	}
	return data
}

func fixtureLedger(t testing.TB, entries, months int) *ledger.Ledger {
	t.Helper()
	store := storage.NewMemoryStorage()
	store.Snapshot = fixture.Snapshot(entries, months)
	l, err := ledger.New(store, func() time.Time { return fixture.Now })
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestChartSeriesMatchesPerMonthSummaries(t *testing.T) {
	l := fixtureLedger(t, 3_000, 30)
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 7, Date: time.Date(2026, time.March, 1, 0, 30, 0, 0, time.FixedZone("UTC+14", 14*3600))})
	mustAdd(t, l, ledger.Entry{Kind: ledger.Expense, Amount: 3, Date: time.Date(2025, time.December, 31, 23, 0, 0, 0, time.FixedZone("UTC-10", -10*3600))})

	ends := []time.Time{fixture.Now, time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC), time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2030, time.July, 4, 0, 0, 0, 0, time.UTC)}
	for _, end := range ends {
		for _, n := range []int{1, 3, 6, 12, 24, 40} {
			got := l.ChartSeries(end, n)
			want := seriesFromSummaries(l, end, n)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("ChartSeries(%s, %d)\n got  %+v\n want %+v", end.Format("2006-01"), n, got, want)
			}
		}
	}
}

func TestChartSeriesIgnoresEntriesOutsideWindow(t *testing.T) {
	l, _ := newLedger(t)
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 1000, Date: time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)})
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 50, Date: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)})
	mustAdd(t, l, ledger.Entry{Kind: ledger.Expense, Amount: 20, Date: time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)})
	mustAdd(t, l, ledger.Entry{Kind: ledger.Expense, Amount: 900, Date: time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)})

	data := l.ChartSeries(month(2026, time.March), 3)

	if !reflect.DeepEqual(data.Income, []int64{50, 0, 0}) || !reflect.DeepEqual(data.Expenses, []int64{0, 0, 20}) {
		t.Fatalf("entries outside the window leaked in: %+v", data)
	}
	if !reflect.DeepEqual(data.Balance, []int64{50, 50, 30}) {
		t.Fatalf("running balance must start from zero at the first month of the window: %+v", data.Balance)
	}
}

func TestChartSeriesFollowsUpdateAcrossMonths(t *testing.T) {
	l, _ := newLedger(t)
	e := mustAdd(t, l, ledger.Entry{Kind: ledger.Expense, Amount: 300, Date: time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)})

	e.Date = time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)
	if err := l.Update(e); err != nil {
		t.Fatal(err)
	}
	data := l.ChartSeries(month(2026, time.March), 3)

	if !reflect.DeepEqual(data.Expenses, []int64{0, 0, 300}) {
		t.Fatalf("expected the expense to move from Jan to Mar, got %+v", data.Expenses)
	}
	if !reflect.DeepEqual(data.Balance, []int64{0, 0, -300}) {
		t.Fatalf("unexpected running balance %+v", data.Balance)
	}
}

func TestChartSeriesWindowLargerThanHistory(t *testing.T) {
	l, _ := newLedger(t)
	mustAdd(t, l, ledger.Entry{Kind: ledger.Income, Amount: 100, Date: time.Date(2026, time.February, 3, 0, 0, 0, 0, time.UTC)})

	data := l.ChartSeries(month(2026, time.March), 24)

	if len(data.Months) != 24 || data.Months[0] != "Apr 24" || data.Months[23] != "Mar 26" {
		t.Fatalf("unexpected months %v", data.Months)
	}
	for i := range 22 {
		if data.Income[i] != 0 || data.Expenses[i] != 0 || data.Balance[i] != 0 {
			t.Fatalf("month %d before the history should be zero: %+v", i, data)
		}
	}
	if data.Income[22] != 100 || data.Balance[22] != 100 || data.Balance[23] != 100 {
		t.Fatalf("unexpected tail %+v", data)
	}
}

func TestPerfBudgetChartSeries24(t *testing.T) {
	if raceEnabled {
		t.Skip("perf budgets are not meaningful under -race")
	}
	if testing.Short() {
		t.Skip("perf budgets skipped in -short mode")
	}
	budget := chartSeriesBudgetLax
	if os.Getenv("TRUF_PERF_STRICT") == "1" {
		budget = chartSeriesBudgetTight
	}
	l := fixtureLedger(t, chartSeriesBudgetEntries, 24)

	samples := make([]time.Duration, chartSeriesBudgetRuns)
	for i := range samples {
		start := time.Now()
		_ = l.ChartSeries(fixture.Now, 24)
		samples[i] = time.Since(start)
	}
	for i := 1; i < len(samples); i++ {
		for j := i; j > 0 && samples[j] < samples[j-1]; j-- {
			samples[j], samples[j-1] = samples[j-1], samples[j]
		}
	}
	got := samples[len(samples)/2]
	t.Logf("ChartSeries(24) median: %v (budget %v, %d entries)", got, budget, chartSeriesBudgetEntries)
	if got > budget {
		t.Fatalf("ChartSeries(24) too slow: median %v > budget %v", got, budget)
	}
}
