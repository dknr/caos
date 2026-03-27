package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/dknr/caos/internal/shared"
	"github.com/dknr/caos/store"
	"github.com/gin-gonic/gin"
)

// Server represents a CAOS server.
type Server struct {
	DataStore store.DataStore
	MetaStore store.MetaStore
	Addr      string // e.g., ":31923"
	server    *http.Server
}

// NewServer creates a new CAOS server with the given stores and address.
func NewServer(ds store.DataStore, ms store.MetaStore, addr string) *Server {
	return &Server{
		DataStore: ds,
		MetaStore: ms,
		Addr:      addr,
	}
}

// Start starts the server and returns immediately. It does not block.
func (s *Server) Start() error {
	gin.SetMode(gin.ReleaseMode) // Use release mode to reduce logging
	r := gin.New()

	// Middleware for logging and recovery
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Define routes
	r.POST("/data", s.handleDataPost)
	r.GET("/data/:addr", s.handleDataGet)
	r.DELETE("/data/:addr", s.handleDataDelete)
	r.HEAD("/data/:addr", s.handleDataHead)

	s.server = &http.Server{
		Addr:    s.Addr,
		Handler: r,
	}

	// Start the server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	return nil
}

// Stop stops the server gracefully.
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleDataPost handles POST /data requests.
func (s *Server) handleDataPost(c *gin.Context) {
	// Get Content-Type header if provided
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Parse and validate Content-Type
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.Contains(mediatype, "/") {
		c.JSON(400, gin.H{"error": "invalid content type"})
		return
	}
	contentType = mediatype

	// Read first 512 bytes to detect actual content type
	sample := make([]byte, 512)
	n, err := c.Request.Body.Read(sample)
	if err != nil && err != io.EOF {
		c.JSON(500, gin.H{"error": "failed to read body"})
		return
	}
	sample = sample[:n]

	// If declared type is text/* validate UTF-8 (using the sample as a heuristic)
	if strings.HasPrefix(contentType, "text/") && !utf8.Valid(sample) {
		c.JSON(400, gin.H{"error": "invalid UTF-8 data for text content type"})
		return
	}

	// Create a reader that first reads the sample, then the rest of the body
	fullReader := io.MultiReader(bytes.NewReader(sample), c.Request.Body)

	// Store the data
	addr, size, err := s.DataStore.Put(c.Request.Context(), fullReader)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to store data"})
		return
	}

	// Set the size and type
	if err := s.MetaStore.AddObject(c.Request.Context(), addr, size, contentType); err != nil {
		if err != store.ErrObjectExists {
			c.JSON(500, gin.H{"error": "failed to set metadata"})
			return
		}
		// Object already exists; metadata is preserved as required by spec.
	}

	// Return the address as plain text
	c.Header("Content-Type", "text/plain")
	c.String(200, addr)
}

// handleDataGet handles GET /data/:addr requests.
func (s *Server) handleDataGet(c *gin.Context) {
	addr := c.Param("addr")

	// Validate address format: 64 hex characters
	if !shared.AddrRegex.MatchString(addr) {
		c.JSON(400, gin.H{"error": "invalid address format"})
		return
	}

	// Check if the address exists
	exists, err := s.MetaStore.HasObject(c.Request.Context(), addr)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to check address"})
		return
	}
	if !exists {
		c.JSON(404, gin.H{"error": "address not found"})
		return
	}

	// Retrieve the data
	dataReader, err := s.DataStore.Get(c.Request.Context(), addr)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(404, gin.H{"error": "address not found"})
		} else {
			c.JSON(500, gin.H{"error": "failed to retrieve data"})
		}
		return
	}
	defer dataReader.Close()

	// Get the type
	contentType, err := s.MetaStore.GetType(c.Request.Context(), addr)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			contentType = "application/octet-stream"
		} else {
			c.JSON(500, gin.H{"error": "failed to get type"})
			return
		}
	}

	// Set the Content-Type header
	c.Header("Content-Type", contentType)

	// Copy the data to the response writer
	if _, err := io.Copy(c.Writer, dataReader); err != nil {
		// If we get an error while writing, we can't really do much because the headers might have been sent.
		// We'll just log it (but we don't have a logger). For now, we'll ignore.
		c.AbortWithStatus(500)
		return
	}
}

// handleDataDelete handles DELETE /data/:addr requests.
func (s *Server) handleDataDelete(c *gin.Context) {
	addr := c.Param("addr")

	// Validate address format: 64 hex characters
	if !shared.AddrRegex.MatchString(addr) {
		c.JSON(400, gin.H{"error": "invalid address format"})
		return
	}

	// Delete the data
	err := s.DataStore.Delete(c.Request.Context(), addr)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to delete data"})
		return
	}

	// Return 204 No Content
	c.Status(204)
}

// handleDataHead handles HEAD /data/:addr requests.
func (s *Server) handleDataHead(c *gin.Context) {
	addr := c.Param("addr")

	// Validate address format: 64 hex characters
	if !shared.AddrRegex.MatchString(addr) {
		c.JSON(400, gin.H{"error": "invalid address format"})
		return
	}

	// Check if the address exists using MetaStore
	exists, err := s.MetaStore.HasObject(c.Request.Context(), addr)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to check address"})
		return
	}
	if !exists {
		c.JSON(404, gin.H{"error": "address not found"})
		return
	}

	// Get the type and size
	contentType, err := s.MetaStore.GetType(c.Request.Context(), addr)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			contentType = "application/octet-stream"
		} else {
			c.JSON(500, gin.H{"error": "failed to get type"})
			return
		}
	}
	size, err := s.MetaStore.GetSize(c.Request.Context(), addr)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// This shouldn't happen if HasObject returned true, but handle gracefully
			c.JSON(500, gin.H{"error": "failed to get size"})
			return
		}
		c.JSON(500, gin.H{"error": "failed to get size"})
		return
	}

	// Set headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", size))
	c.Status(200)
}


