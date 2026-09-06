package ui

import "github.com/gabriel-luiz/truf/internal/ui/components"

func (m *Model) hints() []components.Hint {
	if t := m.activeTable(); t != nil && t.Editing {
		return []components.Hint{
			{Key: "enter", Verb: "next column"},
			{Key: "esc", Verb: "cancel"},
		}
	}
	if m.focusedPanel == PanelMenu {
		return []components.Hint{
			{Key: "↑↓", Verb: "move"},
			{Key: "enter", Verb: "open"},
			{Key: "tab", Verb: "content"},
			{Key: "q", Verb: "quit"},
		}
	}
	switch m.currentView {
	case ViewOverview:
		return []components.Hint{
			{Key: "pgup/pgdn", Verb: "range"},
			{Key: "tab", Verb: "menu"},
			{Key: "esc", Verb: "back"},
		}
	case ViewIncome, ViewExpenses:
		return []components.Hint{
			{Key: "↑↓", Verb: "move"},
			{Key: "enter", Verb: "edit"},
			{Key: "n", Verb: "new"},
			{Key: "d", Verb: "delete"},
			{Key: "[ ]", Verb: "month"},
			{Key: "esc", Verb: "back"},
		}
	}
	return []components.Hint{
		{Key: "tab", Verb: "menu"},
		{Key: "esc", Verb: "back"},
	}
}
