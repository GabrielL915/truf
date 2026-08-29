package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Primary    = lipgloss.Color("#FC9F5B")
	Secondary  = lipgloss.Color("#FBD1A2")
	Accent     = lipgloss.Color("#7DCFB6")
	Success    = lipgloss.Color("#33CA7F")
	Danger     = lipgloss.Color("#FF6B6B")
	Muted      = lipgloss.Color("#666666")
	Background = lipgloss.Color("#1E1E1E")
	Surface    = lipgloss.Color("#2D2D2D")
	Text       = lipgloss.Color("#FFFFFF")
	TextMuted  = lipgloss.Color("#888888")

	BaseStyle = lipgloss.NewStyle().
			Background(Background).
			Foreground(Text)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted).
			Padding(1)

	FocusedPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(1)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(Text).
			PaddingLeft(2)

	MenuItemSelectedStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true).
				PaddingLeft(1)

	MenuTitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			MarginBottom(1)

	ChartIncomeColor  = Success
	ChartExpenseColor = Danger
	ChartBalanceColor = Accent

	StatusBarStyle = lipgloss.NewStyle().
			Background(Surface).
			Foreground(Text).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	StatusValueStyle = lipgloss.NewStyle().
				Foreground(Text)

	HelpStyle = lipgloss.NewStyle().
			Foreground(TextMuted)
)

func Logo() string {
	return lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Render("TRUF")
}
