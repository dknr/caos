package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

// Data provides content-addressed blob storage on the filesystem.
// Blobs are stored as files named by their SHA-256 hex string.
type Data struct {
	dir string
}

// OpenData opens (or creates) a data store rooted at dir.
func OpenData(dir string) (*Data, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	slog.Info("Opened data store", "path", dir)
	return &Data{dir: dir}, nil
}

// AddData writes a blob from r, computes its SHA-256 hash, and stores it.
// Returns the hex-encoded SHA-256 address.
func (d *Data) AddData(r io.Reader) (string, error) {
	// Write to a temp file first, then atomically rename
	tmpFile, err := os.CreateTemp(d.dir, "*.tmp")
	if err != nil {
		return "", err
	}
	tmpName := tmpFile.Name()

	h := sha256.New()
	tee := io.TeeReader(r, h)

	_, err = io.Copy(tmpFile, tee)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return "", err
	}

	addr := hex.EncodeToString(h.Sum(nil))
	finalPath := filepath.Join(d.dir, addr)

	// Atomic rename
	if err := os.Rename(tmpName, finalPath); err != nil {
		os.Remove(tmpName)
		return "", err
	}

	return addr, nil
}

// GetData opens a blob for reading by its address.
// Returns the reader and the content length.
// Returns os.ErrNotExist if the blob is not found.
func (d *Data) GetData(addr string) (io.ReadCloser, int64, error) {
	path := filepath.Join(d.dir, addr)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, 0, os.ErrNotExist
		}
		return nil, 0, err
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}

	return f, stat.Size(), nil
}

// HasData returns true if a blob with the given address exists.
func (d *Data) HasData(addr string) bool {
	path := filepath.Join(d.dir, addr)
	_, err := os.Stat(path)
	return err == nil
}

// DelData removes a blob by address. Idempotent.
func (d *Data) DelData(addr string) error {
	path := filepath.Join(d.dir, addr)
	if err := os.Remove(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}
