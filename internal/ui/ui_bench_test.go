package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabriel-luiz/truf/internal/fixture"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/storage"
)

func benchModel(b *testing.B, entries int) *Model {
	b.Helper()
	store := storage.NewMemoryStorage()
	store.Snapshot = fixture.Snapshot(entries, 24)
	l, err := ledger.New(store, func() time.Time { return fixture.Now })
	if err != nil {
		b.Fatal(err)
	}
	m := NewModel(l)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return m
}

func BenchmarkViewOverview(b *testing.B) {
	m := benchModel(b, 10_000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

func BenchmarkViewTable(b *testing.B) {
	m := benchModel(b, 10_000)
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

func BenchmarkKeyDownThenView(b *testing.B) {
	m := benchModel(b, 10_000)
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		_ = m.View()
	}
}

func BenchmarkResize(b *testing.B) {
	m := benchModel(b, 10_000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := 100 + i%40
		m.Update(tea.WindowSizeMsg{Width: w, Height: 40})
		_ = m.View()
	}
}
