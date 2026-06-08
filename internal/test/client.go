package test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"caos.one/caos/client"
)

// SuiteClient wraps client.Client with additional test helpers.
type SuiteClient struct {
	inner *client.Client
	hc    *http.Client
	base  string
}

// NewSuiteClient creates a test client with the given API key set for write operations.
func NewSuiteClient(serverURL string, apiKey string) *SuiteClient {
	inner := client.New(serverURL)
	inner.SetAPIKey(apiKey)
	return &SuiteClient{
		inner: inner,
		hc:    &http.Client{},
		base:  strings.TrimRight(serverURL, "/"),
	}
}

// AddData delegates to the inner client.
func (c *SuiteClient) AddData(r io.Reader) (string, error) {
	return c.inner.AddData(r)
}

// GetTag delegates to the inner client.
func (c *SuiteClient) GetTag(addr, tag string) (string, error) {
	return c.inner.GetTag(addr, tag)
}

// SetTag delegates to the inner client.
func (c *SuiteClient) SetTag(addr, tag, value string) error {
	return c.inner.SetTag(addr, tag, value)
}

// DelTag delegates to the inner client.
func (c *SuiteClient) DelTag(addr, tag string) error {
	return c.inner.DelTag(addr, tag)
}

// ResolveAddr delegates to the inner client.
func (c *SuiteClient) ResolveAddr(addr string) ([]string, error) {
	return c.inner.ResolveAddr(addr)
}

// GetName delegates to the inner client.
func (c *SuiteClient) GetName(name string) (string, error) {
	return c.inner.GetName(name)
}

// SetName delegates to the inner client.
func (c *SuiteClient) SetName(name, addr string) error {
	return c.inner.SetName(name, addr)
}

// GetTags delegates to the inner client.
func (c *SuiteClient) GetTags(addr string) (map[string]string, error) {
	return c.inner.GetTags(addr)
}

// UploadThenVerify uploads data and verifies the returned hash matches.
func (c *SuiteClient) UploadThenVerify(r io.Reader, expectedHash string) (string, error) {
	addr, err := c.inner.AddData(r)
	if err != nil {
		return "", err
	}
	if expectedHash != "" && addr != expectedHash {
		return "", fmt.Errorf("hash mismatch: got %s, want %s", addr, expectedHash)
	}
	return addr, nil
}

// GetDataWithStatus downloads data and returns the response status.
func (c *SuiteClient) GetDataWithStatus(addr string) (io.ReadCloser, string, int, error) {
	resp, err := c.hc.Get(c.base + "/data/" + addr)
	if err != nil {
		return nil, "", 0, err
	}
	return resp.Body, resp.Header.Get("Content-Type"), resp.StatusCode, nil
}

// ResolveAddrWithStatus resolves an address and returns the status code.
func (c *SuiteClient) ResolveAddrWithStatus(addr string) ([]string, int, error) {
	resp, err := c.hc.Get(c.base + "/addr/" + addr)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	status := resp.StatusCode
	if status != 200 && status != 300 && status != 404 {
		return nil, status, fmt.Errorf("unexpected status %d", status)
	}
	return nil, status, nil
}

// GetPathWithStatus gets a path page and returns the status code.
func (c *SuiteClient) GetPathWithStatus(addr string) (string, int, error) {
	resp, err := c.hc.Get(c.base + "/path/" + addr + "/")
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body), resp.StatusCode, nil
}

// PathEscape wraps url.PathEscape.
func PathEscape(s string) string {
	return url.PathEscape(s)
}

// PushDir delegates to the inner client.
func (c *SuiteClient) PushDir(root string) (string, int, error) {
	return c.inner.PushDir(root)
}

// PullAddr delegates to the inner client.
func (c *SuiteClient) PullAddr(pathAddr, outDir string) error {
	return c.inner.PullAddr(pathAddr, outDir)
}

// --- Auth test helpers — make raw requests without API key ---

func (c *SuiteClient) AddDataNoAuth(r io.Reader) (int, error) {
	req, err := http.NewRequest(http.MethodPost, c.base+"/data", r)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *SuiteClient) SetTagNoAuth(addr, tag, value string) (int, error) {
	req, err := http.NewRequest(http.MethodPut,
		c.base+"/tags/"+addr+"/"+tag,
		strings.NewReader(value))
	if err != nil {
		return 0, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *SuiteClient) DelTagNoAuth(addr, tag string) (int, error) {
	req, err := http.NewRequest(http.MethodDelete,
		c.base+"/tags/"+addr+"/"+tag, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *SuiteClient) SetNameNoAuth(name, addr string) (int, error) {
	req, err := http.NewRequest(http.MethodPost,
		c.base+"/name/"+url.PathEscape(name),
		strings.NewReader(addr))
	if err != nil {
		return 0, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *SuiteClient) AddDataWithWrongKey(r io.Reader) (int, error) {
	req, err := http.NewRequest(http.MethodPost, c.base+"/data", r)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-API-Key", "wrong-key")
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *SuiteClient) WriteRequestNoAuth(method, path string, body io.Reader) (int, error) {
	req, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "text/plain")
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	// drain body
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func (c *SuiteClient) WriteRequestWrongKey(method, path string, body io.Reader) (int, error) {
	req, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "text/plain")
	}
	req.Header.Set("X-API-Key", "wrong-key")
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func (c *SuiteClient) ReadRequest(method, path string) (int, error) {
	req, err := http.NewRequest(method, c.base+path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func (c *SuiteClient) WriteRequestWithKey(method, path, key string, body io.Reader) (int, error) {
	req, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "text/plain")
	}
	req.Header.Set("X-API-Key", key)
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}
