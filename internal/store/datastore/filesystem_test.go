package datastore

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dknr/caos/internal/shared"
)

func TestFilesystemDatastore_PutGet(t *testing.T) {
	// Create a temporary directory for the test
	dir, err := os.MkdirTemp("", "caos-datastore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	ds := NewFilesystemDatastore(dir)
	ctx := context.Background()

	// Test putting and getting data
	data := strings.NewReader("hello world")
	addr, size, err := ds.Put(ctx, data)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Validate address format (64 hex characters)
	if !shared.AddrRegex.MatchString(addr) {
		t.Fatalf("expected 64 char hex address, got %s", addr)
	}

	// Validate size
	if size != 11 {
		t.Fatalf("expected size 11, got %d", size)
	}

	// Get the data back
	reader, err := ds.Get(ctx, addr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer reader.Close()

	// Read the data back
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading data failed: %v", err)
	}
	if string(got) != "hello world" {
		t.Fatalf("expected 'hello world', got %s", string(got))
	}
}

func TestFilesystemDatastore_Has(t *testing.T) {
	// Create a temporary directory for the test
	dir, err := os.MkdirTemp("", "caos-datastore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	ds := NewFilesystemDatastore(dir)
	ctx := context.Background()

	// Test Has for existing addr
	data := strings.NewReader("test")
	addr, _, err := ds.Put(ctx, data)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	exists, err := ds.Has(ctx, addr)
	if err != nil {
		t.Fatalf("Has failed: %v", err)
	}
	if !exists {
		t.Fatalf("expected Has to return true for existing addr")
	}

	// Test Has for non-existing addr
	fakeAddr := "0000000000000000000000000000000000000000000000000000000000000000"
	if !shared.AddrRegex.MatchString(fakeAddr) {
		t.Fatalf("test fakeAddr is not a valid hex address")
	}
	exists, err = ds.Has(ctx, fakeAddr)
	if err != nil {
		t.Fatalf("Has failed: %v", err)
	}
	if exists {
		t.Fatalf("expected Has to return false for non-existing addr")
	}
}

func TestFilesystemDatastore_Delete(t *testing.T) {
	// Create a temporary directory for the test
	dir, err := os.MkdirTemp("", "caos-datastore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	ds := NewFilesystemDatastore(dir)
	ctx := context.Background()

	// Put some data
	data := strings.NewReader("to delete")
	addr, _, err := ds.Put(ctx, data)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify it exists
	exists, err := ds.Has(ctx, addr)
	if err != nil {
		t.Fatalf("Has failed: %v", err)
	}
	if !exists {
		t.Fatalf("expected addr to exist after Put")
	}

	// Delete it
	err = ds.Delete(ctx, addr)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	exists, err = ds.Has(ctx, addr)
	if err != nil {
		t.Fatalf("Has failed: %v", err)
	}
	if exists {
		t.Fatalf("expected addr to not exist after Delete")
	}
}