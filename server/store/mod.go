package store

import (
	"io"
)

// Store combines meta (Pebble) and data (filesystem) storage.
type Store struct {
	Meta *Meta
	Data *Data
}

// Open opens both meta and data stores rooted at dir.
//   - Meta is at <dir>/meta (Pebble)
//   - Data is at <dir>/data (filesystem)
func Open(root string) (*Store, error) {
	metaDir := root + "/meta"
	dataDir := root + "/data"

	meta, err := OpenMeta(metaDir)
	if err != nil {
		return nil, err
	}

	data, err := OpenData(dataDir)
	if err != nil {
		meta.Close()
		return nil, err
	}

	return &Store{Meta: meta, Data: data}, nil
}

// Close closes both stores.
func (s *Store) Close() error {
	err1 := s.Meta.Close()
	err2 := s.Data.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// Ensure Data satisfies io.Closer.
var _ io.Closer = (*Data)(nil)

// Close is a no-op for the filesystem data store (individual files are managed
// by the OS). Provided to satisfy io.Closer.
func (d *Data) Close() error {
	return nil
}
