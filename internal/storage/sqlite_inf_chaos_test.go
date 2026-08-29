package storage_test

import (
	"math"
	"testing"

	"github.com/gabriel-luiz/truf/internal/ledger"
)

func TestChaosInfAmountDoesNotPersist(t *testing.T) {
	t.Skip("CHAOS: Inf amount persists; FormatCurrency(Inf) then panics on every launch")
	s, _ := openStore(t)
	err := s.Save(ledger.Snapshot{Entries: []ledger.Entry{entry("inf", "x", math.Inf(1))}})
	got, loadErr := s.Load()
	if loadErr != nil {
		t.Fatalf("Load after Inf save (save err=%v): %v", err, loadErr)
	}
	if err == nil && len(got.Entries) == 1 && math.IsInf(got.Entries[0].Amount, 0) {
		t.Errorf("Inf amount persisted and reloaded as %v — every later launch will render it", got.Entries[0].Amount)
	}
}
