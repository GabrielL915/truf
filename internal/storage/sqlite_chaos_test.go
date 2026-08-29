package storage_test

import (
	"database/sql"
	"fmt"
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

func entry(id, desc string, amount int64) ledger.Entry {
	return ledger.Entry{
		ID: id, Date: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		Description: desc, Category: "Food", Amount: amount, Kind: ledger.Expense,
	}
}

func TestChaosEmptyPathReturnsErrorNotPanic(t *testing.T) {
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

func TestChaosLargeCentsRoundTripExactly(t *testing.T) {
	s, _ := openStore(t)
	const big = int64(1) << 62
	if err := s.Save(ledger.Snapshot{Entries: []ledger.Entry{entry("big", "x", big), entry("one", "y", 1)}}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range got.Entries {
		switch e.ID {
		case "big":
			if e.Amount != big {
				t.Errorf("big amount = %d, want %d", e.Amount, big)
			}
		case "one":
			if e.Amount != 1 {
				t.Errorf("one cent = %d", e.Amount)
			}
		}
	}
}

func TestChaosUnknownKindIsNotSilentlyCoerced(t *testing.T) {
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
			for range 100 {
				var es []ledger.Entry
				for j := range 20 {
					es = append(es, entry(fmt.Sprintf("e%d", j), "d", int64(j)))
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

func TestChaosLegacyFloatDatabaseMigratesToCents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec(`
		CREATE TABLE entries (
			id TEXT PRIMARY KEY, date TEXT NOT NULL, description TEXT NOT NULL,
			category TEXT NOT NULL, amount REAL NOT NULL, entry_type TEXT NOT NULL
		);
		INSERT INTO entries VALUES ('a', '2026-03-01', 'coffee', 'Food', 4.5, 'expense');
		INSERT INTO entries VALUES ('b', '2026-03-02', 'rent', 'Housing', 1234.56, 'expense');
		INSERT INTO entries VALUES ('c', '2026-03-03', 'drift', 'Food', 0.1+0.2, 'expense');
		INSERT INTO entries VALUES ('d', '2026-03-04', 'salary', 'Salary', 4200, 'income');
	`)
	raw.Close()
	if err != nil {
		t.Fatal(err)
	}

	s, err := storage.NewSQLiteStorage(path)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	defer s.Close()

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load legacy: %v", err)
	}
	want := map[string]int64{"a": 450, "b": 123456, "c": 30, "d": 420000}
	if len(got.Entries) != len(want) {
		t.Fatalf("loaded %d entries, want %d", len(got.Entries), len(want))
	}
	for _, e := range got.Entries {
		if e.Amount != want[e.ID] {
			t.Errorf("entry %s: amount %d cents, want %d", e.ID, e.Amount, want[e.ID])
		}
	}

	s2, err := storage.NewSQLiteStorage(path)
	if err != nil {
		t.Fatalf("reopen migrated db: %v", err)
	}
	s2.Close()
}

func TestChaosLargeSnapshotSaveIsBounded(t *testing.T) {
	s, _ := openStore(t)
	var es []ledger.Entry
	for i := range 20000 {
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
