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
	// Ensure the root directory exists
	if err := os.MkdirAll(f.root, 0o755); err != nil {
		return "", 0, err
	}
	// Create temporary directory
	tempDir := filepath.Join(f.root, "temp")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", 0, err
	}
	tmpFile, err := os.CreateTemp(tempDir, "tmp-")
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = tmpFile.Close() }()
	hasher := sha256.New()
	writer := io.MultiWriter(tmpFile, hasher)
	n, err := io.Copy(writer, r)
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", 0, err
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", 0, err
	}
	addr := hex.EncodeToString(hasher.Sum(nil))
	finalPath := filepath.Join(f.root, addr)
	if err := os.Rename(tmpFile.Name(), finalPath); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", 0, err
	}
	return addr, n, nil
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


