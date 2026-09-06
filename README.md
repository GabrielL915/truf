# truf

Personal finance tracker in the terminal — income, expenses and a balance chart, month by month.

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

## Keys

| Key | Action |
| --- | --- |
| `Tab` | switch between menu and content |
| `↑` `↓` / `k` `j` | navigate |
| `Enter` | open a view, or start editing a row |
| `[` `]` / `h` `l` | previous / next month (Income and Expenses) |
| `n` | new entry |
| `d` | delete entry |
| `Esc` | back to Overview |
| `PgUp` / `PgDn` | widen / narrow the chart range |
| `q` | quit |

While editing a row, `Tab`/`Enter` moves to the next column and `Esc` cancels.

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
internal/ui       Bubble Tea model, update, view, components
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
