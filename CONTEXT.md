# TRUF CLI — domain vocabulary

TRUF is a terminal app for tracking personal income and expenses month by month.

## Words

**Entry** — a single income or expense. Carries `ID`, `Date`, `Description`, `Category`, `Amount` and `Kind`. An entry's month is *derived from its `Date`*; there is no separate month key. Editing a date into another month moves the entry to that month.

**Kind** — `Income` or `Expense`. One field on `Entry` instead of two parallel collections, so there is a single code path per operation.

**Category** — `Name`, `Kind`, `Order`. Categories are global, not per-month. They are a convenience for defaulting and display: an entry's `Category` is free text and is never validated against the list.

**Snapshot** — the whole persisted state as a flat value: `{Entries, Categories}`. It is what crosses the storage seam, in both directions.

**Storage adapter** — anything satisfying `ledger.Storage`: `Load() (Snapshot, error)` and `Save(Snapshot) error`. Two exist: `storage.SQLiteStorage` (the real one; `Close` is a concrete method on it, called from `main`) and `storage.MemoryStorage` (tests). Adding a third means implementing two methods.

**Clock** — `func() time.Time`, injected into the Ledger. Everything time-dependent (today's date on a new entry, the current month) goes through it, so tests are deterministic.

**Ledger** — the single owner of every Entry. It is constructed with a Storage adapter and a Clock, loads the snapshot up front, and is the only thing that mutates entries:

- `Entries(month, kind)`, `Categories(kind)` — read side
- `Add(entry) (Entry, error)`, `Update(entry) error`, `Remove(id) error` — mutation
- `Summary(month)`, `ChartSeries(endMonth, n)`, `TotalBalance()` — derived values
- `Now()` — the injected clock

Every mutation persists the full snapshot synchronously and returns the storage error. In-memory state is updated even when the save fails, so the UI keeps showing what the user typed while the status bar shows the error.

The Ledger holds no selected month. Month is an explicit argument on every call; the UI model owns the selection (currently always the clock's current month).

## Layout

- `internal/ledger` — Ledger, Entry, Kind, Category, Snapshot, Storage interface, Clock, derived values.
- `internal/storage` — the SQLite and memory adapters.
- `internal/seed` — `--seed` sample data, inserted through `Ledger.Add`.
- `internal/ui` — Bubble Tea model/update/view. `EntryTable` is a *view* over `Ledger.Entries`: it holds a copy for cursor and inline editing, and the model writes every new/delete/commit back through the Ledger and re-reads.
- `pkg/utils` — date and currency formatting/parsing.

## Conventions

Go code here carries no comments — names are expected to do the work.
