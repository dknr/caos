package store

import "context"

// MetaStore defines the interface for storing and retrieving metadata (size and type) for content addresses.
type MetaStore interface {
	// AddObject adds a new object with the given address, size, and type.
	// Returns an error if the object already exists.
	AddObject(ctx context.Context, addr string, size int64, typ string) error
	// HasObject returns true if the object with the given address exists.
	HasObject(ctx context.Context, addr string) (bool, error)
	// SetType updates the type for the given address.
	SetType(ctx context.Context, addr string, typ string) error
	// GetType retrieves the type for the given address.
	GetType(ctx context.Context, addr string) (string, error)
	// GetSize retrieves the size for the given address.
	GetSize(ctx context.Context, addr string) (int64, error)
	// Close closes the MetaStore, releasing any resources.
	Close() error
}