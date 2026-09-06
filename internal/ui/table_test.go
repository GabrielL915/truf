package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
)

func tableModel(t *testing.T) *Model {
	t.Helper()
	m, _ := chaosModel(t)
	sized(t, m)
	feb := time.Date(2026, time.February, 3, 0, 0, 0, 0, time.UTC)
	for _, e := range []ledger.Entry{
		{Kind: ledger.Expense, Date: feb, Description: "rent", Category: "Housing", Amount: 60000},
		{Kind: ledger.Expense, Date: feb.AddDate(0, 0, 4), Description: "groceries with a very long description that must be cut", Category: "Food", Amount: 30000},
		{Kind: ledger.Expense, Date: feb.AddDate(0, 0, 9), Description: "bus", Category: "Transportation", Amount: 10000},
	} {
		if _, err := m.ledger.Add(e); err != nil {
			t.Fatal(err)
		}
	}
	key(m, "[")
	key(m, "3")
	return m
}

func TestTableHeaderLineCountMonthAndTotal(t *testing.T) {
	m := tableModel(t)
	lines := assertShape(t, m)
	if !strings.Contains(lines[1], "Expenses · 3 entries · February 2026") || !strings.Contains(lines[1], "Total $1.000,00") {
		t.Errorf("header line = %q", lines[1])
	}
	if !strings.Contains(lines[2], "DATE") || !strings.Contains(lines[2], "AMOUNT") {
		t.Errorf("column headers = %q", lines[2])
	}
}

func TestTableRowsCursorAmountsAndTruncation(t *testing.T) {
	m := tableModel(t)
	lines := frame(m)
	rows := strings.Join(lines[4:7], "\n")
	for _, want := range []string{"03/02", "rent", "Housing", "$600,00", "07/02", "$300,00", "12/02", "bus", "$100,00", "…"} {
		if !strings.Contains(rows, want) {
			t.Errorf("rows lack %q:\n%s", want, rows)
		}
	}
	if strings.Contains(rows, "2026") {
		t.Errorf("rows should not repeat the year:\n%s", rows)
	}
	if !strings.HasPrefix(lines[4], "▌") {
		t.Errorf("first row should carry the cursor marker: %q", lines[4])
	}
	key(m, "down")
	lines = frame(m)
	if strings.HasPrefix(lines[4], "▌") || !strings.HasPrefix(lines[5], "▌") {
		t.Errorf("cursor marker did not move to the second row:\n%s\n%s", lines[4], lines[5])
	}
}

func TestTableFooterShowsDistribution(t *testing.T) {
	m := tableModel(t)
	lines := frame(m)
	footer := lines[len(lines)-2]
	if !strings.Contains(footer, "Housing 60%") || !strings.Contains(footer, "Food 30%") || !strings.Contains(footer, "Transportation 10%") {
		t.Errorf("footer = %q", footer)
	}
	if strings.Contains(footer, " of ") {
		t.Errorf("no pagination expected when every row fits: %q", footer)
	}
	if !strings.HasPrefix(strings.TrimSpace(lines[len(lines)-3]), "─") {
		t.Errorf("footer should sit under a rule: %q", lines[len(lines)-3])
	}
}

func TestTableFooterPaginatesWhenRowsOverflow(t *testing.T) {
	m := tableModel(t)
	for i := range 40 {
		if _, err := m.ledger.Add(ledger.Entry{Kind: ledger.Expense, Date: m.month, Description: fmt.Sprintf("row %d", i), Category: fmt.Sprintf("Cat %d", i%8), Amount: 100}); err != nil {
			t.Fatal(err)
		}
	}
	m.refreshTables()
	footer := frame(m)[frameHeight-2]
	if !strings.Contains(footer, "1–") || !strings.Contains(footer, " of 43") {
		t.Errorf("footer should paginate: %q", footer)
	}
	for range 42 {
		key(m, "down")
	}
	footer = frame(m)[frameHeight-2]
	if !strings.Contains(footer, "–43 of 43") {
		t.Errorf("footer should follow the cursor to the end: %q", footer)
	}
	if !strings.Contains(footer, "Other") {
		t.Errorf("distribution should collapse small categories into Other: %q", footer)
	}
}

func TestTableEmptyMonthMessageAndNoFooter(t *testing.T) {
	m := tableModel(t)
	key(m, "]")
	key(m, "]")
	lines := assertShape(t, m)
	out := strings.Join(lines, "\n")
	if !strings.Contains(out, "No entries in April 2026. Press n to add one.") {
		t.Errorf("empty month message missing:\n%s", out)
	}
	if strings.Contains(out, " of ") || strings.Contains(out, "%") {
		t.Errorf("empty month should have no footer:\n%s", out)
	}
}

func TestTableEditingHighlightsCellAndSwitchesHints(t *testing.T) {
	m := tableModel(t)
	key(m, "enter")
	key(m, "tab")
	typeText(m, "!")
	lines := frame(m)
	if !strings.Contains(lines[4], "rent!▏") {
		t.Errorf("editing cell should show the buffer and a text cursor: %q", lines[4])
	}
	if last := lines[frameHeight-1]; !strings.Contains(last, "next column") || !strings.Contains(last, "cancel") {
		t.Errorf("help bar should show editing hints: %q", last)
	}
	key(m, "esc")
	if last := frame(m)[frameHeight-1]; strings.Contains(last, "cancel") {
		t.Errorf("help bar should return to table hints: %q", last)
	}
}
