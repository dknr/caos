package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/dknr/caos/internal/shared"
	"github.com/dknr/caos/store"
	"github.com/gin-gonic/gin"
)

// Server represents a CAOS server.
type Server struct {
	DataStore store.DataStore
	MetaStore store.MetaStore
	Addr      string // e.g., ":31923"
}

// NewServer creates a new CAOS server with the given stores and address.
func NewServer(ds store.DataStore, ms store.MetaStore, addr string) *Server {
	return &Server{
		DataStore: ds,
		MetaStore: ms,
		Addr:      addr,
	}
}

// Start starts the server and blocks until it is stopped.
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

	// Start the server
	return r.Run(s.Addr)
}

// handleDataPost handles POST /data requests.
func (s *Server) handleDataPost(c *gin.Context) {
	// Get Content-Type header if provided
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Validate Content-Type (for simplicity, we accept any non-empty string)
	if contentType == "" {
		c.JSON(400, gin.H{"error": "invalid content type"})
		return
	}

	// Store the data
	addr, size, err := s.DataStore.Put(c.Request.Context(), c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to store data"})
		return
	}

	// Set the size and type
	if err := s.MetaStore.AddObject(c.Request.Context(), addr, size, contentType); err != nil {
		c.JSON(500, gin.H{"error": "failed to set metadata"})
		return
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


