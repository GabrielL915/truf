package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/gabriel-luiz/truf/internal/ui/marks"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
)

func (m *Model) View() string {
	l := m.layout
	if l.width == 0 || l.height == 0 {
		return "Loading..."
	}
	if l.tooSmall {
		return m.renderTooSmall()
	}

	m.helpBar.SetHints(m.hints())

	return fit(m.tabBar.View(), l.width, tabBarHeight) + "\n" +
		fit(m.renderContent(), l.width, l.contentHeight) + "\n" +
		fit(m.helpBar.View(), l.width, helpBarHeight)
}

func (m *Model) renderContent() string {
	switch m.currentView {
	case ViewIncome:
		return m.incomeTable.View()
	case ViewExpenses:
		return m.expenseTable.View()
	case ViewCategories:
		return m.renderPlaceholder("Categories")
	case ViewSettings:
		return m.renderPlaceholder("Settings")
	default:
		return m.renderOverview()
	}
}

func (m *Model) renderOverview() string {
	if m.empty {
		return m.renderEmptyState()
	}
	return m.summary.View() + "\n\n" + m.chart.View()
}

func (m *Model) renderEmptyState() string {
	return m.centred(
		styles.AccentStyle.Render(marks.Large),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.Text).Render("Nothing here yet."),
		styles.MutedStyle.Render("Press 2 or 3, then n to add your first entry."),
		styles.MutedStyle.Render("Or run truf --seed for sample data."),
	)
}

func (m *Model) renderPlaceholder(title string) string {
	return m.centred(
		styles.AccentStyle.Render(marks.Large),
		"",
		styles.TitleStyle.Render(title),
		styles.MutedStyle.Render("coming soon"),
	)
}

func (m *Model) centred(blocks ...string) string {
	for i, b := range blocks {
		blocks[i] = lipgloss.NewStyle().Width(lipgloss.Width(b)).Align(lipgloss.Center).Render(b)
	}
	body := lipgloss.JoinVertical(lipgloss.Center, blocks...)
	return lipgloss.Place(m.layout.width, m.layout.contentHeight, lipgloss.Center, lipgloss.Center, body)
}

func (m *Model) renderTooSmall() string {
	msg := styles.MutedStyle.Render("Terminal too small: TRUF needs at least 80×24.")
	return lipgloss.Place(m.layout.width, m.layout.height, lipgloss.Center, lipgloss.Center, msg)
}

func fit(block string, width, height int) string {
	lines := strings.Split(block, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		if lipgloss.Width(line) > width {
			lines[i] = ansi.Truncate(line, width, "")
		}
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
