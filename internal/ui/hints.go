package ui

import "github.com/gabriel-luiz/truf/internal/ui/components"

func (m *Model) hints() []components.Hint {
	if t := m.activeTable(); t != nil && t.Editing {
		return []components.Hint{
			{Key: "type", Verb: "edit cell"},
			{Key: "enter", Verb: "next column"},
			{Key: "esc", Verb: "cancel"},
		}
	}
	if m.currentView == ViewOverview && m.empty {
		return []components.Hint{
			{Key: "2", Verb: "income"},
			{Key: "3", Verb: "expenses"},
			{Key: "n", Verb: "new entry"},
			{Key: "q", Verb: "quit"},
		}
	}
	switch m.currentView {
	case ViewOverview:
		return []components.Hint{
			{Key: "1-5", Verb: "view"},
			{Key: "[ ]", Verb: "month"},
			{Key: "pgup/pgdn", Verb: "range"},
			{Key: "q", Verb: "quit"},
		}
	case ViewIncome, ViewExpenses:
		return []components.Hint{
			{Key: "↑↓", Verb: "move"},
			{Key: "enter", Verb: "edit"},
			{Key: "n", Verb: "new"},
			{Key: "d", Verb: "delete"},
			{Key: "[ ]", Verb: "month"},
			{Key: "esc", Verb: "overview"},
		}
	}
	return []components.Hint{
		{Key: "1-5", Verb: "view"},
		{Key: "[ ]", Verb: "month"},
		{Key: "esc", Verb: "overview"},
	}
}
