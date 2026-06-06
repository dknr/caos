package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"caos.one/caos/server/store"
)

// apiImpl implements ServerInterface with a Store backend.
type apiImpl struct {
	store   *store.Store
	homePath string
}

// compile-time check
var _ ServerInterface = (*apiImpl)(nil)

// Get implements GET / — root redirect.
func (a *apiImpl) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", a.homePath)
	w.WriteHeader(http.StatusFound)
}

// GetAddrAddr implements GET /addr/{addr} — address resolution.
func (a *apiImpl) GetAddrAddr(w http.ResponseWriter, r *http.Request, addr string) {
	if len(addr) < 6 {
		http.Error(w, `{"error":"address too short (min 6 chars)"}`, http.StatusNotFound)
		return
	}

	addrs, err := a.store.Meta.GetAddrs([]byte(addr))
	if err != nil {
		slog.Error("GetAddrAddr", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	switch len(addrs) {
	case 0:
		http.Error(w, `{"error":"no matching addresses"}`, http.StatusNotFound)
	case 1:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]string{string(addrs[0])})
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMultipleChoices)
		encoded := make([]string, len(addrs))
		for i, a := range addrs {
			encoded[i] = string(a)
		}
		json.NewEncoder(w).Encode(encoded)
	}
}

// PostData implements POST /data — data upload.
func (a *apiImpl) PostData(w http.ResponseWriter, r *http.Request) {
	limited := io.LimitReader(r.Body, 100<<20) // 100 MB max upload

	// Sniff first bytes for content type during upload
	buf := make([]byte, 512)
	n, _ := io.ReadFull(limited, buf)
	buf = buf[:n]

	// Reconstruct the full stream: sniffed prefix + remaining data
	fullReader := io.MultiReader(bytes.NewReader(buf), limited)

	addr, err := a.store.Data.AddData(fullReader)
	if err != nil {
		slog.Error("PostData", "error", err)
		http.Error(w, `{"error":"failed to store data"}`, http.StatusInternalServerError)
		return
	}

	// Register address in meta store
	a.store.Meta.AddAddr([]byte(addr))

	// Detect content type from sniffed bytes
	contentType := detectContentTypeBytes(buf)
	if contentType != "" {
		a.store.Meta.SetTag([]byte(addr), []byte("type"), []byte(contentType))
	}

	slog.Info("Data stored", "addr", addr, "type", contentType)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(addr))
}

// GetDataAddr implements GET /data/{addr} — data retrieval.
func (a *apiImpl) GetDataAddr(w http.ResponseWriter, r *http.Request, addr string) {
	if len(addr) < 6 {
		http.Error(w, `{"error":"address too short (min 6 chars)"}`, http.StatusNotFound)
		return
	}

	// Check for partial address — resolve first
	if len(addr) < 64 {
		addrs, err := a.store.Meta.GetAddrs([]byte(addr))
		if err != nil {
			slog.Error("GetDataAddr", "error", err)
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		switch len(addrs) {
		case 0:
			http.Error(w, `{"error":"address not found"}`, http.StatusNotFound)
			return
		case 1:
			addr = string(addrs[0])
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMultipleChoices)
			encoded := make([]string, len(addrs))
			for i, a := range addrs {
				encoded[i] = string(a)
			}
			json.NewEncoder(w).Encode(encoded)
			return
		}
	}

	// Full address — serve the blob
	rc, size, err := a.store.Data.GetData(addr)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, `{"error":"address not found"}`, http.StatusNotFound)
			return
		}
		slog.Error("GetDataAddr", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	// Determine content type from type tag
	contentType := "application/octet-stream"
	typeTag, err := a.store.Meta.GetTag([]byte(addr), []byte("type"))
	if err == nil && typeTag != nil {
		contentType = string(typeTag)
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, rc)
}

// GetNameName implements GET /name/{name} — name resolution.
func (a *apiImpl) GetNameName(w http.ResponseWriter, r *http.Request, name string) {
	addr, err := a.store.Meta.GetName([]byte(name))
	if err != nil {
		slog.Error("GetNameName", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if addr == nil {
		http.Error(w, `{"error":"name not found"}`, http.StatusNotFound)
		return
	}

	// Set redirect
	w.Header().Set("Location", "/data/"+string(addr))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(addr)
}

// PostNameName implements POST /name/{name} — set name alias.
func (a *apiImpl) PostNameName(w http.ResponseWriter, r *http.Request, name string) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 128))
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}
	addr := strings.TrimSpace(string(body))
	if len(addr) != 64 {
		http.Error(w, `{"error":"body must be a 64-char SHA-256 address"}`, http.StatusBadRequest)
		return
	}

	// Verify the address exists
	if !a.store.Meta.HasAddr([]byte(addr)) {
		http.Error(w, `{"error":"address does not exist in store"}`, http.StatusNotFound)
		return
	}

	if err := a.store.Meta.SetName([]byte(name), []byte(addr)); err != nil {
		slog.Error("PostNameName", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(name + " " + addr))
}

// GetPathAddr implements GET /path/{addr} — path autoindex.
func (a *apiImpl) GetPathAddr(w http.ResponseWriter, r *http.Request, addr string) {
	// Ensure trailing slash
	if !strings.HasSuffix(r.URL.Path, "/") {
		w.Header().Set("Location", r.URL.Path+"/")
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	paths, err := openPathFile(a.store, addr)
	if err != nil {
		writePathError(w, err)
		return
	}

	// Check for index.html
	idx := findPath(paths, "index.html")
	if idx != nil {
		typeTag, _ := a.store.Meta.GetTag([]byte(idx.Addr), []byte("type"))
		if string(typeTag) == "text/html" {
			rc, _, err := a.store.Data.GetData(string(idx.Addr))
			if err == nil {
				defer rc.Close()
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				io.Copy(w, rc)
				return
			}
		}
	}

	// Autoindex HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	writeAutoindex(w, addr, paths)
}

// GetPathAddrName implements GET /path/{addr}/{name} — path file serving.
func (a *apiImpl) GetPathAddrName(w http.ResponseWriter, r *http.Request, addr string, name string) {
	paths, err := openPathFile(a.store, addr)
	if err != nil {
		writePathError(w, err)
		return
	}

	match := findPath(paths, name)
	if match == nil {
		http.Error(w, `{"error":"name not found in path"}`, http.StatusNotFound)
		return
	}

	// Resolve the file's address (may be partial)
	addrs, err := a.store.Meta.GetAddrs([]byte(match.Addr))
	if err != nil {
		slog.Error("GetPathAddrName", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	switch len(addrs) {
	case 0:
		http.Error(w, `{"error":"address not found"}`, http.StatusNotFound)
		return
	case 1:
		// Serve the file
		rc, size, err := a.store.Data.GetData(string(addrs[0]))
		if err != nil {
			http.Error(w, `{"error":"file not found"}`, http.StatusNotFound)
			return
		}
		defer rc.Close()

		contentType := "application/octet-stream"
		typeTag, _ := a.store.Meta.GetTag(addrs[0], []byte("type"))
		if typeTag != nil {
			contentType = string(typeTag)
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.WriteHeader(http.StatusOK)
		io.Copy(w, rc)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMultipleChoices)
		encoded := make([]string, len(addrs))
		for i, a := range addrs {
			encoded[i] = string(a)
		}
		json.NewEncoder(w).Encode(encoded)
	}
}

// GetTagsAddr implements GET /tags/{addr} — get all tags.
func (a *apiImpl) GetTagsAddr(w http.ResponseWriter, r *http.Request, addr string) {
	if len(addr) != 64 {
		http.Error(w, `{"error":"address must be a full 64-char SHA-256"}`, http.StatusBadRequest)
		return
	}
	tags, err := a.store.Meta.GetTags([]byte(addr))
	if err != nil {
		slog.Error("GetTagsAddr", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = map[string]string{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tags)
}

// DeleteTagsAddrTag implements DELETE /tags/{addr}/{tag} — delete tag.
func (a *apiImpl) DeleteTagsAddrTag(w http.ResponseWriter, r *http.Request, addr string, tag string) {
	if len(addr) != 64 {
		http.Error(w, `{"error":"address must be a full 64-char SHA-256"}`, http.StatusBadRequest)
		return
	}
	if tag == "type" || tag == "size" {
		http.Error(w, `{"error":"cannot delete inherent tag: `+tag+`"}`, http.StatusBadRequest)
		return
	}
	if err := a.store.Meta.DelTag([]byte(addr), []byte(tag)); err != nil {
		slog.Error("DeleteTagsAddrTag", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetTagsAddrTag implements GET /tags/{addr}/{tag} — get single tag.
func (a *apiImpl) GetTagsAddrTag(w http.ResponseWriter, r *http.Request, addr string, tag string) {
	if len(addr) != 64 {
		http.Error(w, `{"error":"address must be a full 64-char SHA-256"}`, http.StatusBadRequest)
		return
	}
	val, err := a.store.Meta.GetTag([]byte(addr), []byte(tag))
	if err != nil {
		slog.Error("GetTagsAddrTag", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if val == nil {
		http.Error(w, `{"error":"tag not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

// PutTagsAddrTag implements PUT /tags/{addr}/{tag} — set tag.
func (a *apiImpl) PutTagsAddrTag(w http.ResponseWriter, r *http.Request, addr string, tag string) {
	if len(addr) != 64 {
		http.Error(w, `{"error":"address must be a full 64-char SHA-256"}`, http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, `{"error":"empty body"}`, http.StatusBadRequest)
		return
	}

	// Verify the address exists
	if !a.store.Meta.HasAddr([]byte(addr)) {
		http.Error(w, `{"error":"address does not exist in store"}`, http.StatusNotFound)
		return
	}

	if err := a.store.Meta.SetTag([]byte(addr), []byte(tag), body); err != nil {
		slog.Error("PutTagsAddrTag", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Path helpers ---

type pathEntry struct {
	Addr string
	Name string
}

// openPathFile reads and parses a caos/path object.
func openPathFile(s *store.Store, addr string) ([]pathEntry, error) {
	if len(addr) < 64 {
		addrs, err := s.Meta.GetAddrs([]byte(addr))
		if err != nil {
			return nil, pathError{http.StatusInternalServerError, ""}
		}
		switch len(addrs) {
		case 0:
			return nil, pathError{http.StatusNotFound, "unknown addr"}
		case 1:
			addr = string(addrs[0])
		default:
			return nil, pathError{http.StatusMultipleChoices, ""}
		}
	}

	// Check type tag
	typeTag, err := s.Meta.GetTag([]byte(addr), []byte("type"))
	if err != nil || string(typeTag) != "caos/path" {
		return nil, pathError{http.StatusNotFound, "not a path object"}
	}

	// Read the path data
	rc, _, err := s.Data.GetData(addr)
	if err != nil {
		return nil, pathError{http.StatusNotFound, "no data for address"}
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, pathError{http.StatusInternalServerError, ""}
	}

	return parsePathIndex(string(data)), nil
}

func parsePathIndex(text string) []pathEntry {
	var entries []pathEntry
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 && len(parts[0]) >= 6 {
			entries = append(entries, pathEntry{Addr: parts[0], Name: parts[1]})
		}
	}
	return entries
}

func findPath(entries []pathEntry, name string) *pathEntry {
	for _, e := range entries {
		if e.Name == name {
			return &e
		}
	}
	return nil
}

type pathError struct {
	status int
	reason string
}

func (e pathError) Error() string {
	if e.reason != "" {
		return e.reason
	}
	return http.StatusText(e.status)
}

func writePathError(w http.ResponseWriter, err error) {
	if pe, ok := err.(pathError); ok {
		if pe.reason != "" {
			http.Error(w, `{"error":"`+pe.reason+`"}`, pe.status)
		} else {
			w.WriteHeader(pe.status)
		}
		return
	}
	http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
}

func writeAutoindex(w http.ResponseWriter, addr string, entries []pathEntry) {
	w.Write([]byte("<!DOCTYPE html><html><head>"))
	fmt.Fprintf(w, "<title>caos: %s</title>", addr[:8])
	w.Write([]byte("<meta charset=\"utf-8\"/></head><body>"))
	w.Write([]byte("<h1>Index of " + addr[:8] + "</h1><ul>"))
	for _, e := range entries {
		fmt.Fprintf(w, "<li><a href=\"%s\">%s</a></li>\n", e.Name, e.Name)
	}
	w.Write([]byte("</ul></body></html>"))
}

// detectContentTypeBytes sniffs the content type from the first bytes of data.
func detectContentTypeBytes(buf []byte) string {
	if len(buf) == 0 {
		return ""
	}

	ct := http.DetectContentType(buf)
	// DetectContentType returns "text/plain; charset=utf-8" for unknown
	// We map that back to application/octet-stream for binary detection
	if ct == "text/plain; charset=utf-8" && !isText(buf) {
		return "application/octet-stream"
	}
	return ct
}

func isText(b []byte) bool {
	// Simple heuristic: if all bytes are printable or whitespace, it's text
	for _, c := range b {
		if c <= 0x08 || c == 0x0B || c == 0x0C || (c >= 0x0E && c <= 0x1F) {
			return false
		}
	}
	return true
}

