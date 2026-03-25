package datastore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/dknr/caos/store"
)

// addrRegex matches a 64-character hexadecimal string (SHA-256)
var addrRegex = regexp.MustCompile("^[0-9a-fA-F]{64}$")

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
	// Read all data from the reader
	data, err := io.ReadAll(r)
	if err != nil {
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
	return addr, int64(len(data)), nil
}

func (f *filesystemDatastore) Get(ctx context.Context, addr string) (io.ReadCloser, error) {
	// Validate address format (64 hex characters)
	if !addrRegex.MatchString(addr) {
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
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (f *filesystemDatastore) Has(ctx context.Context, addr string) (bool, error) {
	// Validate address format (64 hex characters)
	if !addrRegex.MatchString(addr) {
		return false, nil
	}
	filePath := filepath.Join(f.root, addr)
	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *filesystemDatastore) Delete(ctx context.Context, addr string) error {
	// Validate address format (64 hex characters)
	if !addrRegex.MatchString(addr) {
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
	return nil
}


