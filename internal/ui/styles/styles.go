package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Accent  = lipgloss.Color("#FC9F5B")
	Success = lipgloss.Color("#33CA7F")
	Danger  = lipgloss.Color("#FF6B6B")
	Net     = lipgloss.Color("#7DCFB6")

	Text     = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#F2F2F2"}
	Muted    = lipgloss.AdaptiveColor{Light: "#6B6B6B", Dark: "#8A8A8A"}
	Chrome   = lipgloss.AdaptiveColor{Light: "#C4C4C4", Dark: "#4A4A4A"}
	Surface  = lipgloss.AdaptiveColor{Light: "#E6E6E6", Dark: "#2D2D2D"}
	OnAccent = lipgloss.Color("#1E1E1E")

	TextStyle   = lipgloss.NewStyle().Foreground(Text)
	MutedStyle  = lipgloss.NewStyle().Foreground(Muted)
	ChromeStyle = lipgloss.NewStyle().Foreground(Chrome)
	AccentStyle = lipgloss.NewStyle().Foreground(Accent)
	TitleStyle  = lipgloss.NewStyle().Foreground(Accent).Bold(true)
	DangerStyle = lipgloss.NewStyle().Foreground(Danger)

	BarStyle = lipgloss.NewStyle().Background(Surface).Foreground(Text)
	KeyStyle = lipgloss.NewStyle().Background(Surface).Foreground(Accent).Bold(true)

	ChartIncomeColor  = Success
	ChartExpenseColor = Danger
	ChartBalanceColor = Net
)

func KindColor(income bool) lipgloss.TerminalColor {
	if income {
		return Success
	}
	return Danger
}

func SignColor(v int64) lipgloss.TerminalColor {
	switch {
	case v > 0:
		return Success
	case v < 0:
		return Danger
	}
	return Text
}

func Logo() string {
	return TitleStyle.Render("TRUF")
}
