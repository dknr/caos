package metastore

import (
	"context"
	"sync"

	"github.com/dknr/caos/store"
)

// NewInMemoryMetaStore returns an in-memory implementation of store.MetaStore.
func NewInMemoryMetaStore() store.MetaStore {
	return &inMemoryMetaStore{
		objs: make(map[string]objMetadata),
		mu:   sync.RWMutex{},
	}
}

type objMetadata struct {
	size int64
	typ  string
}

type inMemoryMetaStore struct {
	objs map[string]objMetadata
	mu   sync.RWMutex
}

func (m *inMemoryMetaStore) AddObject(ctx context.Context, addr string, size int64, typ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.objs[addr]; exists {
		return store.ErrNotFound // Using ErrNotFound to indicate "already exists" for now - we might want a specific error
	}
	m.objs[addr] = objMetadata{size: size, typ: typ}
	return nil
}

func (m *inMemoryMetaStore) SetType(ctx context.Context, addr string, typ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if obj, exists := m.objs[addr]; exists {
		obj.typ = typ
		m.objs[addr] = obj
		return nil
	}
	return store.ErrNotFound
}

func (m *inMemoryMetaStore) GetType(ctx context.Context, addr string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if obj, exists := m.objs[addr]; exists {
		return obj.typ, nil
	}
	return "", store.ErrNotFound
}

func (m *inMemoryMetaStore) GetSize(ctx context.Context, addr string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if obj, exists := m.objs[addr]; exists {
		return obj.size, nil
	}
	return 0, store.ErrNotFound
}

// HasObject returns true if the object with the given address exists.
func (m *inMemoryMetaStore) HasObject(ctx context.Context, addr string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.objs[addr]
	return exists, nil
}

func (m *inMemoryMetaStore) Close() error {
	// No resources to close for in-memory store
	return nil
}