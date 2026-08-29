package ledger

import (
	"fmt"
	"sort"
	"time"

	"github.com/gabriel-luiz/truf/pkg/utils"
	"github.com/google/uuid"
)

type Kind string

const (
	Income  Kind = "income"
	Expense Kind = "expense"
)

type Entry struct {
	ID          string
	Date        time.Time
	Description string
	Category    string
	Amount      float64
	Kind        Kind
}

type Category struct {
	Name  string
	Kind  Kind
	Order int
}

type Summary struct {
	TotalIncome    float64
	TotalExpenses  float64
	Balance        float64
	CategoryTotals map[string]float64
}

type ChartData struct {
	Months   []string
	Income   []float64
	Expenses []float64
	Balance  []float64
}

type Snapshot struct {
	Entries    []Entry
	Categories []Category
}

type Storage interface {
	Load() (Snapshot, error)
	Save(Snapshot) error
}

type Clock func() time.Time

type Ledger struct {
	store      Storage
	clock      Clock
	entries    []Entry
	categories []Category
}

func New(store Storage, clock Clock) (*Ledger, error) {
	snapshot, err := store.Load()
	if err != nil {
		return nil, err
	}

	categories := snapshot.Categories
	if len(categories) == 0 {
		categories = DefaultCategories()
	}

	return &Ledger{
		store:      store,
		clock:      clock,
		entries:    snapshot.Entries,
		categories: categories,
	}, nil
}

func DefaultCategories() []Category {
	return []Category{
		{Name: "Salary", Kind: Income, Order: 1},
		{Name: "Freelance", Kind: Income, Order: 2},
		{Name: "Investments", Kind: Income, Order: 3},
		{Name: "Other Income", Kind: Income, Order: 4},

		{Name: "Housing", Kind: Expense, Order: 1},
		{Name: "Food", Kind: Expense, Order: 2},
		{Name: "Transportation", Kind: Expense, Order: 3},
		{Name: "Healthcare", Kind: Expense, Order: 4},
		{Name: "Education", Kind: Expense, Order: 5},
		{Name: "Entertainment", Kind: Expense, Order: 6},
		{Name: "Utilities", Kind: Expense, Order: 7},
		{Name: "Other Expenses", Kind: Expense, Order: 8},
	}
}

func (l *Ledger) Now() time.Time {
	return l.clock()
}

func (l *Ledger) Entries(month time.Time, kind Kind) []Entry {
	var out []Entry
	for _, e := range l.entries {
		if e.Kind == kind && sameMonth(e.Date, month) {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date.Before(out[j].Date) })
	return out
}

func (l *Ledger) Add(e Entry) (Entry, error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	if e.Date.IsZero() {
		e.Date = l.clock()
	}
	if e.Category == "" {
		if cats := l.Categories(e.Kind); len(cats) > 0 {
			e.Category = cats[0].Name
		}
	}

	l.entries = append(l.entries, e)
	return e, l.persist()
}

func (l *Ledger) Update(e Entry) error {
	for i, existing := range l.entries {
		if existing.ID == e.ID {
			l.entries[i] = e
			return l.persist()
		}
	}
	return fmt.Errorf("entry not found: %s", e.ID)
}

func (l *Ledger) Remove(id string) error {
	for i, e := range l.entries {
		if e.ID == id {
			l.entries = append(l.entries[:i], l.entries[i+1:]...)
			return l.persist()
		}
	}
	return fmt.Errorf("entry not found: %s", id)
}

func (l *Ledger) Categories(kind Kind) []Category {
	var out []Category
	for _, c := range l.categories {
		if c.Kind == kind {
			out = append(out, c)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out
}

func (l *Ledger) Summary(month time.Time) Summary {
	s := Summary{CategoryTotals: make(map[string]float64)}

	for _, e := range l.entries {
		if !sameMonth(e.Date, month) {
			continue
		}
		if e.Kind == Income {
			s.TotalIncome += e.Amount
		} else {
			s.TotalExpenses += e.Amount
		}
		s.CategoryTotals[e.Category] += e.Amount
	}

	s.Balance = s.TotalIncome - s.TotalExpenses
	return s
}

func (l *Ledger) ChartSeries(endMonth time.Time, n int) ChartData {
	data := ChartData{
		Months:   make([]string, 0, n),
		Income:   make([]float64, 0, n),
		Expenses: make([]float64, 0, n),
		Balance:  make([]float64, 0, n),
	}
	if n <= 0 {
		return data
	}

	running := 0.0
	for i := n - 1; i >= 0; i-- {
		month := utils.AddMonths(utils.FirstOfMonth(endMonth), -i)
		summary := l.Summary(month)
		running += summary.Balance

		data.Months = append(data.Months, utils.FormatMonthYear(month))
		data.Income = append(data.Income, summary.TotalIncome)
		data.Expenses = append(data.Expenses, summary.TotalExpenses)
		data.Balance = append(data.Balance, running)
	}

	return data
}

func (l *Ledger) TotalBalance() float64 {
	var total float64
	for _, e := range l.entries {
		if e.Kind == Income {
			total += e.Amount
		} else {
			total -= e.Amount
		}
	}
	return total
}

func (l *Ledger) persist() error {
	entries := make([]Entry, len(l.entries))
	copy(entries, l.entries)
	categories := make([]Category, len(l.categories))
	copy(categories, l.categories)
	return l.store.Save(Snapshot{Entries: entries, Categories: categories})
}

func sameMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}
