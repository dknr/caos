package store

import "errors"

// ErrNotFound is returned when a requested resource (address or tag) is not found.
var ErrNotFound = errors.New("not found")
