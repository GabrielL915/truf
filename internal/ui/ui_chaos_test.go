package ui

import (
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/storage"
	"github.com/gabriel-luiz/truf/internal/ui/components"
)

var chaosNow = time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)

func chaosModel(t *testing.T) (*Model, *storage.MemoryStorage) {
	t.Helper()
	store := storage.NewMemoryStorage()
	l, err := ledger.New(store, func() time.Time { return chaosNow })
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(l)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return m, store
}

func key(m *Model, k string) {
	var msg tea.KeyMsg
	switch k {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	case " ":
		msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	m.Update(msg)
}

func clearBuffer(m *Model) {
	for range m.incomeTable.EditBuffer {
		key(m, "backspace")
	}
}

func typeText(m *Model, s string) {
	for _, r := range s {
		key(m, string(r))
	}
}

func openIncomeAndNew(t *testing.T, m *Model) {
	t.Helper()
	key(m, "down")
	key(m, "enter")
	key(m, "n")
	if !m.incomeTable.Editing {
		t.Fatal("expected to be editing a new income row")
	}
}

func TestChaosTypingAccentedCharacterIsKept(t *testing.T) {
	m, _ := chaosModel(t)
	openIncomeAndNew(t, m)
	key(m, "tab")
	typeText(m, "çã")
	key(m, " ")
	key(m, "😀")
	if got := m.incomeTable.EditBuffer; got != "çã 😀" {
		t.Errorf("buffer = %q, want %q", got, "çã 😀")
	}
	key(m, "enter")
	key(m, "enter")
	key(m, "enter")
	if got := m.ledger.Entries(chaosNow, ledger.Income)[0].Description; got != "çã 😀" {
		t.Errorf("persisted description = %q", got)
	}
}

func TestChaosBackspaceOnMultibyteKeepsValidUTF8(t *testing.T) {
	tbl := components.NewEntryTable("x", ledger.Income)
	tbl.SetEntries([]ledger.Entry{{ID: "1", Date: chaosNow}})
	tbl.StartEdit()
	if err := tbl.NextColumn(); err != nil {
		t.Fatal(err)
	}
	tbl.EditBuffer = ""
	tbl.TypeChar('a')
	tbl.TypeChar('ç')
	tbl.TypeChar('😀')
	tbl.Backspace()
	if tbl.EditBuffer != "aç" {
		t.Errorf("after backspace of emoji, buffer = %q", tbl.EditBuffer)
	}
	tbl.Backspace()
	if tbl.EditBuffer != "a" || !utf8.ValidString(tbl.EditBuffer) {
		t.Errorf("after backspace of ç, buffer = %q (valid utf8: %v)", tbl.EditBuffer, utf8.ValidString(tbl.EditBuffer))
	}
	tbl.Backspace()
	tbl.Backspace()
	if tbl.EditBuffer != "" {
		t.Errorf("over-backspace buffer = %q", tbl.EditBuffer)
	}
}

func TestChaosInvalidAmountEditSurfacesFeedbackAndKeepsEditing(t *testing.T) {
	m, _ := chaosModel(t)
	openIncomeAndNew(t, m)
	key(m, "tab")
	key(m, "tab")
	key(m, "tab")
	clearBuffer(m)
	typeText(m, "ab")
	key(m, "enter")
	if !m.incomeTable.Editing || m.incomeTable.EditingColumn != 3 {
		t.Fatalf("invalid amount should keep the amount cell in edit mode (editing=%v col=%d)", m.incomeTable.Editing, m.incomeTable.EditingColumn)
	}
	if m.err == nil {
		t.Error("amount 'ab' discarded with no error surfaced to status bar")
	}
	clearBuffer(m)
	typeText(m, "1,5")
	key(m, "enter")
	if m.incomeTable.Editing {
		t.Fatal("still editing after valid amount")
	}
	if m.err != nil {
		t.Errorf("error not cleared after valid amount: %v", m.err)
	}
	if got := m.ledger.Entries(chaosNow, ledger.Income)[0].Amount; got != 150 {
		t.Errorf("amount = %d cents, want 150", got)
	}
}

func TestChaosInvalidDateEditSurfacesFeedback(t *testing.T) {
	m, _ := chaosModel(t)
	openIncomeAndNew(t, m)
	clearBuffer(m)
	typeText(m, "02/30/2026")
	key(m, "enter")
	if !m.incomeTable.Editing || m.incomeTable.EditingColumn != 0 {
		t.Fatal("invalid date should keep the date cell in edit mode")
	}
	if m.err == nil {
		t.Error("invalid date accepted silently")
	}
	key(m, "esc")
	if m.incomeTable.Editing {
		t.Error("esc did not cancel edit")
	}
}

func TestChaosEditingDateIntoOtherMonthKeepsUIConsistent(t *testing.T) {
	m, _ := chaosModel(t)
	openIncomeAndNew(t, m)
	clearBuffer(m)
	typeText(m, "02/10/2026")
	key(m, "enter")
	key(m, "enter")
	key(m, "enter")
	key(m, "enter")
	if m.incomeTable.Editing {
		t.Fatalf("still editing: err=%v", m.err)
	}
	if n := len(m.incomeTable.Entries); n != 0 {
		t.Errorf("entry moved to February still shown in March table (%d rows)", n)
	}
	if n := len(m.ledger.Entries(time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC), ledger.Income)); n != 1 {
		t.Errorf("ledger has %d February entries, want 1", n)
	}
	key(m, "enter")
	key(m, "d")
	_ = m.View()
}

func TestChaosDeleteOnEmptyTableDoesNotPanic(t *testing.T) {
	m, _ := chaosModel(t)
	key(m, "down")
	key(m, "enter")
	key(m, "d")
	key(m, "enter")
	key(m, "tab")
	_ = m.View()
}

func TestChaosDeleteLastRowKeepsCursorInRange(t *testing.T) {
	m, _ := chaosModel(t)
	key(m, "down")
	key(m, "enter")
	for range 3 {
		key(m, "n")
		key(m, "esc")
	}
	key(m, "down")
	key(m, "down")
	key(m, "down")
	key(m, "d")
	key(m, "d")
	key(m, "d")
	if _, ok := m.incomeTable.Current(); ok {
		t.Error("Current() returned an entry on empty table")
	}
	_ = m.View()
}

func TestChaosSaveFailureKeepsTypedDataAndShowsError(t *testing.T) {
	m, store := chaosModel(t)
	store.SaveErr = errors.New("disk full")
	key(m, "down")
	key(m, "enter")
	key(m, "n")
	if m.err == nil {
		t.Error("save error not surfaced")
	}
	if len(m.incomeTable.Entries) != 1 {
		t.Errorf("table shows %d rows, want the typed row kept", len(m.incomeTable.Entries))
	}
	_ = m.View()
}

func TestChaosRenderTinyAndHugeSizes(t *testing.T) {
	m, _ := chaosModel(t)
	openIncomeAndNew(t, m)
	key(m, "tab")
	typeText(m, strings.Repeat("字", 60))
	key(m, "enter")
	key(m, "enter")
	key(m, "enter")
	for _, sz := range [][2]int{{1, 1}, {5, 3}, {20, 5}, {2000, 500}} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("View at %v panicked: %v", sz, r)
				}
			}()
			m.Update(tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
			out := m.View()
			if !utf8.ValidString(out) {
				t.Errorf("View at %v produced invalid UTF-8", sz)
			}
		}()
	}
	key(m, "esc")
	for _, sz := range [][2]int{{1, 1}, {10, 4}} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("chart View at %v panicked: %v", sz, r)
				}
			}()
			m.Update(tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
			_ = m.View()
		}()
	}
}

func TestChaosChartWithSinglePointAndFlatData(t *testing.T) {
	c := components.NewChart()
	c.SetSize(40, 10)
	for _, d := range []ledger.ChartData{
		{Months: []string{"Mar 26"}, Income: []int64{0}, Expenses: []int64{0}, Balance: []int64{0}},
		{Months: []string{"Feb 26", "Mar 26"}, Income: []int64{-1e15, 1e15}, Expenses: []int64{0, 0}, Balance: []int64{0, 0}},
		{Months: []string{"Feb 26", "Mar 26"}, Income: []int64{1}, Expenses: []int64{1, 2}, Balance: []int64{1, 2}},
		{Months: []string{"Feb 26", "Mar 26"}, Income: []int64{1 << 62, 1 << 62}, Expenses: []int64{0, 0}, Balance: []int64{-(1 << 62), 1 << 62}},
	} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("chart panicked on %v: %v", d.Months, r)
				}
			}()
			c.SetData(d)
			_ = c.View()
		}()
	}
}

func TestChaosChartRangeKeysStayWithinBounds(t *testing.T) {
	m, _ := chaosModel(t)
	key(m, "enter")
	for range 50 {
		m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	}
	if m.chartMonths > 24 {
		t.Errorf("chartMonths = %d", m.chartMonths)
	}
	for range 50 {
		m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	}
	if m.chartMonths < 3 {
		t.Errorf("chartMonths = %d", m.chartMonths)
	}
}
