package storage

import (
	"path/filepath"
	"testing"

	"github.com/gabriel-luiz/truf/internal/fixture"
)

func benchDB(b *testing.B) *SQLiteStorage {
	b.Helper()
	s, err := NewSQLiteStorage(filepath.Join(b.TempDir(), "bench.db"))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { s.Close() })
	return s
}

func BenchmarkSQLiteSave10k(b *testing.B) {
	s := benchDB(b)
	snap := fixture.Snapshot(10_000, 24)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := s.Save(snap); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSQLiteLoad10k(b *testing.B) {
	s := benchDB(b)
	if err := s.Save(fixture.Snapshot(10_000, 24)); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Load(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSQLiteOpenLoad1k(b *testing.B) {
	path := filepath.Join(b.TempDir(), "startup.db")
	s, err := NewSQLiteStorage(path)
	if err != nil {
		b.Fatal(err)
	}
	if err := s.Save(fixture.Snapshot(1_000, 12)); err != nil {
		b.Fatal(err)
	}
	s.Close()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, err := NewSQLiteStorage(path)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := s.Load(); err != nil {
			b.Fatal(err)
		}
		s.Close()
	}
}
