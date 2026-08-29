package storage

import (
	"github.com/gabriel-luiz/truf/internal/ledger"
)

type MemoryStorage struct {
	Snapshot ledger.Snapshot
	Saves    []ledger.Snapshot
	LoadErr  error
	SaveErr  error
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (m *MemoryStorage) Load() (ledger.Snapshot, error) {
	if m.LoadErr != nil {
		return ledger.Snapshot{}, m.LoadErr
	}
	return m.Snapshot, nil
}

func (m *MemoryStorage) Save(s ledger.Snapshot) error {
	m.Saves = append(m.Saves, s)
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.Snapshot = s
	return nil
}

func (m *MemoryStorage) LastSave() (ledger.Snapshot, bool) {
	if len(m.Saves) == 0 {
		return ledger.Snapshot{}, false
	}
	return m.Saves[len(m.Saves)-1], true
}
