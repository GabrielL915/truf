// Package fixture builds deterministic synthetic ledger data for benchmarks and tests.
package fixture

import (
	"fmt"
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
)

// Now is the fixed clock every fixture is anchored on.
var Now = time.Date(2026, time.June, 15, 12, 0, 0, 0, time.UTC)

// Entries returns n entries spread evenly over the `months` months ending at Now,
// alternating income/expense and cycling through a few categories.
func Entries(n, months int) []ledger.Entry {
	if months < 1 {
		months = 1
	}
	out := make([]ledger.Entry, 0, n)
	for i := 0; i < n; i++ {
		monthsBack := i % months
		day := 1 + (i % 28)
		date := time.Date(Now.Year(), Now.Month()-time.Month(monthsBack), day, 9, 0, 0, 0, time.UTC)
		e := ledger.Entry{
			ID:          fmt.Sprintf("fx-%06d", i),
			Date:        date,
			Description: fmt.Sprintf("Lançamento sintético nº %d", i),
			Kind:        ledger.Expense,
			Category:    "Food",
		}
		if i%3 == 0 {
			e.Kind = ledger.Income
			e.Category = "Salary"
		}
		setAmount(&e.Amount, 1000+(i%500)*10)
		out = append(out, e)
	}
	return out
}

// Snapshot wraps Entries in a Snapshot with the default categories.
func Snapshot(n, months int) ledger.Snapshot {
	return ledger.Snapshot{Entries: Entries(n, months), Categories: ledger.DefaultCategories()}
}

// setAmount assigns through a pointer so it compiles whether Amount is float64 or int64 cents.
func setAmount[T ~int64 | ~float64](dst *T, v int) { *dst = T(v) }
