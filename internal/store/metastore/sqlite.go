package metastore

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/dknr/caos/store"
)

// NewSQLiteMetaStore returns a SQLite implementation of store.MetaStore.
// The dataSourceName is the path to the SQLite database file.
func NewSQLiteMetaStore(dataSourceName string) (store.MetaStore, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}
	if err := prepareSQLiteMetaStoreDb(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("preparing sqlite database: %w", err)
	}
	return &sqliteMetaStore{db: db}, nil
}

type sqliteMetaStore struct {
	db *sql.DB
}

func prepareSQLiteMetaStoreDb(db *sql.DB) error {
	// Create objs table with addr, size, and type
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS objs (
			addr TEXT PRIMARY KEY,
			size INTEGER NOT NULL,
			type TEXT NOT NULL
		);
	`); err != nil {
		return err
	}
	// Set journal mode to WAL for better concurrency
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return err
	}
	return nil
}

func (m *sqliteMetaStore) AddObject(ctx context.Context, addr string, size int64, typ string) error {
	// Insert or ignore object; using INSERT OR IGNORE to handle duplicates idempotently
	_, err := m.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO objs (addr, size, type) VALUES (?, ?, ?)
	`, addr, size, typ)
	if err != nil {
		return fmt.Errorf("inserting or ignoring addr into objs: %w", err)
	}
	return nil
}

func (m *sqliteMetaStore) SetType(ctx context.Context, addr string, typ string) error {
	// Update only the type, leaving the size unchanged.
	_, err := m.db.ExecContext(ctx, `
		UPDATE objs SET type = ? WHERE addr = ?
	`, typ, addr)
	if err != nil {
		return fmt.Errorf("updating type: %w", err)
	}
	return nil
}

func (m *sqliteMetaStore) GetType(ctx context.Context, addr string) (string, error) {
	var typ string
	err := m.db.QueryRowContext(ctx, `
		SELECT type FROM objs WHERE addr = ?
	`, addr).Scan(&typ)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", store.ErrNotFound
		}
		return "", fmt.Errorf("querying type: %w", err)
	}
	return typ, nil
}

func (m *sqliteMetaStore) GetSize(ctx context.Context, addr string) (int64, error) {
	var size int64
	err := m.db.QueryRowContext(ctx, `
		SELECT size FROM objs WHERE addr = ?
	`, addr).Scan(&size)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, store.ErrNotFound
		}
		return 0, fmt.Errorf("querying size: %w", err)
	}
	return size, nil
}

// HasObject returns true if the object with the given address exists.
func (m *sqliteMetaStore) HasObject(ctx context.Context, addr string) (bool, error) {
	var exists bool
	err := m.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM objs WHERE addr = ?)
	`, addr).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking object existence: %w", err)
	}
	return exists, nil
}

// Close closes the database connection.
func (m *sqliteMetaStore) Close() error {
	return m.db.Close()
}