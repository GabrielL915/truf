package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/ui/components"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

type ViewType int

const (
	ViewOverview ViewType = iota
	ViewIncome
	ViewExpenses
	ViewCategories
	ViewSettings
	viewCount
)

var viewLabels = []string{"Overview", "Income", "Expenses", "Categories", "Settings"}

type Model struct {
	ledger *ledger.Ledger
	month  time.Time

	tabBar       *components.TabBar
	summary      *components.Summary
	chart        *components.Chart
	helpBar      *components.HelpBar
	incomeTable  *components.EntryTable
	expenseTable *components.EntryTable

	currentView ViewType
	chartMonths int
	empty       bool

	layout layout

	err error
}

func NewModel(book *ledger.Ledger) *Model {
	m := &Model{
		ledger:       book,
		month:        utils.FirstOfMonth(book.Now()),
		tabBar:       components.NewTabBar(viewLabels),
		summary:      components.NewSummary(),
		chart:        components.NewChart(),
		helpBar:      components.NewHelpBar(),
		incomeTable:  components.NewEntryTable("Income", ledger.Income),
		expenseTable: components.NewEntryTable("Expenses", ledger.Expense),
		currentView:  ViewOverview,
		chartMonths:  6,
	}

	m.tabBar.SetMonth(m.month)
	m.refreshOverview()
	m.refreshTables()
	return m
}

func (m *Model) SetStorageLabel(label string) {
	m.helpBar.StorageLabel = label
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

func (m *Model) setView(v ViewType) {
	if v < 0 || v >= viewCount {
		return
	}
	m.currentView = v
	m.tabBar.SetActive(int(v))
	switch v {
	case ViewOverview:
		m.refreshOverview()
	case ViewIncome, ViewExpenses:
		m.refreshTables()
	}
}

func (m *Model) cycleView(delta int) {
	n := int(viewCount)
	m.setView(ViewType((int(m.currentView) + delta + n) % n))
}

func (m *Model) refreshOverview() {
	m.chart.SetData(m.ledger.ChartSeries(m.month, m.chartMonths))
	m.chart.SetActiveMonth(utils.FormatMonthYear(m.month))

	current := m.ledger.Summary(m.month)
	previous := m.ledger.Summary(utils.AddMonths(m.month, -1))
	oldest, ok := m.ledger.Oldest()
	m.empty = !ok

	s := m.summary
	s.Month = m.month
	s.Income = current.TotalIncome
	s.Expenses = current.TotalExpenses
	s.Net = current.Balance
	s.NetChange = current.Balance - previous.Balance
	s.IncomeCount = len(m.ledger.Entries(m.month, ledger.Income))
	s.ExpenseCount = len(m.ledger.Entries(m.month, ledger.Expense))
	s.Balance = m.ledger.TotalBalance()
	s.Since = oldest
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
	m.tabBar.SetMonth(m.month)
	m.incomeTable.ResetCursor()
	m.expenseTable.ResetCursor()
	m.refreshTables()
	m.refreshOverview()
}

func (m *Model) setErr(err error) {
	m.err = err
	m.helpBar.SetError(err)
}

func (m *Model) resize(width, height int) {
	m.layout = computeLayout(width, height)
	l := m.layout

	m.tabBar.SetWidth(l.width)
	m.summary.SetWidth(l.width)
	m.chart.SetSize(l.width, l.contentHeight-components.SummaryHeight-1)
	m.incomeTable.SetSize(l.width, l.contentHeight)
	m.expenseTable.SetSize(l.width, l.contentHeight)
	m.helpBar.SetWidth(l.width)
}

func categoryNames(categories []ledger.Category) []string {
	names := make([]string, 0, len(categories))
	for _, c := range categories {
		names = append(names, c.Name)
	}
	return names
}
