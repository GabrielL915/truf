package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
)

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	leftPanel := m.menu.View()
	rightPanel := m.renderRightPanel()

	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	statusBar := m.statusBar.View()

	fullView := lipgloss.JoinVertical(lipgloss.Left, mainArea, statusBar)

	return styles.BaseStyle.Render(fullView)
}

func (m *Model) renderRightPanel() string {
	switch m.currentView {
	case ViewOverview:
		return m.chart.View()
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
	leftWidth := m.width * 25 / 100
	rightWidth := m.width - leftWidth
	mainHeight := m.height - 3

	content := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		Render(title) + "\n\n" +
		lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Render(message)

	return styles.PanelStyle.
		Width(rightWidth - 2).
		Height(mainHeight - 2).
		Render(content)
}
