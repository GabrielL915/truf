# Ledger module: single owner of Entry mutation

Status: ready-for-agent

## Problem Statement

When I add, edit or delete an income/expense in TRUF, the change is applied in two places at once (the table on screen and the in-memory month data), and the app relies on a later "sync" step to make them agree again. This has already produced latent bugs: a new entry briefly exists twice, deletion only works by accident, and every edit rewrites the whole database. None of this behaviour can be tested without driving the full terminal UI, so regressions go unnoticed.

## Solution

Introduce a **Ledger**: one module that owns every Entry, applies add/update/remove, derives summaries and chart series, and persists after each change. The on-screen table becomes a view over Ledger data instead of a second owner. The Ledger takes a **Clock** and a **Storage adapter** at construction, so it can be exercised in tests with a fixed date and an in-memory store.

## User Stories

1. As a TRUF user, I want a new entry to appear exactly once, so that my totals are right.
2. As a TRUF user, I want deleting an entry to remove it reliably, so that it doesn't come back after restart.
3. As a TRUF user, I want each edit saved immediately, so that quitting or crashing never loses work.
4. As a TRUF user, I want to see an error in the status bar if a save fails, so that I know my data isn't persisted.
5. As a TRUF user, I want a new entry to default to today's date and the first category of its kind, so that I type less.
6. As a TRUF user, I want editing an entry's date into another month to move it to that month, so that entries always live where their date says.
7. As a TRUF user, I want free-text categories accepted, so that the editor never blocks me.
8. As a TRUF user, I want the overview chart and balance to reflect my edits as soon as I return to Overview.
9. As a TRUF user, I want `--seed` to keep working, so that I can demo the app with sample data.
10. As a developer, I want to add/update/remove entries through one interface, so that ownership is unambiguous.
11. As a developer, I want the Ledger testable with an in-memory store and a fixed clock, so that tests are fast and deterministic.
12. As a developer, I want Entry to carry its Kind (income/expense), so that I don't maintain parallel code paths per slice.
13. As a developer, I want month membership derived from an entry's date, so that there is no separate month key to keep in sync.
14. As a developer, I want the Storage seam reduced to load/save of a flat snapshot, so that a second adapter is trivial.
15. As a developer, I want categories stored once, globally, so that the model doesn't lie about them being per-month.
16. As a developer, I want storage parse errors surfaced rather than swallowed, so that corrupt rows are visible.
17. As a developer, I want the read side (summary, chart series, total balance) on the Ledger, so that all consumers share one implementation.
18. As a developer, I want a CONTEXT.md naming Ledger, Entry, Kind, Snapshot, Storage adapter and Clock, so that future work uses the same words.
19. As a developer, I want the Go code free of comments, so that names carry the meaning (per project owner's preference).

## Implementation Decisions

- New package `ledger` absorbs the current `models` package (types + calculations). `models` is deleted.
- Types: `Entry{ID, Date, Description, Category, Amount, Kind}`; `Kind` with values Income and Expense; `Category{Name, Kind, Order}`; `Summary`; `ChartData`; `Snapshot{Entries, Categories}`.
- Constructor takes a Storage adapter and a `Clock` (`func() time.Time`); it loads the snapshot on construction and returns an error if load fails.
- Interface: `Entries(month, kind)`, `Add(entry)` (stamps a UUID; stamps Date from Clock when zero; defaults Category to the first of its kind when empty), `Update(entry)`, `Remove(id)`, `Categories(kind)`, `Summary(month)`, `ChartSeries(endMonth, n)`, `TotalBalance()`.
- Every mutation persists synchronously via `Save(Snapshot)` and returns the storage error. In-memory state is still updated on save failure.
- Month addressing is explicit on every call; the Ledger holds no "selected month" state. The UI model owns the selected month (currently always the Clock's current month).
- Storage interface shrinks to `Load() (Snapshot, error)` and `Save(Snapshot) error`. `GetMonths` is removed. `Close` remains a concrete method on the SQLite adapter, called from main.
- SQLite adapter keeps delete-and-reinsert internally; the `entry_type` column maps to Kind; the `month_key` column is dropped from the schema (new table shape; no migration—existing DB is recreated only if schema mismatch is detected, otherwise `month_key` is simply ignored on read and written as derived from date to stay compatible).
- A memory adapter is added for tests.
- Seed data is inserted through `Add`, not by building a snapshot.
- `EntryTable` keeps its current key-handling API but is fed a copy of the entries and stops being a source of truth: the UI model reacts to new/delete/commit by calling the Ledger and re-reading `Entries`.
- The UI model's error field is rendered in the status bar.
- All comments are stripped from Go files under `truf-cli`.
- `truf-cli/CONTEXT.md` is created.

## Testing Decisions

- Good tests exercise the Ledger only through its interface with the memory adapter and a fixed Clock; they assert on returned values and on what the adapter has been asked to save, never on internal fields.
- Ledger tests: Add stamps ID and Date and persists; Add defaults category; Update that changes month moves the entry between months; Remove by ID persists and unknown ID errors; Summary totals per month; ChartSeries over n months with running balance; empty ledger yields n zero-valued months ending at the clock's month; save failure is returned to the caller.
- SQLite adapter test: Save then Load on a temp-file DB round-trips entries and categories.
- No prior art: the repo has no tests. Standard `testing` package, no assertion library.

## Out of Scope

- Collapsing EntryTable key handling into `HandleKey/Outcome` (architecture review candidate 2).
- Month navigation UI (candidate 4 beyond the Clock injection).
- Category validation / Categories panel, Settings panel.
- Layout function (candidate 5).
- The `Fin/` project.

## Further Notes

Derived from the architecture review of 2026-08-29 and a three-round design interview. If a tracker is configured later, this file can be moved there with the `ready-for-agent` label.
