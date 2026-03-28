package datastore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/dknr/caos/internal/shared"
	"github.com/dknr/caos/store"
)

// Storage defines the interface for persisting byte data by key.
type Storage interface {
	// PutData stores the data by key (address).
	PutData(ctx context.Context, key string, data []byte) error
	// GetData retrieves the data for the given key.
	GetData(ctx context.Context, key string) ([]byte, error)
	// HasData returns true if the data for the given key exists.
	HasData(ctx context.Context, key string) (bool, error)
	// DeleteData removes the data for the given key.
	DeleteData(ctx context.Context, key string) error
}

// datastoreImpl implements store.DataStore using a Storage backend.
type datastoreImpl struct {
	storage Storage
}

// NewDataStore creates a DataStore that uses the given Storage backend.
func NewDataStore(storage Storage) store.DataStore {
	return &datastoreImpl{storage: storage}
}

func (d *datastoreImpl) Put(ctx context.Context, r io.Reader) (string, int64, error) {
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	// Read all data from the reader
	data, err := io.ReadAll(r)
	if err != nil {
		return "", 0, err
	}
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	// Compute SHA-256 hash
	hash := sha256.Sum256(data)
	addr := hex.EncodeToString(hash[:])
	// Store the data using the storage backend
	if err := d.storage.PutData(ctx, addr, data); err != nil {
		return "", 0, err
	}
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	return addr, int64(len(data)), nil
}

func (d *datastoreImpl) Get(ctx context.Context, addr string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Validate address format (64 hex characters)
	if !shared.AddrRegex.MatchString(addr) {
		return nil, store.ErrNotFound
	}
	// Get the data from the storage backend
	data, err := d.storage.GetData(ctx, addr)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (d *datastoreImpl) Has(ctx context.Context, addr string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	// Validate address format (64 hex characters)
	if !shared.AddrRegex.MatchString(addr) {
		return false, fmt.Errorf("invalid address format")
	}
	// Check if data exists in the storage backend
	exists, err := d.storage.HasData(ctx, addr)
	if err != nil {
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return exists, nil
}

func (d *datastoreImpl) Delete(ctx context.Context, addr string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Validate address format (64 hex characters)
	if !shared.AddrRegex.MatchString(addr) {
		return nil
	}
	// Delete the data from the storage backend
	if err := d.storage.DeleteData(ctx, addr); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}