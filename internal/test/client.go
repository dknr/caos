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

// NewSuiteClient creates a test client.
func NewSuiteClient(serverURL string) *SuiteClient {
	return &SuiteClient{
		inner: client.New(serverURL),
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
