package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/ui/components"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if t := m.activeTable(); t != nil && t.Editing {
		return m.handleTableEdit(msg, t)
	}

	switch msg.String() {
	case "q":
		if m.currentView == ViewOverview && m.focusedPanel == PanelMenu {
			return m, tea.Quit
		}
		return m, nil

	case "ctrl+c":
		return m, tea.Quit

	case "tab":
		if m.focusedPanel == PanelMenu {
			m.focusedPanel = PanelContent
		} else {
			m.focusedPanel = PanelMenu
		}
		m.updateFocus()
		return m, nil

	case "esc":
		if m.currentView != ViewOverview {
			m.currentView = ViewOverview
			m.focusedPanel = PanelMenu
			m.updateFocus()
			m.refreshChart()
		}
		return m, nil

	case "up", "k":
		return m.handleUp()

	case "down", "j":
		return m.handleDown()

	case "pgup":
		if m.currentView == ViewOverview && m.chartMonths < 24 {
			m.chartMonths += 3
			m.refreshChart()
		}
		return m, nil

	case "pgdown":
		if m.currentView == ViewOverview && m.chartMonths > 3 {
			m.chartMonths -= 3
			m.refreshChart()
		}
		return m, nil

	case "enter":
		return m.handleEnter()

	case "n":
		return m.handleNewEntry()

	case "d":
		return m.handleDeleteEntry()
	}

	return m, nil
}

func (m *Model) handleUp() (tea.Model, tea.Cmd) {
	if m.focusedPanel == PanelMenu {
		m.menu.Up()
	} else if t := m.activeTable(); t != nil {
		t.Up()
	}
	return m, nil
}

func (m *Model) handleDown() (tea.Model, tea.Cmd) {
	if m.focusedPanel == PanelMenu {
		m.menu.Down()
	} else if t := m.activeTable(); t != nil {
		t.Down()
	}
	return m, nil
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.focusedPanel == PanelMenu {
		switch m.menu.SelectedItem().Key {
		case "overview":
			m.currentView = ViewOverview
			m.refreshChart()
		case "income":
			m.currentView = ViewIncome
			m.refreshTables()
		case "expenses":
			m.currentView = ViewExpenses
			m.refreshTables()
		case "categories":
			m.currentView = ViewCategories
		case "settings":
			m.currentView = ViewSettings
		}
		m.focusedPanel = PanelContent
		m.updateFocus()
		return m, nil
	}

	if t := m.activeTable(); t != nil {
		t.StartEdit()
	}
	return m, nil
}

func (m *Model) handleNewEntry() (tea.Model, tea.Cmd) {
	t := m.activeTable()
	if m.focusedPanel != PanelContent || t == nil {
		return m, nil
	}

	created, err := m.ledger.Add(ledger.Entry{Kind: t.Kind, Date: m.newEntryDate()})
	m.setErr(err)

	m.refreshTables()
	t.SelectByID(created.ID)
	t.StartEdit()
	return m, nil
}

func (m *Model) handleDeleteEntry() (tea.Model, tea.Cmd) {
	t := m.activeTable()
	if m.focusedPanel != PanelContent || t == nil {
		return m, nil
	}

	current, ok := t.Current()
	if !ok {
		return m, nil
	}

	m.setErr(m.ledger.Remove(current.ID))
	m.refreshTables()
	return m, nil
}

func (m *Model) handleTableEdit(msg tea.KeyMsg, t *components.EntryTable) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		t.CancelEdit()
	case "tab", "enter":
		t.NextColumn()
		if !t.Editing {
			m.commitEdit(t)
		}
	case "backspace":
		t.Backspace()
	default:
		if len(msg.String()) == 1 {
			t.TypeChar(rune(msg.String()[0]))
		}
	}
	return m, nil
}

func (m *Model) commitEdit(t *components.EntryTable) {
	edited, ok := t.Current()
	if !ok {
		return
	}

	m.setErr(m.ledger.Update(edited))
	m.refreshTables()
	t.SelectByID(edited.ID)
}

func (m *Model) newEntryDate() time.Time {
	now := m.ledger.Now()
	if utils.FirstOfMonth(now).Equal(m.month) {
		return now
	}
	return m.month
}
