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
	Amount      int64
	Kind        Kind
}

type Category struct {
	Name  string
	Kind  Kind
	Order int
}

type Summary struct {
	TotalIncome    int64
	TotalExpenses  int64
	Balance        int64
	CategoryTotals map[string]int64
}

type ChartData struct {
	Months   []string
	Income   []int64
	Expenses []int64
	Balance  []int64
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
	if err := validate(e); err != nil {
		return Entry{}, err
	}
	for _, existing := range l.entries {
		if existing.ID == e.ID {
			return Entry{}, fmt.Errorf("duplicate entry id: %s", e.ID)
		}
	}

	l.entries = append(l.entries, e)
	return e, l.persist()
}

func validate(e Entry) error {
	if e.Kind != Income && e.Kind != Expense {
		return fmt.Errorf("unknown kind: %q", e.Kind)
	}
	if e.Amount < 0 {
		return fmt.Errorf("amount must not be negative: %d", e.Amount)
	}
	return nil
}

func (l *Ledger) Update(e Entry) error {
	if err := validate(e); err != nil {
		return err
	}
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
	s := Summary{CategoryTotals: make(map[string]int64)}

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
	if n <= 0 {
		return ChartData{}
	}
	data := ChartData{
		Months:   make([]string, n),
		Income:   make([]int64, n),
		Expenses: make([]int64, n),
		Balance:  make([]int64, n),
	}

	first := utils.AddMonths(utils.FirstOfMonth(endMonth), -(n - 1))
	firstIndex := monthIndex(first)
	for _, e := range l.entries {
		i := monthIndex(e.Date) - firstIndex
		if i < 0 || i >= n {
			continue
		}
		if e.Kind == Income {
			data.Income[i] += e.Amount
		} else {
			data.Expenses[i] += e.Amount
		}
	}

	var running int64
	for i := range n {
		data.Months[i] = utils.FormatMonthYear(utils.AddMonths(first, i))
		running += data.Income[i] - data.Expenses[i]
		data.Balance[i] = running
	}

	return data
}

func monthIndex(t time.Time) int {
	return t.Year()*12 + int(t.Month()) - 1
}

func (l *Ledger) Oldest() (time.Time, bool) {
	var oldest time.Time
	for _, e := range l.entries {
		if oldest.IsZero() || e.Date.Before(oldest) {
			oldest = e.Date
		}
	}
	return oldest, !oldest.IsZero()
}

func (l *Ledger) TotalBalance() int64 {
	var total int64
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
