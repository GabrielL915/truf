package components

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/gabriel-luiz/truf/internal/ui/marks"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

type TabBar struct {
	Labels []string
	Active int
	Month  time.Time
	Width  int
}

func NewTabBar(labels []string) *TabBar {
	return &TabBar{Labels: labels}
}

func (b *TabBar) SetActive(i int) {
	b.Active = i
}

func (b *TabBar) SetMonth(month time.Time) {
	b.Month = month
}

func (b *TabBar) SetWidth(width int) {
	b.Width = width
}

func (b *TabBar) View() string {
	bar := styles.BarStyle
	active := lipgloss.NewStyle().Background(styles.Accent).Foreground(styles.OnAccent).Bold(true)

	var left strings.Builder
	left.WriteString(bar.Render(" "))
	left.WriteString(styles.AccentStyle.Background(styles.Surface).Render(marks.Small))
	left.WriteString(bar.Bold(true).Render(" TRUF"))
	for i, label := range b.Labels {
		tab := " " + strconv.Itoa(i+1) + " " + label + " "
		if i == b.Active {
			left.WriteString(active.Render(tab))
		} else {
			left.WriteString(bar.Render(tab))
		}
	}

	right := ""
	for _, candidate := range []string{
		"‹ " + b.Month.Format("Jan 2006") + " › ",
		"‹ " + utils.FormatMonthYear(b.Month) + " › ",
		"‹ " + utils.FormatMonthYear(b.Month) + " ›",
	} {
		if lipgloss.Width(left.String())+lipgloss.Width(candidate) <= b.Width {
			right = bar.Render(candidate)
			break
		}
	}

	gap := b.Width - lipgloss.Width(left.String()) - lipgloss.Width(right)
	line := left.String()
	if gap >= 0 {
		line += bar.Render(strings.Repeat(" ", gap)) + right
	} else {
		line = ansi.Truncate(line, b.Width, "")
	}
	return bar.Width(b.Width).Render(line)
}
