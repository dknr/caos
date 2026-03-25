package datastore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dknr/caos/internal/shared"
	"github.com/dknr/caos/store"
)

// NewFilesystemDatastore returns a filesystem implementation of store.DataStore.
// The root directory is where the data files will be stored.
func NewFilesystemDatastore(root string) store.DataStore {
	return &filesystemDatastore{
		root: root,
	}
}

type filesystemDatastore struct {
	root string
}

func (f *filesystemDatastore) Put(ctx context.Context, r io.Reader) (string, int64, error) {
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
	// Ensure the root directory exists
	if err := os.MkdirAll(f.root, 0o755); err != nil {
		return "", 0, err
	}
	// Write the data to a file named by the address
	filePath := filepath.Join(f.root, addr)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", 0, err
	}
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	return addr, int64(len(data)), nil
}

func (f *filesystemDatastore) Get(ctx context.Context, addr string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Validate address format (64 hex characters)
	if !shared.AddrRegex.MatchString(addr) {
		return nil, store.ErrNotFound
	}
	filePath := filepath.Join(f.root, addr)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (f *filesystemDatastore) Has(ctx context.Context, addr string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	// Validate address format (64 hex characters)
	if !shared.AddrRegex.MatchString(addr) {
		return false, fmt.Errorf("invalid address format")
	}
	filePath := filepath.Join(f.root, addr)
	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return true, nil
}

func (f *filesystemDatastore) Delete(ctx context.Context, addr string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Validate address format (64 hex characters)
	if !shared.AddrRegex.MatchString(addr) {
		return nil
	}
	filePath := filepath.Join(f.root, addr)
	err := os.Remove(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}


