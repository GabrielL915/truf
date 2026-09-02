package seed

import (
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

type entry struct {
	description string
	category    string
	amount      int64
}

type month struct {
	income   []entry
	expenses []entry
}

func Fake(l *ledger.Ledger) error {
	months := []month{
		{
			income: []entry{
				{"Salary", "Salary", 4200},
				{"Freelance project", "Freelance", 800},
			},
			expenses: []entry{
				{"Rent", "Housing", 1200},
				{"Groceries", "Food", 380},
				{"Bus pass", "Transportation", 95},
				{"Gym membership", "Healthcare", 45},
				{"Netflix", "Entertainment", 15},
				{"Internet", "Utilities", 60},
			},
		},
		{
			income: []entry{
				{"Salary", "Salary", 4200},
			},
			expenses: []entry{
				{"Rent", "Housing", 1200},
				{"Groceries", "Food", 420},
				{"Uber", "Transportation", 75},
				{"Doctor visit", "Healthcare", 120},
				{"Online course", "Education", 99},
				{"Electricity", "Utilities", 85},
			},
		},
		{
			income: []entry{
				{"Salary", "Salary", 4200},
				{"Investments return", "Investments", 350},
			},
			expenses: []entry{
				{"Rent", "Housing", 1200},
				{"Groceries", "Food", 395},
				{"Bus pass", "Transportation", 95},
				{"Restaurant", "Food", 180},
				{"Spotify", "Entertainment", 10},
				{"Phone bill", "Utilities", 50},
				{"Clothes", "Other Expenses", 230},
			},
		},
		{
			income: []entry{
				{"Salary", "Salary", 4500},
				{"Freelance project", "Freelance", 1200},
			},
			expenses: []entry{
				{"Rent", "Housing", 1200},
				{"Groceries", "Food", 360},
				{"Car fuel", "Transportation", 110},
				{"Pharmacy", "Healthcare", 65},
				{"Books", "Education", 45},
				{"Cinema", "Entertainment", 35},
				{"Water bill", "Utilities", 40},
			},
		},
		{
			income: []entry{
				{"Salary", "Salary", 4500},
			},
			expenses: []entry{
				{"Rent", "Housing", 1200},
				{"Groceries", "Food", 410},
				{"Bus pass", "Transportation", 95},
				{"Dentist", "Healthcare", 200},
				{"Internet", "Utilities", 60},
				{"Games", "Entertainment", 60},
			},
		},
		{
			income: []entry{
				{"Salary", "Salary", 4500},
				{"Investments return", "Investments", 520},
			},
			expenses: []entry{
				{"Rent", "Housing", 1200},
				{"Groceries", "Food", 220},
				{"Uber", "Transportation", 45},
				{"Phone bill", "Utilities", 50},
			},
		},
	}

	current := utils.FirstOfMonth(l.Now())

	for i, m := range months {
		start := utils.AddMonths(current, -(len(months) - 1 - i))

		for _, e := range m.income {
			if err := add(l, e, start.AddDate(0, 0, 1), ledger.Income); err != nil {
				return err
			}
		}

		for _, e := range m.expenses {
			if err := add(l, e, start.AddDate(0, 0, 5), ledger.Expense); err != nil {
				return err
			}
		}
	}

	return nil
}

func add(l *ledger.Ledger, e entry, date time.Time, kind ledger.Kind) error {
	_, err := l.Add(ledger.Entry{
		Date:        date,
		Description: e.description,
		Category:    e.category,
		Amount:      e.amount * 100,
		Kind:        kind,
	})
	return err
}
