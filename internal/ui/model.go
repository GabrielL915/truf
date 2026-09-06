package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/ui/components"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

type Panel int

const (
	PanelMenu Panel = iota
	PanelContent
)

type ViewType int

const (
	ViewOverview ViewType = iota
	ViewIncome
	ViewExpenses
	ViewCategories
	ViewSettings
)

type Model struct {
	ledger *ledger.Ledger
	month  time.Time

	menu         *components.Menu
	chart        *components.Chart
	statusBar    *components.StatusBar
	incomeTable  *components.EntryTable
	expenseTable *components.EntryTable

	focusedPanel Panel
	currentView  ViewType
	chartMonths  int

	width  int
	height int

	err error
}

func NewModel(book *ledger.Ledger) *Model {
	m := &Model{
		ledger:       book,
		month:        utils.FirstOfMonth(book.Now()),
		menu:         components.NewMenu(),
		chart:        components.NewChart(),
		statusBar:    components.NewStatusBar(),
		incomeTable:  components.NewEntryTable("Income", ledger.Income),
		expenseTable: components.NewEntryTable("Expenses", ledger.Expense),
		focusedPanel: PanelMenu,
		currentView:  ViewOverview,
		chartMonths:  6,
	}

	m.refreshChart()
	m.refreshTables()
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) table(kind ledger.Kind) *components.EntryTable {
	if kind == ledger.Income {
		return m.incomeTable
	}
	return m.expenseTable
}

func (m *Model) activeTable() *components.EntryTable {
	switch m.currentView {
	case ViewIncome:
		return m.incomeTable
	case ViewExpenses:
		return m.expenseTable
	}
	return nil
}

func (m *Model) refreshChart() {
	data := m.ledger.ChartSeries(m.month, m.chartMonths)
	m.chart.SetData(data)

	if len(data.Months) > 0 {
		m.statusBar.SetTimeRange(data.Months[0], data.Months[len(data.Months)-1])
	}
	m.statusBar.SetBalance(m.ledger.TotalBalance())
}

func (m *Model) refreshTables() {
	for _, kind := range []ledger.Kind{ledger.Income, ledger.Expense} {
		t := m.table(kind)
		t.SetMonth(m.month)
		t.SetEntries(m.ledger.Entries(m.month, kind))
		t.SetCategories(categoryNames(m.ledger.Categories(kind)))
	}
}

func (m *Model) shiftMonth(delta int) {
	m.month = utils.AddMonths(m.month, delta)
	m.incomeTable.ResetCursor()
	m.expenseTable.ResetCursor()
	m.refreshTables()
	m.refreshChart()
}

func (m *Model) setErr(err error) {
	m.err = err
	m.statusBar.SetError(err)
}

func (m *Model) updateFocus() {
	m.menu.Focused = m.focusedPanel == PanelMenu
	contentFocused := m.focusedPanel == PanelContent
	m.chart.Focused = contentFocused && m.currentView == ViewOverview
	m.incomeTable.Focused = contentFocused && m.currentView == ViewIncome
	m.expenseTable.Focused = contentFocused && m.currentView == ViewExpenses
}

func (m *Model) updateSizes() {
	leftWidth := m.width * 25 / 100
	rightWidth := m.width - leftWidth
	mainHeight := m.height - 3

	m.menu.SetSize(leftWidth, mainHeight)
	m.chart.SetSize(rightWidth, mainHeight)
	m.incomeTable.SetSize(rightWidth, mainHeight)
	m.expenseTable.SetSize(rightWidth, mainHeight)
	m.statusBar.SetWidth(m.width)
}

func categoryNames(categories []ledger.Category) []string {
	names := make([]string, 0, len(categories))
	for _, c := range categories {
		names = append(names, c.Name)
	}
	return names
}
