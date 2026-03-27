package store

import "errors"

// ErrNotFound is returned when a requested resource (address or tag) is not found.
var ErrNotFound = errors.New("not found")

// ErrObjectExists is returned when trying to add an object that already exists.
var ErrObjectExists = errors.New("object already exists")
