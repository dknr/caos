package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/dknr/caos/store"
)

// addrRegex matches a 64-character hexadecimal string (SHA-256)
var addrRegex = regexp.MustCompile("^[0-9a-fA-F]{64}$")

// Client provides methods to interact with a CAOS server over HTTP.
type Client struct {
	BaseURL string
	HTTPClient *http.Client
}

// NewClient creates a new CAOS client.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Add stores data to the CAOS server and returns its address.
func (c *Client) Add(ctx context.Context, r io.Reader, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/data", r)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var addr bytes.Buffer
	if _, err := io.Copy(&addr, resp.Body); err != nil {
		return "", err
	}

	// Validate that we got a 64-character hex address
	addrStr := addr.String()
	if !addrRegex.MatchString(addrStr) {
		return "", fmt.Errorf("invalid address received from server: %s", addrStr)
	}

	return addrStr, nil
}

// Get retrieves data from the CAOS server for the given address.
func (c *Client) Get(ctx context.Context, addr string) (io.ReadCloser, error) {
	// Validate address format
	if !addrRegex.MatchString(addr) {
		return nil, fmt.Errorf("address too short")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/data/"+addr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	return resp.Body, nil
}

// Has checks if the CAOS server has data for the given address.
func (c *Client) Has(ctx context.Context, addr string) (bool, error) {
	// Validate address format
	if !addrRegex.MatchString(addr) {
		return false, fmt.Errorf("address too short")
	}

	// We'll use the Get endpoint and check for 404
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/data/"+addr, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("server returned unexpected status %d: %s", resp.StatusCode, resp.Status)
	}
}

// Delete removes data from the CAOS server for the given address.
// Note: This endpoint is not defined in the Level 0 API, so we're implementing it for completeness.
func (c *Client) Delete(ctx context.Context, addr string) error {
	// Validate address format
	if !addrRegex.MatchString(addr) {
		return fmt.Errorf("address too short")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+"/data/"+addr, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// GetTag retrieves a tag value from the CAOS server for the given address and tag.
func (c *Client) GetTag(ctx context.Context, addr, tag string) (string, error) {
	// Validate address format
	if !addrRegex.MatchString(addr) {
		return "", fmt.Errorf("address too short")
	}

	u, err := url.Parse(c.BaseURL + "/data/" + addr + "/tags/" + tag)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", store.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var value bytes.Buffer
	if _, err := io.Copy(&value, resp.Body); err != nil {
		return "", err
	}

	return value.String(), nil
}

// SetTag sets a tag value on the CAOS server for the given address and tag.
func (c *Client) SetTag(ctx context.Context, addr, tag, value string) error {
	// Validate address format
	if !addrRegex.MatchString(addr) {
		return fmt.Errorf("address too short")
	}

	u, err := url.Parse(c.BaseURL + "/data/" + addr + "/tags/" + tag)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewBufferString(value))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// GetTags retrieves all tags from the CAOS server for the given address.
func (c *Client) GetTags(ctx context.Context, addr string) (map[string]string, error) {
	// Validate address format
	if !addrRegex.MatchString(addr) {
		return nil, fmt.Errorf("address too short")
	}

	u, err := url.Parse(c.BaseURL + "/data/" + addr + "/tags")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, store.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var tags map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// DeleteTag removes a tag from the CAOS server for the given address and tag.
// Note: This endpoint is not defined in the Level 0 API, so we're implementing it for completeness.
func (c *Client) DeleteTag(ctx context.Context, addr, tag string) error {
	// Validate address format
	if !addrRegex.MatchString(addr) {
		return fmt.Errorf("address too short")
	}

	u, err := url.Parse(c.BaseURL + "/data/" + addr + "/tags/" + tag)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

