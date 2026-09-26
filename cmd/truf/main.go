package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/seed"
	"github.com/gabriel-luiz/truf/internal/storage"
	"github.com/gabriel-luiz/truf/internal/ui"
)

const defaultDBPath = "~/.truf/truf.db"

var version = "dev"

const usage = `Usage: truf [flags]

Flags:
  --db <path>   ledger database (default ` + defaultDBPath + `)
  --seed        fill the ledger with sample data and exit
  --version     print the version and exit
  -h, --help    show this help
`

type options struct {
	db      string
	seed    bool
	version bool
}

func parseFlags(args []string, stdout, stderr io.Writer) (opts options, code int, exit bool) {
	fs := flag.NewFlagSet("truf", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {}
	fs.StringVar(&opts.db, "db", defaultDBPath, "")
	fs.BoolVar(&opts.seed, "seed", false, "")
	fs.BoolVar(&opts.version, "version", false, "")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, usage)
			return opts, 0, true
		}
		fmt.Fprint(stderr, usage)
		return opts, 2, true
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected argument: %s\n", fs.Arg(0))
		fmt.Fprint(stderr, usage)
		return opts, 2, true
	}
	if opts.db == "" {
		fmt.Fprintln(stderr, "--db needs a path")
		fmt.Fprint(stderr, usage)
		return opts, 2, true
	}
	if opts.version {
		fmt.Fprintf(stdout, "truf %s\n", version)
		return opts, 0, true
	}
	return opts, 0, false
}

func main() {
	opts, code, exit := parseFlags(os.Args[1:], os.Stdout, os.Stderr)
	if exit {
		os.Exit(code)
	}

	store, err := storage.NewSQLiteStorage(opts.db)
	if err != nil {
		fail("Error initializing storage: %v", err)
	}
	defer store.Close()

	book, err := ledger.New(store, time.Now)
	if err != nil {
		fail("Error loading data: %v", err)
	}

	if opts.seed {
		if err := seed.Fake(book); err != nil {
			fail("Error seeding data: %v", err)
		}
		fmt.Printf("Sample data seeded into %s. Run truf to view it.\n", opts.db)
		return
	}

	model := ui.NewModel(book)
	model.SetStorageLabel(opts.db)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fail("Error running program: %v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
