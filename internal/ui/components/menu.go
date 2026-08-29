package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
)

type MenuItem struct {
	Label string
	Key   string
}

type Menu struct {
	Items    []MenuItem
	Selected int
	Focused  bool
	Width    int
	Height   int
}

func NewMenu() *Menu {
	return &Menu{
		Items: []MenuItem{
			{Label: "Overview", Key: "overview"},
			{Label: "Income", Key: "income"},
			{Label: "Expenses", Key: "expenses"},
			{Label: "Categories", Key: "categories"},
			{Label: "Settings", Key: "settings"},
		},
		Selected: 0,
		Focused:  true,
	}
}

func (m *Menu) Up() {
	if m.Selected > 0 {
		m.Selected--
	}
}

func (m *Menu) Down() {
	if m.Selected < len(m.Items)-1 {
		m.Selected++
	}
}

func (m *Menu) SelectedItem() MenuItem {
	if m.Selected >= 0 && m.Selected < len(m.Items) {
		return m.Items[m.Selected]
	}
	return MenuItem{}
}

func (m *Menu) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

func (m *Menu) View() string {
	var sb strings.Builder

	title := styles.MenuTitleStyle.Render(styles.Logo())
	sb.WriteString(title)
	sb.WriteString("\n\n")

	for i, item := range m.Items {
		var line string
		if i == m.Selected {
			cursor := ">"
			if m.Focused {
				cursor = lipgloss.NewStyle().Foreground(styles.Primary).Render(">")
			}
			line = cursor + " " + styles.MenuItemSelectedStyle.Render(item.Label)
		} else {
			line = styles.MenuItemStyle.Render(item.Label)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	content := sb.String()

	var panelStyle lipgloss.Style
	if m.Focused {
		panelStyle = styles.FocusedPanelStyle
	} else {
		panelStyle = styles.PanelStyle
	}

	return panelStyle.
		Width(m.Width - 2).
		Height(m.Height - 2).
		Render(content)
}
