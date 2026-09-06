package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

	content := lipgloss.JoinHorizontal(lipgloss.Top, m.menu.View(), m.renderContent())

	m.helpBar.SetHints(m.hints())

	return fit(m.renderTabBar(), l.width, tabBarHeight) + "\n" +
		fit(content, l.width, l.contentHeight) + "\n" +
		fit(m.helpBar.View(), l.width, helpBarHeight)
}

func (m *Model) renderTabBar() string {
	return styles.BarStyle.Width(m.layout.width).Render(" " + styles.Logo())
}

func (m *Model) renderContent() string {
	switch m.currentView {
	case ViewIncome:
		return m.incomeTable.View()
	case ViewExpenses:
		return m.expenseTable.View()
	case ViewCategories:
		return m.renderPlaceholder("Categories", "Category management coming soon...")
	case ViewSettings:
		return m.renderPlaceholder("Settings", "Settings coming soon...")
	default:
		return m.chart.View()
	}
}

func (m *Model) renderPlaceholder(title, message string) string {
	return lipgloss.NewStyle().
		Width(m.layout.mainWidth).
		Height(m.layout.contentHeight).
		Render(styles.TitleStyle.Render(title) + "\n\n" + styles.MutedStyle.Render(message))
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
