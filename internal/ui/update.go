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
		m.resize(msg.Width, msg.Height)
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.setErr(nil)

	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.layout.tooSmall {
		if msg.String() == "q" {
			return m, tea.Quit
		}
		return m, nil
	}

	if t := m.activeTable(); t != nil && t.Editing {
		return m.handleTableEdit(msg, t)
	}

	switch msg.String() {
	case "q":
		if m.currentView == ViewOverview {
			return m, tea.Quit
		}

	case "1", "2", "3", "4", "5":
		m.setView(ViewType(msg.String()[0] - '1'))

	case "tab":
		m.cycleView(1)

	case "shift+tab":
		m.cycleView(-1)

	case "esc":
		if m.currentView != ViewOverview {
			m.setView(ViewOverview)
		}

	case "up", "k":
		if t := m.activeTable(); t != nil {
			t.Up()
		}

	case "down", "j":
		if t := m.activeTable(); t != nil {
			t.Down()
		}

	case "pgup":
		if m.currentView == ViewOverview && m.chartMonths < 24 {
			m.chartMonths += 3
			m.refreshOverview()
		}

	case "pgdown":
		if m.currentView == ViewOverview && m.chartMonths > 3 {
			m.chartMonths -= 3
			m.refreshOverview()
		}

	case "[", "h":
		m.shiftMonth(-1)

	case "]", "l":
		m.shiftMonth(1)

	case "enter":
		if t := m.activeTable(); t != nil {
			t.StartEdit()
		}

	case "n":
		m.handleNewEntry()

	case "d":
		m.handleDeleteEntry()
	}

	return m, nil
}

func (m *Model) handleNewEntry() {
	t := m.activeTable()
	if t == nil {
		return
	}

	created, err := m.ledger.Add(ledger.Entry{Kind: t.Kind, Date: m.newEntryDate()})
	m.setErr(err)

	m.refreshTables()
	t.SelectByID(created.ID)
	t.StartEdit()
}

func (m *Model) handleDeleteEntry() {
	t := m.activeTable()
	if t == nil {
		return
	}

	current, ok := t.Current()
	if !ok {
		return
	}

	m.setErr(m.ledger.Remove(current.ID))
	m.refreshTables()
}

func (m *Model) handleTableEdit(msg tea.KeyMsg, t *components.EntryTable) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		t.CancelEdit()
	case "tab", "enter":
		if err := t.NextColumn(); err != nil {
			m.setErr(err)
			return m, nil
		}
		if !t.Editing {
			m.commitEdit(t)
		}
	case "backspace":
		t.Backspace()
	default:
		switch msg.Type {
		case tea.KeySpace:
			t.TypeChar(' ')
		case tea.KeyRunes:
			for _, r := range msg.Runes {
				t.TypeChar(r)
			}
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
