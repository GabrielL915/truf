package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

type StatusBar struct {
	TimeRange    string
	TotalBalance int64
	Err          error
	Width        int
}

func NewStatusBar() *StatusBar {
	return &StatusBar{
		TimeRange: "Last 6 months",
	}
}

func (s *StatusBar) SetTimeRange(start, end string) {
	s.TimeRange = fmt.Sprintf("%s - %s", start, end)
}

func (s *StatusBar) SetBalance(balance int64) {
	s.TotalBalance = balance
}

func (s *StatusBar) SetError(err error) {
	s.Err = err
}

func (s *StatusBar) SetWidth(width int) {
	s.Width = width
}

func (s *StatusBar) View() string {
	if s.Err != nil {
		return styles.StatusBarStyle.
			Width(s.Width).
			Render(lipgloss.NewStyle().Foreground(styles.Danger).Render("Error: " + s.Err.Error()))
	}

	timeRange := styles.StatusKeyStyle.Render("Range: ") +
		styles.StatusValueStyle.Render(s.TimeRange)

	balanceStyle := styles.StatusValueStyle
	if s.TotalBalance < 0 {
		balanceStyle = lipgloss.NewStyle().Foreground(styles.Danger)
	} else if s.TotalBalance > 0 {
		balanceStyle = lipgloss.NewStyle().Foreground(styles.Success)
	}
	balance := styles.StatusKeyStyle.Render("Balance: ") +
		balanceStyle.Render(utils.FormatCurrency(s.TotalBalance))

	help := styles.HelpStyle.Render("Tab: switch panel • ↑↓: navigate • q: quit")

	availableSpace := s.Width - lipgloss.Width(timeRange) - lipgloss.Width(balance) - lipgloss.Width(help) - 4

	leftPad := availableSpace / 2
	rightPad := availableSpace - leftPad
	if leftPad < 1 {
		leftPad = 1
	}
	if rightPad < 1 {
		rightPad = 1
	}

	content := timeRange +
		strings.Repeat(" ", leftPad) +
		balance +
		strings.Repeat(" ", rightPad) +
		help

	return styles.StatusBarStyle.
		Width(s.Width).
		Render(content)
}
