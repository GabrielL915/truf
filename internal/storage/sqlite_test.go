package storage_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/storage"
)

func TestSQLiteRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "truf.db")

	store, err := storage.NewSQLiteStorage(path)
	if err != nil {
		t.Fatalf("NewSQLiteStorage: %v", err)
	}
	defer store.Close()

	snapshot := ledger.Snapshot{
		Entries: []ledger.Entry{
			{
				ID:          "a",
				Date:        time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC),
				Description: "Salary",
				Category:    "Salary",
				Amount:      4500,
				Kind:        ledger.Income,
			},
			{
				ID:          "b",
				Date:        time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC),
				Description: "Rent",
				Category:    "Housing",
				Amount:      1200,
				Kind:        ledger.Expense,
			},
		},
		Categories: []ledger.Category{
			{Name: "Salary", Kind: ledger.Income, Order: 1},
			{Name: "Housing", Kind: ledger.Expense, Order: 1},
		},
	}

	if err := store.Save(snapshot); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Entries) != len(snapshot.Entries) {
		t.Fatalf("expected %d entries, got %d", len(snapshot.Entries), len(loaded.Entries))
	}
	for i, want := range snapshot.Entries {
		got := loaded.Entries[i]
		if got.ID != want.ID || got.Description != want.Description || got.Category != want.Category ||
			got.Amount != want.Amount || got.Kind != want.Kind || !got.Date.Equal(want.Date) {
			t.Fatalf("entry %d: expected %+v, got %+v", i, want, got)
		}
	}

	if len(loaded.Categories) != len(snapshot.Categories) {
		t.Fatalf("expected %d categories, got %+v", len(snapshot.Categories), loaded.Categories)
	}
}

func TestSQLiteLoadOnFreshDBHasDefaultCategories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "truf.db")

	store, err := storage.NewSQLiteStorage(path)
	if err != nil {
		t.Fatalf("NewSQLiteStorage: %v", err)
	}
	defer store.Close()

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Entries) != 0 {
		t.Fatalf("expected no entries, got %d", len(loaded.Entries))
	}
	if len(loaded.Categories) != len(ledger.DefaultCategories()) {
		t.Fatalf("expected default categories, got %d", len(loaded.Categories))
	}
}
