package datastore

import (
	"context"

	"github.com/dknr/caos/store"
)

// NewInMemoryDatastore returns an in-memory implementation of store.DataStore.
// This is primarily useful for testing.
func NewInMemoryDatastore() store.DataStore {
	return NewDataStore(NewMemoryStorage())
}

// memoryStorage implements Storage using an in-memory map.
type memoryStorage struct {
	data map[string][]byte
}

// NewMemoryStorage returns an in-memory implementation of Storage.
// This is primarily useful for testing.
func NewMemoryStorage() Storage {
	return &memoryStorage{
		data: make(map[string][]byte),
	}
}

func (m *memoryStorage) PutData(ctx context.Context, key string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Store the data
	m.data[key] = data
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (m *memoryStorage) GetData(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if data, ok := m.data[key]; ok {
		return data, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, store.ErrNotFound
}

func (m *memoryStorage) HasData(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	_, ok := m.data[key]
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return ok, nil
}

func (m *memoryStorage) DeleteData(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(m.data, key)
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}