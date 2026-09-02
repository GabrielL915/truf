package ui

import (
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabriel-luiz/truf/internal/fixture"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/storage"
)

// raceEnabled is flipped on by race_test.go under -race, where budgets are meaningless.
var raceEnabled = false

// Budgets are ~10x the measured local numbers so slow CI runners pass while a
// real regression (an O(n²) loop, a per-frame re-query of the whole DB) still fails.
// Set TRUF_PERF_STRICT=1 to run them at ~3x for local profiling.
const (
	budgetEntries       = 10_000
	budgetFrames        = 200
	budgetOverviewLax   = 25 * time.Millisecond
	budgetTableLax      = 10 * time.Millisecond
	budgetOverviewTight = 7 * time.Millisecond
	budgetTableTight    = 3 * time.Millisecond
)

func budgets(t *testing.T) (overview, table time.Duration) {
	t.Helper()
	if raceEnabled {
		t.Skip("perf budgets are not meaningful under -race")
	}
	if testing.Short() {
		t.Skip("perf budgets skipped in -short mode")
	}
	if os.Getenv("TRUF_PERF_STRICT") == "1" {
		return budgetOverviewTight, budgetTableTight
	}
	return budgetOverviewLax, budgetTableLax
}

func budgetModel(t *testing.T) *Model {
	t.Helper()
	store := storage.NewMemoryStorage()
	store.Snapshot = fixture.Snapshot(budgetEntries, 24)
	l, err := ledger.New(store, func() time.Time { return fixture.Now })
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(l)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return m
}

// medianFrame renders `frames` times and returns the median, which is robust to GC pauses.
func medianFrame(m *Model, frames int) time.Duration {
	samples := make([]time.Duration, frames)
	for i := range samples {
		start := time.Now()
		_ = m.View()
		samples[i] = time.Since(start)
	}
	// insertion sort: n is tiny
	for i := 1; i < len(samples); i++ {
		for j := i; j > 0 && samples[j] < samples[j-1]; j-- {
			samples[j], samples[j-1] = samples[j-1], samples[j]
		}
	}
	return samples[len(samples)/2]
}

func TestPerfBudgetOverviewFrame(t *testing.T) {
	budget, _ := budgets(t)
	m := budgetModel(t)
	got := medianFrame(m, budgetFrames)
	t.Logf("overview frame median: %v (budget %v, %d entries)", got, budget, budgetEntries)
	if got > budget {
		t.Fatalf("overview frame too slow: median %v > budget %v", got, budget)
	}
}

func TestPerfBudgetTableFrame(t *testing.T) {
	_, budget := budgets(t)
	m := budgetModel(t)
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := medianFrame(m, budgetFrames)
	t.Logf("table frame median: %v (budget %v, %d entries)", got, budget, budgetEntries)
	if got > budget {
		t.Fatalf("table frame too slow: median %v > budget %v", got, budget)
	}
}
