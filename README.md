<p align="center"><img src="assets/truf.png" width="160" alt="TRUF"></p>

# truf

Personal finance tracker in the terminal — income, expenses and a balance chart, month by month.

Three full-width bands: a tab bar with the five views and the active month, the content, and a help
bar with the keys that work right now. The Overview opens with the month summary (Income, Expenses,
`Net · <month>`, `Balance · all time`) above the balance chart; Income and Expenses are borderless
tables with a pinned footer showing the category distribution. Nothing paints a background except
the two bars, the active tab and the cursor row, so the app inherits your terminal theme.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and SQLite.

## Run

```sh
go run ./cmd/truf
```

Data lives in `~/.truf/truf.db`.

To try it with sample data:

```sh
go run ./cmd/truf --seed
```

Flags:

| Flag | Action |
| --- | --- |
| `--db <path>` | use another ledger database (default `~/.truf/truf.db`) |
| `--seed` | fill the ledger with sample data and exit |
| `--version` | print the version and exit |
| `-h`, `--help` | show usage |

Unknown flags print the usage and exit with status 2.

## Keys

| Key | Where | Action |
| --- | --- | --- |
| `1` `2` `3` `4` `5` | everywhere | jump to Overview, Income, Expenses, Categories, Settings |
| `Tab` / `Shift+Tab` | everywhere | next / previous view |
| `Esc` | any view | back to Overview; while editing, cancel the edit |
| `[` `]` / `h` `l` | everywhere | previous / next month |
| `↑` `↓` / `k` `j` | tables | move the cursor |
| `Enter` | tables | edit the row; while editing, next column |
| `n` / `d` | tables | new / delete entry |
| `PgUp` / `PgDn` | Overview | widen / narrow the chart range |
| `q` | Overview | quit |
| `Ctrl+C` | everywhere | quit |

The help bar at the bottom lists the keys for the current view and shows the database path on the
right. A failed save replaces it with the error until the next key. TRUF needs at least 80×24.

## Layout

The `ledger` package owns every entry: it applies add/update/remove, derives summaries and chart
series, and persists after each change. The table on screen is a view over it, never a second owner.
See [CONTEXT.md](CONTEXT.md) for the vocabulary and [docs/specs/ledger.md](docs/specs/ledger.md) for
the spec behind it.

```
cmd/truf          entrypoint
internal/ledger   Entry, Kind, Snapshot, Ledger, Clock, Storage interface
internal/storage  SQLite and in-memory adapters
internal/seed     --seed sample data
internal/ui       Bubble Tea model, update, view, layout, components, styles, marks
assets            the logo; internal/ui/marks renders it as braille via go generate
pkg/utils         date and currency helpers
```

## Tests

```sh
go test ./...
```

## Knowledge base

Decisions, conventions, gotchas and open issues are indexed in `docs/knowledge/truf.jsonl`
(one JSON object per line). Query it before digging through code:

```sh
scripts/kb.sh search sqlite save     # one line per hit
scripts/kb.sh show decision.int-cents
scripts/kb.sh add --id x.y --type decision --topic money --summary "..." --refs pkg/utils/currency.go
scripts/kb.sh help
```

Requires `jq`. CI runs `scripts/kb.sh check` (valid JSON, unique ids, every `refs` path exists).
Specs are drafted locally in `docs/specs/` (gitignored) and distilled into a `spec.*` entry.
