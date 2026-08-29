package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gabriel-luiz/truf/internal/ledger"
	_ "modernc.org/sqlite"
)

const dateLayout = "2006-01-02"

type SQLiteStorage struct {
	db          *sql.DB
	path        string
	hasMonthKey bool
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	if dbPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = filepath.Join(home, dbPath[1:])
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	s := &SQLiteStorage{db: db, path: dbPath}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *SQLiteStorage) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS entries (
		id TEXT PRIMARY KEY,
		date TEXT NOT NULL,
		description TEXT NOT NULL,
		category TEXT NOT NULL,
		amount REAL NOT NULL,
		entry_type TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		cat_type TEXT NOT NULL,
		cat_order INTEGER NOT NULL,
		UNIQUE(name, cat_type)
	);

	CREATE INDEX IF NOT EXISTS idx_entries_type ON entries(entry_type);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	hasMonthKey, err := s.columnExists("entries", "month_key")
	if err != nil {
		return err
	}
	s.hasMonthKey = hasMonthKey

	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	stmt, err := s.db.Prepare("INSERT INTO categories (name, cat_type, cat_order) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range ledger.DefaultCategories() {
		if _, err := stmt.Exec(c.Name, string(c.Kind), c.Order); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLiteStorage) columnExists(table, column string) (bool, error) {
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid, notNull, pk int
		var name, colType string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}

	return false, rows.Err()
}

func (s *SQLiteStorage) Load() (ledger.Snapshot, error) {
	var snapshot ledger.Snapshot

	categories, err := s.loadCategories()
	if err != nil {
		return snapshot, err
	}
	snapshot.Categories = categories

	rows, err := s.db.Query(`
		SELECT id, date, description, category, amount, entry_type
		FROM entries
		ORDER BY date ASC
	`)
	if err != nil {
		return snapshot, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, dateStr, desc, cat, entryType string
		var amount float64

		if err := rows.Scan(&id, &dateStr, &desc, &cat, &amount, &entryType); err != nil {
			return ledger.Snapshot{}, err
		}

		date, err := time.Parse(dateLayout, dateStr)
		if err != nil {
			return ledger.Snapshot{}, fmt.Errorf("entry %s: invalid date %q: %w", id, dateStr, err)
		}

		kind := ledger.Expense
		if entryType == string(ledger.Income) {
			kind = ledger.Income
		}

		snapshot.Entries = append(snapshot.Entries, ledger.Entry{
			ID:          id,
			Date:        date,
			Description: desc,
			Category:    cat,
			Amount:      amount,
			Kind:        kind,
		})
	}

	return snapshot, rows.Err()
}

func (s *SQLiteStorage) loadCategories() ([]ledger.Category, error) {
	rows, err := s.db.Query("SELECT name, cat_type, cat_order FROM categories ORDER BY cat_order")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []ledger.Category
	for rows.Next() {
		var name, catType string
		var order int
		if err := rows.Scan(&name, &catType, &order); err != nil {
			return nil, err
		}
		categories = append(categories, ledger.Category{
			Name:  name,
			Kind:  ledger.Kind(catType),
			Order: order,
		})
	}

	return categories, rows.Err()
}

func (s *SQLiteStorage) Save(snapshot ledger.Snapshot) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM entries"); err != nil {
		return err
	}

	insert := `INSERT INTO entries (id, date, description, category, amount, entry_type) VALUES (?, ?, ?, ?, ?, ?)`
	if s.hasMonthKey {
		insert = `INSERT INTO entries (id, date, description, category, amount, entry_type, month_key) VALUES (?, ?, ?, ?, ?, ?, ?)`
	}

	stmt, err := tx.Prepare(insert)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range snapshot.Entries {
		args := []any{
			e.ID,
			e.Date.Format(dateLayout),
			e.Description,
			e.Category,
			e.Amount,
			string(e.Kind),
		}
		if s.hasMonthKey {
			args = append(args, e.Date.Format("2006-01"))
		}
		if _, err := stmt.Exec(args...); err != nil {
			return err
		}
	}

	if err := s.saveCategories(tx, snapshot.Categories); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLiteStorage) saveCategories(tx *sql.Tx, categories []ledger.Category) error {
	if len(categories) == 0 {
		return nil
	}

	if _, err := tx.Exec("DELETE FROM categories"); err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO categories (name, cat_type, cat_order) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range categories {
		if _, err := stmt.Exec(c.Name, string(c.Kind), c.Order); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
