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

	sb.WriteString(" " + styles.Logo())
	sb.WriteString("\n\n")

	for i, item := range m.Items {
		var line string
		switch {
		case i == m.Selected && m.Focused:
			line = styles.AccentStyle.Render("▌") + styles.TitleStyle.Render(item.Label)
		case i == m.Selected:
			line = styles.ChromeStyle.Render("▌") + styles.AccentStyle.Render(item.Label)
		default:
			line = " " + styles.TextStyle.Render(item.Label)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		MaxHeight(m.Height).
		Render(sb.String())
}
