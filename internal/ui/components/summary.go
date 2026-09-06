package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

const SummaryHeight = 3

type Summary struct {
	Month        time.Time
	Income       int64
	Expenses     int64
	Net          int64
	Balance      int64
	IncomeCount  int
	ExpenseCount int
	NetChange    int64
	Since        time.Time
	Width        int
}

func NewSummary() *Summary {
	return &Summary{}
}

func (s *Summary) SetWidth(width int) {
	s.Width = width
}

type card struct {
	label     string
	value     string
	color     lipgloss.TerminalColor
	secondary string
}

func (s *Summary) cards() []card {
	prev := utils.AddMonths(s.Month, -1)
	change := utils.FormatCurrency(s.NetChange)
	if s.NetChange >= 0 {
		change = "+" + change
	}
	since := "no entries yet"
	if !s.Since.IsZero() {
		since = "since " + utils.FormatMonthYear(s.Since)
	}
	return []card{
		{"Income", utils.FormatCurrency(s.Income), styles.Success, entries(s.IncomeCount)},
		{"Expenses", utils.FormatCurrency(s.Expenses), styles.Danger, entries(s.ExpenseCount)},
		{"Net · " + utils.GetMonthShort(s.Month.Month()), utils.FormatCurrency(s.Net), styles.SignColor(s.Net), change + " vs " + utils.GetMonthShort(prev.Month())},
		{"Balance · all time", utils.FormatCurrency(s.Balance), styles.SignColor(s.Balance), since},
	}
}

func entries(n int) string {
	if n == 1 {
		return "1 entry"
	}
	return fmt.Sprintf("%d entries", n)
}

func (s *Summary) View() string {
	cards := s.cards()
	inner := max(s.Width-2, 4*8)
	colWidth := inner / len(cards)
	lines := [SummaryHeight]strings.Builder{}
	for _, c := range cards {
		lines[0].WriteString(styles.MutedStyle.Width(colWidth).Render(c.label))
		lines[1].WriteString(lipgloss.NewStyle().Foreground(c.color).Bold(true).Width(colWidth).Render(c.value))
		lines[2].WriteString(styles.MutedStyle.Width(colWidth).Render(c.secondary))
	}
	out := make([]string, SummaryHeight)
	for i := range lines {
		out[i] = " " + lines[i].String()
	}
	return strings.Join(out, "\n")
}
