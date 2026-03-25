package datastore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/dknr/caos/store"
)

// NewInMemoryDatastore returns an in-memory implementation of store.DataStore.
// This is primarily useful for testing.
func NewInMemoryDatastore() store.DataStore {
	return &inMemoryDatastore{
		data: make(map[string][]byte),
	}
}

type inMemoryDatastore struct {
	data map[string][]byte
}

func (m *inMemoryDatastore) Put(ctx context.Context, r io.Reader) (string, int64, error) {
	// Read all data from the reader
	data, err := io.ReadAll(r)
	if err != nil {
		return "", 0, err
	}
	// Compute SHA-256 hash
	hash := sha256.Sum256(data)
	addr := hex.EncodeToString(hash[:])
	// Store the data
	m.data[addr] = data
	return addr, int64(len(data)), nil
}

func (m *inMemoryDatastore) Get(ctx context.Context, addr string) (io.ReadCloser, error) {
	if data, ok := m.data[addr]; ok {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	return nil, store.ErrNotFound
}

func (m *inMemoryDatastore) Has(ctx context.Context, addr string) (bool, error) {
	_, ok := m.data[addr]
	return ok, nil
}

func (m *inMemoryDatastore) Delete(ctx context.Context, addr string) error {
	delete(m.data, addr)
	return nil
}