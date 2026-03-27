package store

import (
	"context"
	"io"
)

// DataStore defines the interface for storing and retrieving binary data by content address.
type DataStore interface {
	// Put stores the data from the reader and returns its content address (SHA-256 hex string) and size in bytes.
	Put(ctx context.Context, r io.Reader) (string, int64, error)
	// Get retrieves the data for the given address.
	Get(ctx context.Context, addr string) (io.ReadCloser, error)
	// Has returns true if the data for the given address exists.
	Has(ctx context.Context, addr string) (bool, error)
	// Delete removes the data for the given address.
	Delete(ctx context.Context, addr string) error
}
