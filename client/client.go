// Package client provides an HTTP client for the caos content-addressed object store.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Client connects to a caos server.
type Client struct {
	base string
	hc   *http.Client
}

// New creates a new client targeting the given server URL.
func New(serverURL string) *Client {
	return &Client{
		base: strings.TrimRight(serverURL, "/"),
		hc:   &http.Client{},
	}
}

// AddData uploads data from r and returns the SHA-256 address.
func (c *Client) AddData(r io.Reader) (string, error) {
	resp, err := c.hc.Post(c.base+"/data", "application/octet-stream", r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

// GetData downloads a blob by its address and returns the body and content type.
func (c *Client) GetData(addr string) (io.ReadCloser, string, error) {
	resp, err := c.hc.Get(c.base + "/data/" + addr)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return resp.Body, resp.Header.Get("Content-Type"), nil
}

// ResolveAddr resolves a partial or full address.
// Returns matching addresses, and an error if the response is not 200 or 300.
func (c *Client) ResolveAddr(addr string) ([]string, error) {
	resp, err := c.hc.Get(c.base + "/addr/" + addr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200, 300:
		var addrs []string
		if err := json.NewDecoder(resp.Body).Decode(&addrs); err != nil {
			return nil, err
		}
		return addrs, nil
	default:
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}

// GetTags returns all tags for an address.
func (c *Client) GetTags(addr string) (map[string]string, error) {
	resp, err := c.hc.Get(c.base + "/tags/" + addr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var tags map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// GetTag returns a single tag value.
func (c *Client) GetTag(addr, tag string) (string, error) {
	resp, err := c.hc.Get(c.base + "/tags/" + addr + "/" + tag)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// SetTag sets a tag value on an address.
func (c *Client) SetTag(addr, tag, value string) error {
	req, err := http.NewRequest(http.MethodPut,
		c.base+"/tags/"+addr+"/"+tag,
		strings.NewReader(value))
	if err != nil {
		return err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

// DelTag deletes a tag from an address.
func (c *Client) DelTag(addr, tag string) error {
	req, err := http.NewRequest(http.MethodDelete,
		c.base+"/tags/"+addr+"/"+tag, nil)
	if err != nil {
		return err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

// GetName resolves a name to its address.
func (c *Client) GetName(name string) (string, error) {
	resp, err := c.hc.Get(c.base + "/name/" + url.PathEscape(name))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// SetName maps a name to an address.
func (c *Client) SetName(name, addr string) error {
	req, err := http.NewRequest(http.MethodPost,
		c.base+"/name/"+url.PathEscape(name),
		strings.NewReader(addr))
	if err != nil {
		return err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

// GetPath downloads the autoindex page for a path addr.
func (c *Client) GetPath(addr string) (string, error) {
	resp, err := c.hc.Get(c.base + "/path/" + addr + "/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetPathFile downloads a specific file from a path.
func (c *Client) GetPathFile(addr, name string) (io.ReadCloser, string, error) {
	resp, err := c.hc.Get(c.base + "/path/" + addr + "/" + url.PathEscape(name))
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return resp.Body, resp.Header.Get("Content-Type"), nil
}

// PushDir walks a directory tree, uploads each file, creates a path object,
// and returns the path address and the number of files uploaded.
func (c *Client) PushDir(root string) (string, int, error) {
	var entries []string
	count := 0

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		addr, err := c.AddData(f)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		entries = append(entries, addr+" "+rel)
		count++
		return nil
	})
	if err != nil {
		return "", 0, err
	}

	// Create path object from the listing
	pathContent := strings.Join(entries, "\n")
	pathAddr, err := c.AddData(strings.NewReader(pathContent))
	if err != nil {
		return "", 0, err
	}
	// Tag as caos/path
	if err := c.SetTag(pathAddr, "type", "caos/path"); err != nil {
		return "", 0, err
	}
	return pathAddr, count, nil
}

// PullAddr downloads a path object and writes all its files into outDir.
func (c *Client) PullAddr(pathAddr, outDir string) error {
	// Resolve the path address
	addrs, err := c.ResolveAddr(pathAddr)
	if err != nil {
		return fmt.Errorf("resolve path address: %w", err)
	}
	if len(addrs) != 1 {
		return fmt.Errorf("ambiguous address: %d matches", len(addrs))
	}
	fullAddr := addrs[0]

	// Verify it's a caos/path type
	typeVal, err := c.GetTag(fullAddr, "type")
	if err != nil || typeVal != "caos/path" {
		return fmt.Errorf("not a path object (type=%q)", typeVal)
	}

	// Download the path index
	rc, _, err := c.GetData(fullAddr)
	if err != nil {
		return fmt.Errorf("get path data: %w", err)
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read path data: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		fileAddr := parts[0]
		fileName := parts[1]

		// Reject path traversal and absolute paths
		if strings.HasPrefix(fileName, "/") {
			return fmt.Errorf("refusing to pull absolute path: %s", fileName)
		}
		if strings.Contains(fileName, "..") {
			return fmt.Errorf("refusing to pull path with traversal: %s", fileName)
		}

		filePath := filepath.Join(outDir, fileName)

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(filePath), err)
		}

		// Download the file
		rc2, _, err := c.GetData(fileAddr)
		if err != nil {
			return fmt.Errorf("get %s: %w", fileAddr[:8], err)
		}

		f, err := os.Create(filePath)
		if err != nil {
			rc2.Close()
			return fmt.Errorf("create %s: %w", filePath, err)
		}
		if _, err := io.Copy(f, rc2); err != nil {
			f.Close()
			rc2.Close()
			return fmt.Errorf("write %s: %w", filePath, err)
		}
		f.Close()
		rc2.Close()
	}
	return nil
}
