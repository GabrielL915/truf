package ledger_test

import (
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/internal/fixture"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/storage"
)

const benchEntries = 10_000

func benchLedger(b *testing.B) *ledger.Ledger {
	b.Helper()
	store := storage.NewMemoryStorage()
	store.Snapshot = fixture.Snapshot(benchEntries, 24)
	l, err := ledger.New(store, func() time.Time { return fixture.Now })
	if err != nil {
		b.Fatal(err)
	}
	return l
}

func BenchmarkEntriesMonth(b *testing.B) {
	l := benchLedger(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = l.Entries(fixture.Now, ledger.Expense)
	}
}

func BenchmarkSummary(b *testing.B) {
	l := benchLedger(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = l.Summary(fixture.Now)
	}
}

func BenchmarkChartSeries24(b *testing.B) {
	l := benchLedger(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = l.ChartSeries(fixture.Now, 24)
	}
}

func BenchmarkTotalBalance(b *testing.B) {
	l := benchLedger(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = l.TotalBalance()
	}
}

func BenchmarkAddPersistMemory(b *testing.B) {
	l := benchLedger(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := l.Add(ledger.Entry{Description: "bench", Kind: ledger.Expense, Amount: 1250}); err != nil {
			b.Fatal(err)
		}
	}
}
