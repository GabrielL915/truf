package ui

import (
	"strings"
	"testing"

	"github.com/gabriel-luiz/truf/internal/ui/marks"
)

func markLine() string {
	for _, line := range strings.Split(marks.Large, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s
		}
	}
	return ""
}

func TestOverviewSummaryCardsShowMonthTotals(t *testing.T) {
	m := sized(t, monthModel(t))
	lines := assertShape(t, m)
	top := strings.Join(lines[1:5], "\n")
	for _, want := range []string{"Income", "Expenses", "Net · Mar", "Balance · all time", "$1,00", "$0,00", "$2,00", "1 entry", "0 entries", "+$0,00 vs Feb", "since Feb 26"} {
		if !strings.Contains(top, want) {
			t.Errorf("summary lacks %q:\n%s", want, top)
		}
	}

	key(m, "[")
	top = strings.Join(frame(m)[1:5], "\n")
	for _, want := range []string{"Net · Feb", "$5,00", "$4,00", "2 entries", "+$1,00 vs Jan"} {
		if !strings.Contains(top, want) {
			t.Errorf("summary for Feb lacks %q:\n%s", want, top)
		}
	}
}

func TestOverviewChartHeaderHasLegendAndNoRunningBalance(t *testing.T) {
	m := sized(t, monthModel(t))
	out := frameText(m)
	header := ""
	for _, line := range frame(m) {
		if strings.Contains(line, "Balance · last 6 months") {
			header = line
		}
	}
	if header == "" {
		t.Fatalf("chart header missing:\n%s", out)
	}
	for _, want := range []string{"⣿ Income", "⣿ Expenses", "⣿ Balance"} {
		if !strings.Contains(header, want) {
			t.Errorf("legend %q not in the chart header: %q", want, header)
		}
	}
	if strings.Contains(header, "$") {
		t.Errorf("chart header still shows a running balance: %q", header)
	}
	if !strings.Contains(out, "┈") {
		t.Errorf("chart has no gridlines:\n%s", out)
	}
	if !strings.Contains(out, "Mar 26") || !strings.Contains(out, "Oct 25") {
		t.Errorf("x axis should print every month label:\n%s", out)
	}
}

func TestOverviewEmptyStateShowsMarkAndHints(t *testing.T) {
	m, _ := chaosModel(t)
	sized(t, m)
	lines := assertShape(t, m)
	out := strings.Join(lines, "\n")
	for _, want := range []string{markLine(), "Nothing here yet.", "Press 2 or 3", "--seed"} {
		if !strings.Contains(out, want) {
			t.Errorf("empty state lacks %q:\n%s", want, out)
		}
	}
	if !strings.Contains(lines[len(lines)-1], "2 income") {
		t.Errorf("help bar should show onboarding hints: %q", lines[len(lines)-1])
	}
	key(m, "2")
	key(m, "n")
	key(m, "esc")
	key(m, "esc")
	if strings.Contains(frameText(m), "Nothing here yet.") {
		t.Errorf("empty state should leave once an entry exists")
	}
}

func TestOverviewPlaceholdersShowMarkAndComingSoon(t *testing.T) {
	m := sized(t, monthModel(t))
	for k, title := range map[string]string{"4": "Categories", "5": "Settings"} {
		key(m, k)
		out := strings.Join(assertShape(t, m)[1:], "\n")
		for _, want := range []string{markLine(), title, "coming soon"} {
			if !strings.Contains(out, want) {
				t.Errorf("view %s lacks %q:\n%s", k, want, out)
			}
		}
	}
}
