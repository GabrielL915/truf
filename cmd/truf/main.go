package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/seed"
	"github.com/gabriel-luiz/truf/internal/storage"
	"github.com/gabriel-luiz/truf/internal/ui"
)

const dbPath = "~/.truf/truf.db"

func main() {
	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		fail("Error initializing storage: %v", err)
	}
	defer store.Close()

	book, err := ledger.New(store, time.Now)
	if err != nil {
		fail("Error loading data: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "--seed" {
		if err := seed.Fake(book); err != nil {
			fail("Error seeding data: %v", err)
		}
		fmt.Println("Fake data seeded! Run ./truf.exe to view it.")
		return
	}

	model := ui.NewModel(book)
	model.SetStorageLabel(dbPath)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fail("Error running program: %v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
