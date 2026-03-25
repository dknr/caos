package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dknr/caos/internal/store/datastore"
	"github.com/dknr/caos/internal/store/metastore"
	"github.com/gin-gonic/gin"
)

func TestHandleDataPost(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create in-memory stores for testing
	dataStore := datastore.NewInMemoryDatastore()
	metaStore := metastore.NewInMemoryMetaStore()

	// Create server
	srv := NewServer(dataStore, metaStore, ":0") // :0 lets OS pick a free port

	// Create a test router
	r := gin.New()
	r.POST("/data", srv.handleDataPost)

	// Create a test request body
	body := bytes.NewReader([]byte("hello world"))
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// The response body should be a 64-character hex address
	addr := w.Body.String()
	if len(addr) != 64 {
		t.Fatalf("Expected 64-character address, got %d: %s", len(addr), addr)
	}
	// Validate it's hex
	for _, c := range addr {
		if c < '0' || c > '9' {
			if c < 'a' || c > 'f' {
				if c < 'A' || c > 'F' {
					t.Fatalf("Address contains non-hex character: %c", c)
				}
			}
		}
	}

	// Verify the data was stored
	exists, err := dataStore.Has(context.Background(), addr)
	if err != nil {
		t.Fatalf("Failed to check if address exists: %v", err)
	}
	if !exists {
		t.Fatalf("Address should exist after POST")
	}

	// Verify metadata was stored
	size, err := metaStore.GetSize(context.Background(), addr)
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}
	if size != 11 {
		t.Fatalf("Expected size 11, got %d", size)
	}

	typ, err := metaStore.GetType(context.Background(), addr)
	if err != nil {
		t.Fatalf("Failed to get type: %v", err)
	}
	if typ != "text/plain" {
		t.Fatalf("Expected type 'text/plain', got %s", typ)
	}
}

func TestHandleDataGet(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create in-memory stores for testing
	dataStore := datastore.NewInMemoryDatastore()
	metaStore := metastore.NewInMemoryMetaStore()

	// Create server
	srv := NewServer(dataStore, metaStore, ":0")

	// Create a test router
	r := gin.New()
	r.POST("/data", srv.handleDataPost)
	r.GET("/data/:addr", srv.handleDataGet)

	// First, store some data
	body := bytes.NewReader([]byte("test data"))
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	addr := w.Body.String()

	// Now retrieve the data
	req, err = http.NewRequest(http.MethodGet, "/data/"+addr, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	// Check body
	expected := "test data"
	if w.Body.String() != expected {
		t.Fatalf("Expected body %q, got %q", expected, w.Body.String())
	}
}