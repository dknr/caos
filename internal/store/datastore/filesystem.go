package datastore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dknr/caos/internal/shared"
	"github.com/dknr/caos/store"
)

// NewFilesystemDatastore returns a filesystem implementation of store.DataStore.
// The root directory is where the data files will be stored.
func NewFilesystemDatastore(root string) store.DataStore {
	return NewDataStore(&filesystemStorage{root: root})
}

// filesystemStorage implements Storage using the filesystem.
type filesystemStorage struct {
	root string
}

func (f *filesystemStorage) PutData(ctx context.Context, key string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Ensure the root directory exists
	if err := os.MkdirAll(f.root, 0o755); err != nil {
		return err
	}
	// Create temporary directory
	tempDir := filepath.Join(f.root, "temp")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(tempDir, "tmp-")
	if err != nil {
		return err
	}
	defer func() { _ = tmpFile.Close() }()
	n, err := tmpFile.Write(data)
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return err
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return err
	}
	finalPath := filepath.Join(f.root, key)
	if err := os.Rename(tmpFile.Name(), finalPath); err != nil {
		_ = os.Remove(tmpFile.Name())
		return err
	}
	// Verify we wrote the correct amount of data
	if n != len(data) {
		_ = os.Remove(finalPath)
		return fmt.Errorf("wrote %d bytes, expected %d", n, len(data))
	}
	return nil
}

func (f *filesystemStorage) GetData(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Validate key format (64 hex characters)
	if !shared.AddrRegex.MatchString(key) {
		return nil, store.ErrNotFound
	}
	filePath := filepath.Join(f.root, key)
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
	return data, nil
}

func (f *filesystemStorage) HasData(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	// Validate key format (64 hex characters)
	if !shared.AddrRegex.MatchString(key) {
		return false, fmt.Errorf("invalid key format")
	}
	filePath := filepath.Join(f.root, key)
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

func (f *filesystemStorage) DeleteData(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Validate key format (64 hex characters)
	if !shared.AddrRegex.MatchString(key) {
		return nil
	}
	filePath := filepath.Join(f.root, key)
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