package store

import (
	"io"
	"log/slog"

	"github.com/cockroachdb/pebble"
)

const (
	addrPrefix = "a/"
	tagPrefix  = "t/"
	namePrefix = "n/"
)

// Meta is a Pebble-backed key-value store for caos metadata.
// Key schema:
//
//	a/<addr>       → ""            address existence marker
//	t/<addr>/<tag> → <value>       tag value
//	n/<name>       → <addr>        name→address mapping
type Meta struct {
	db *pebble.DB
}

// OpenMeta opens (or creates) a Pebble meta store at the given path.
func OpenMeta(dir string) (*Meta, error) {
	db, err := pebble.Open(dir, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	slog.Info("Opened meta store", "path", dir)
	return &Meta{db: db}, nil
}

// Close closes the underlying Pebble DB.
func (m *Meta) Close() error {
	return m.db.Close()
}

// ---------------------------------------------------------------------------
// Address operations
// ---------------------------------------------------------------------------

// AddAddr registers an address in the store. It is idempotent.
func (m *Meta) AddAddr(addr []byte) error {
	k := makeKey(addrPrefix, addr)
	return m.db.Set(k, nil, pebble.NoSync)
}

// HasAddr returns true if the address exists in the store.
func (m *Meta) HasAddr(addr []byte) bool {
	k := makeKey(addrPrefix, addr)
	_, closer, err := m.db.Get(k)
	if err != nil {
		return false
	}
	closer.Close()
	return true
}

// GetAddrs returns all stored addresses whose hex string starts with prefix.
// prefix must be at least 6 characters.
func (m *Meta) GetAddrs(prefix []byte) ([][]byte, error) {
	start := makeKey(addrPrefix, prefix)
	end := prefixUpperBound(start)

	iter, err := m.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var results [][]byte
	for iter.SeekGE(start); iter.Valid(); iter.Next() {
		key := iter.Key()
		// Strip the "a/" prefix
		addr := stripPrefix(addrPrefix, key)
		results = append(results, addr)
	}
	return results, iter.Close()
}

// CountAddrs returns the total number of stored addresses.
func (m *Meta) CountAddrs() (int, error) {
	prefix := []byte(addrPrefix)
	iter, err := m.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: prefixUpperBound(prefix),
	})
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	count := 0
	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		count++
	}
	return count, iter.Close()
}

// ---------------------------------------------------------------------------
// Tag operations
// ---------------------------------------------------------------------------

// SetTag sets a tag value for an address. addr must be a full 64-char hex string.
func (m *Meta) SetTag(addr, tag, value []byte) error {
	k := makeTagKey(addr, tag)
	return m.db.Set(k, value, pebble.NoSync)
}

// GetTag returns the value for a specific tag on an address.
// Returns nil if the tag does not exist.
func (m *Meta) GetTag(addr, tag []byte) ([]byte, error) {
	k := makeTagKey(addr, tag)
	val, closer, err := m.db.Get(k)
	if err == pebble.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Copy the value before closing
	v := make([]byte, len(val))
	copy(v, val)
	closer.Close()
	return v, nil
}

// GetTags returns all tags for an address as a map of key→value.
func (m *Meta) GetTags(addr []byte) (map[string]string, error) {
	start := makeTagPrefix(addr)
	end := prefixUpperBound(start)

	iter, err := m.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	tags := make(map[string]string)
	for iter.SeekGE(start); iter.Valid(); iter.Next() {
		key := iter.Key()
		val := iter.Value()
		// Key is "t/<addr>/<tag>", strip prefix to get tag name
		tagName := tagNameFromKey(start, key)
		tags[string(tagName)] = string(val)
	}
	return tags, iter.Close()
}

// DelTag deletes a tag from an address. Idempotent.
func (m *Meta) DelTag(addr, tag []byte) error {
	k := makeTagKey(addr, tag)
	return m.db.Delete(k, pebble.NoSync)
}

// ---------------------------------------------------------------------------
// Name operations
// ---------------------------------------------------------------------------

// SetName maps a human-readable name to an address.
func (m *Meta) SetName(name, addr []byte) error {
	k := makeKey(namePrefix, name)
	return m.db.Set(k, addr, pebble.NoSync)
}

// GetName resolves a name to its address.
// Returns nil if the name does not exist.
func (m *Meta) GetName(name []byte) ([]byte, error) {
	k := makeKey(namePrefix, name)
	val, closer, err := m.db.Get(k)
	if err == pebble.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	v := make([]byte, len(val))
	copy(v, val)
	closer.Close()
	return v, nil
}

// ---------------------------------------------------------------------------
// Key helpers
// ---------------------------------------------------------------------------

func makeKey(prefix string, suffix []byte) []byte {
	k := make([]byte, len(prefix)+len(suffix))
	copy(k, prefix)
	copy(k[len(prefix):], suffix)
	return k
}

func stripPrefix(prefix string, key []byte) []byte {
	if len(key) <= len(prefix) {
		return nil
	}
	v := make([]byte, len(key)-len(prefix))
	copy(v, key[len(prefix):])
	return v
}

func makeTagKey(addr, tag []byte) []byte {
	// t/<addr>/<tag>
	prefix := tagPrefix
	k := make([]byte, len(prefix)+len(addr)+1+len(tag))
	copy(k, prefix)
	copy(k[len(prefix):], addr)
	k[len(prefix)+len(addr)] = '/'
	copy(k[len(prefix)+len(addr)+1:], tag)
	return k
}

func makeTagPrefix(addr []byte) []byte {
	// t/<addr>/
	prefix := tagPrefix
	k := make([]byte, len(prefix)+len(addr)+1)
	copy(k, prefix)
	copy(k[len(prefix):], addr)
	k[len(prefix)+len(addr)] = '/'
	return k
}

func tagNameFromKey(prefix, key []byte) []byte {
	// key is "t/<addr>/<tag>", prefix is "t/<addr>/"
	if len(key) <= len(prefix) {
		return nil
	}
	v := make([]byte, len(key)-len(prefix))
	copy(v, key[len(prefix):])
	return v
}

// prefixUpperBound returns a key that is guaranteed to be greater than any key
// starting with the given prefix. It works by incrementing the last byte.
func prefixUpperBound(prefix []byte) []byte {
	// Copy the prefix and increment the last byte
	b := make([]byte, len(prefix))
	copy(b, prefix)
	// If the last byte is 0xFF, append a 0x00 byte to create a sentinel
	if b[len(b)-1] == 0xFF {
		b = append(b, 0x00)
	} else {
		b[len(b)-1]++
	}
	return b
}

// Ensure Meta satisfies io.Closer.
var _ io.Closer = (*Meta)(nil)
