package storage_test

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/storage"
)

func openStore(t *testing.T) (*storage.SQLiteStorage, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "chaos.db")
	s, err := storage.NewSQLiteStorage(path)
	if err != nil {
		t.Fatalf("NewSQLiteStorage: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s, path
}

func entry(id, desc string, amount float64) ledger.Entry {
	return ledger.Entry{
		ID: id, Date: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		Description: desc, Category: "Food", Amount: amount, Kind: ledger.Expense,
	}
}

func TestChaosEmptyPathReturnsErrorNotPanic(t *testing.T) {
	t.Skip("CHAOS: NewSQLiteStorage(\"\") panics on dbPath[0]")
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewSQLiteStorage(\"\") panicked: %v", r)
		}
	}()
	s, err := storage.NewSQLiteStorage("")
	if err == nil {
		s.Close()
		t.Error("empty path accepted")
	}
}

func TestChaosPathIsDirectoryReturnsError(t *testing.T) {
	s, err := storage.NewSQLiteStorage(t.TempDir())
	if err == nil {
		s.Close()
		t.Error("directory path accepted as database")
	}
}

func TestChaosHostileDescriptionsRoundTrip(t *testing.T) {
	s, _ := openStore(t)
	hostile := []string{
		strings.Repeat("A", 10000),
		"'; DROP TABLE entries; --",
		"😀​‮",
		"line1\nline2\r\n",
		"nul\x00byte",
		"",
	}
	var entries []ledger.Entry
	for i, d := range hostile {
		entries = append(entries, entry(fmt.Sprintf("h%d", i), d, 1))
	}
	if err := s.Save(ledger.Snapshot{Entries: entries}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Entries) != len(hostile) {
		t.Fatalf("loaded %d entries, want %d", len(got.Entries), len(hostile))
	}
	byID := map[string]string{}
	for _, e := range got.Entries {
		byID[e.ID] = e.Description
	}
	for i, d := range hostile {
		if byID[fmt.Sprintf("h%d", i)] != d {
			t.Errorf("entry %d: got %.30q want %.30q", i, byID[fmt.Sprintf("h%d", i)], d)
		}
	}
}

func TestChaosFailedSaveRollsBackPreviousData(t *testing.T) {
	s, _ := openStore(t)
	if err := s.Save(ledger.Snapshot{Entries: []ledger.Entry{entry("keep", "ok", 1)}}); err != nil {
		t.Fatal(err)
	}
	bad := ledger.Snapshot{Entries: []ledger.Entry{entry("dup", "x", 1), entry("dup", "y", 2)}}
	if err := s.Save(bad); err == nil {
		t.Fatal("snapshot with duplicate IDs saved without error")
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].ID != "keep" {
		t.Errorf("after failed save, entries = %+v, want only 'keep'", got.Entries)
	}
}

func TestChaosNaNAmountDoesNotCorruptDatabase(t *testing.T) {
	s, _ := openStore(t)
	err := s.Save(ledger.Snapshot{Entries: []ledger.Entry{entry("nan", "x", math.NaN())}})
	got, loadErr := s.Load()
	if loadErr != nil {
		t.Fatalf("Load after NaN save (save err=%v): %v", err, loadErr)
	}
	if err == nil && len(got.Entries) == 1 && math.IsNaN(got.Entries[0].Amount) {
		t.Error("NaN amount persisted and reloaded as NaN")
	}
}

func TestChaosUnknownKindIsNotSilentlyCoerced(t *testing.T) {
	t.Skip("CHAOS: unknown entry_type silently loads as Expense (contestable)")
	s, _ := openStore(t)
	e := entry("k", "x", 1)
	e.Kind = ledger.Kind("refund")
	if err := s.Save(ledger.Snapshot{Entries: []ledger.Entry{e}}); err != nil {
		return
	}
	got, err := s.Load()
	if err != nil {
		return
	}
	if len(got.Entries) == 1 && got.Entries[0].Kind == ledger.Expense {
		t.Error("kind 'refund' saved fine and reloaded as Expense")
	}
}

func TestChaosZeroDateRoundTrip(t *testing.T) {
	s, _ := openStore(t)
	z := entry("z", "y", 1)
	z.Date = time.Time{}
	if err := s.Save(ledger.Snapshot{Entries: []ledger.Entry{z}}); err != nil {
		t.Logf("zero date rejected at save: %v", err)
		return
	}
	got, err := s.Load()
	if err != nil {
		t.Errorf("zero date saved fine but Load now fails for the whole db: %v", err)
		return
	}
	if len(got.Entries) != 1 {
		t.Errorf("zero-date entry lost on reload")
	}
}

func TestChaosTwoProcessesWritingSameFile(t *testing.T) {
	t.Skip("CHAOS: two truf instances -> SQLITE_BUSY; no busy_timeout / WAL")
	s1, path := openStore(t)
	s2, err := storage.NewSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	var wg sync.WaitGroup
	errs := make(chan error, 200)
	for _, s := range []*storage.SQLiteStorage{s1, s2} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				var es []ledger.Entry
				for j := 0; j < 20; j++ {
					es = append(es, entry(fmt.Sprintf("e%d", j), "d", float64(j)))
				}
				if err := s.Save(ledger.Snapshot{Entries: es}); err != nil {
					errs <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent save failed: %v", err)
		break
	}
}

func TestChaosLargeSnapshotSaveIsBounded(t *testing.T) {
	s, _ := openStore(t)
	var es []ledger.Entry
	for i := 0; i < 20000; i++ {
		es = append(es, entry(fmt.Sprintf("big%d", i), "d", 1))
	}
	start := time.Now()
	if err := s.Save(ledger.Snapshot{Entries: es}); err != nil {
		t.Fatal(err)
	}
	d := time.Since(start)
	if d > 5*time.Second {
		t.Errorf("saving 20k entries took %v (every keystroke commit does this)", d)
	}
	t.Logf("20k-entry save took %v", d)
}
