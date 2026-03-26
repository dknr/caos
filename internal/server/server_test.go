package server

import (
	"bytes"
	"context"
	"fmt"
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

	// Try to add the same data again - should fail because AddObject now returns error for duplicates
	body2 := bytes.NewReader([]byte("hello world"))
	req2, err := http.NewRequest(http.MethodPost, "/data", body2)
	if err != nil {
		t.Fatalf("Failed to create second request: %v", err)
	}
	req2.Header.Set("Content-Type", "text/plain")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	// Should get an error since we're trying to add the same object again
	if w2.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status %d when adding duplicate object, got %d", http.StatusInternalServerError, w2.Code)
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

// TestHandleDataDelete tests the DELETE /data/:addr endpoint.
func TestHandleDataDelete(t *testing.T) {
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
	r.DELETE("/data/:addr", srv.handleDataDelete)

	// First, store some data
	body := bytes.NewReader([]byte("data to delete"))
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d for POST, got %d", http.StatusOK, w.Code)
	}

	addr := w.Body.String()

	// Verify the data exists before deletion
	exists, err := dataStore.Has(context.Background(), addr)
	if err != nil {
		t.Fatalf("Failed to check if address exists: %v", err)
	}
	if !exists {
		t.Fatalf("Address should exist after POST")
	}

	// Now delete the data
	deleteReq, err := http.NewRequest(http.MethodDelete, "/data/"+addr, nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, deleteReq)

	// Should return 204 No Content
	if w2.Code != http.StatusNoContent {
		t.Fatalf("Expected status %d for DELETE, got %d", http.StatusNoContent, w2.Code)
	}

	// Verify the data is gone
	existsAfter, err := dataStore.Has(context.Background(), addr)
	if err != nil {
		t.Fatalf("Failed to check if address exists after deletion: %v", err)
	}
	if existsAfter {
		t.Fatalf("Address should not exist after DELETE")
	}

	// Try to delete again - should still return 204 (idempotent)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, deleteReq)
	if w3.Code != http.StatusNoContent {
		t.Fatalf("Expected status %d for second DELETE, got %d", http.StatusNoContent, w3.Code)
	}
}

// TestHandleDataHead tests the HEAD /data/:addr endpoint.
func TestHandleDataHead(t *testing.T) {
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
	r.HEAD("/data/:addr", srv.handleDataHead)

	// First, store some data
	body := bytes.NewReader([]byte("head test data"))
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d for POST, got %d", http.StatusOK, w.Code)
	}

	addr := w.Body.String()

	// Now make a HEAD request
	headReq, err := http.NewRequest(http.MethodHead, "/data/"+addr, nil)
	if err != nil {
		t.Fatalf("Failed to create HEAD request: %v", err)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, headReq)

	// Should return 200 OK
	if w2.Code != http.StatusOK {
		t.Fatalf("Expected status %d for HEAD, got %d", http.StatusOK, w2.Code)
	}

	// Check headers
	if w2.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Expected Content-Type application/json, got %s", w2.Header().Get("Content-Type"))
	}
	expectedLength := fmt.Sprintf("%d", len("head test data"))
	if w2.Header().Get("Content-Length") != expectedLength {
		t.Fatalf("Expected Content-Length %s, got %s", expectedLength, w2.Header().Get("Content-Length"))
	}

	// Try to HEAD a non-existent address
	// Use a valid 64-character hex address that doesn't exist
	headReq2, err := http.NewRequest(http.MethodHead, "/data/0000000000000000000000000000000000000000000000000000000000000000", nil)
	if err != nil {
		t.Fatalf("Failed to create HEAD request for non-existent address: %v", err)
	}

	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, headReq2)

	// Should return 404 Not Found
	if w3.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d for HEAD on non-existent address, got %d", http.StatusNotFound, w3.Code)
	}
}

// TestHandleDataPostInvalidContentType tests that an invalid Content-Type header returns 400.
func TestHandleDataPostInvalidContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataStore := datastore.NewInMemoryDatastore()
	metaStore := metastore.NewInMemoryMetaStore()
	srv := NewServer(dataStore, metaStore, ":0")
	r := gin.New()
	r.POST("/data", srv.handleDataPost)

	// Invalid Content-Type (no slash)
	body := bytes.NewReader([]byte("some data"))
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "invalid")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d for invalid Content-Type, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestHandleDataPostMissingContentType tests that a missing Content-Type header defaults to application/octet-stream and succeeds.
func TestHandleDataPostMissingContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataStore := datastore.NewInMemoryDatastore()
	metaStore := metastore.NewInMemoryMetaStore()
	srv := NewServer(dataStore, metaStore, ":0")
	r := gin.New()
	r.POST("/data", srv.handleDataPost)

	// No Content-Type header
	body := bytes.NewReader([]byte("some data"))
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	// Do not set Content-Type

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d for missing Content-Type, got %d", http.StatusOK, w.Code)
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
}

// TestHandleDataPostTextNonUTF8 tests that providing non-UTF-8 data for a text/* Content-Type returns 400.
func TestHandleDataPostTextNonUTF8(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataStore := datastore.NewInMemoryDatastore()
	metaStore := metastore.NewInMemoryMetaStore()
	srv := NewServer(dataStore, metaStore, ":0")
	r := gin.New()
	r.POST("/data", srv.handleDataPost)

	// Create a byte slice with invalid UTF-8 (e.g., 0xFF 0xFE)
	invalidUTF8 := []byte{0xFF, 0xFE, 0x00, 0x01}
	body := bytes.NewReader(invalidUTF8)
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d for non-UTF-8 text data, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestHandleDataPostTextUTF8 tests that valid UTF-8 data for a text/* Content-Type succeeds.
func TestHandleDataPostTextUTF8(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataStore := datastore.NewInMemoryDatastore()
	metaStore := metastore.NewInMemoryMetaStore()
	srv := NewServer(dataStore, metaStore, ":0")
	r := gin.New()
	r.POST("/data", srv.handleDataPost)

	// Valid UTF-8 data
	validUTF8 := []byte("Hello, 世界") // Contains Chinese characters
	body := bytes.NewReader(validUTF8)
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d for valid UTF-8 text data, got %d", http.StatusOK, w.Code)
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
}

// TestHandleDataPostApplicationJsonBinary tests that binary data (non-UTF-8) with application/json Content-Type succeeds.
func TestHandleDataPostApplicationJsonBinary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataStore := datastore.NewInMemoryDatastore()
	metaStore := metastore.NewInMemoryMetaStore()
	srv := NewServer(dataStore, metaStore, ":0")
	r := gin.New()
	r.POST("/data", srv.handleDataPost)

	// Binary data that is not valid UTF-8
	binaryData := []byte{0x80, 0x81, 0x82, 0x83, 0xFF, 0xFE}
	body := bytes.NewReader(binaryData)
	req, err := http.NewRequest(http.MethodPost, "/data", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d for application/json binary data, got %d", http.StatusOK, w.Code)
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
}