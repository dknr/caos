package main

import (
    "bytes"
    "context"
    "flag"
    "fmt"
    "io"
    "log"
    "math/rand"
    "net/http"
    "net/url"
    "sync"
    "time"
)

// TestResult represents the outcome of a compliance test.
type TestResult struct {
    name   string
    passed bool
    err    error
}

func main() {
    // Parse command‑line flags.
    serverURL := flag.String("url", "http://localhost:31923", "CAOS server URL")
    flag.Parse()

    base, err := url.Parse(*serverURL)
    if err != nil {
        log.Fatalf("invalid url: %v", err)
    }

    // Run all edge case tests
    tests := []struct {
        name string
        fn   func(string) error
    }{
        {"POST missing Content-Type", testPostMissingContentType},
        {"POST invalid Content-Type", testPostInvalidContentType},
        {"GET malformed address", testGetMalformedAddress},
        {"DELETE non-existent", testDeleteNonexistent},
        {"HEAD non-existent", testHeadNonexistent},
        {"POST empty body", testPostEmptyBody},
        {"GET after DELETE", testGetAfterDelete},
        {"Concurrent POST same data", testConcurrentPostSameData},
        {"Large payload POST", testLargePayload},
    }

    var failed []string
    for _, test := range tests {
        if err := test.fn(base.String()); err != nil {
            log.Printf("Test %s failed: %v", test.name, err)
            failed = append(failed, test.name)
        } else {
            log.Printf("Test %s passed", test.name)
        }
    }

    if len(failed) > 0 {
        log.Fatalf("Failed tests: %v", failed)
    }
    fmt.Println("All compliance tests passed.")
}

func postData(u *url.URL, data []byte, ct string) (string, error) {
    req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", ct)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    resp, err := http.DefaultClient.Do(req.WithContext(ctx))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    addr := string(body)
    if len(addr) != 64 {
        return "", fmt.Errorf("address length %d, expected 64", len(addr))
    }
    return addr, nil
}

func getData(u *url.URL, expected []byte) error {
    resp, err := http.Get(u.String())
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("GET status %d", resp.StatusCode)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    if !bytes.Equal(body, expected) {
        return fmt.Errorf("body mismatch: got %q", string(body))
    }
    return nil
}

func headData(u *url.URL) error {
    req, err := http.NewRequest("HEAD", u.String(), nil)
    if err != nil {
        return err
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("HEAD status %d", resp.StatusCode)
    }
    // Validate Content-Type and Content-Length.
    ct := resp.Header.Get("Content-Type")
    if ct == "" {
        return fmt.Errorf("missing Content-Type header")
    }
    cl := resp.Header.Get("Content-Length")
    if cl == "" {
        return fmt.Errorf("missing Content-Length header")
    }
    return nil
}

func deleteData(u *url.URL) error {
    req, err := http.NewRequest("DELETE", u.String(), nil)
    if err != nil {
        return err
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        return fmt.Errorf("DELETE status %d", resp.StatusCode)
    }
    return nil
}

func getDataNotFound(u *url.URL) error {
    resp, err := http.Get(u.String())
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusNotFound {
        return fmt.Errorf("expected 404, got %d", resp.StatusCode)
    }
    return nil
}

// Test POST missing Content-Type
func testPostMissingContentType(baseURL string) error {
    postURL := baseURL + "/data"
    req, err := http.NewRequest("POST", postURL, bytes.NewReader([]byte("test")))
    if err != nil {
        return err
    }
    // No Content-Type header
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    resp, err := http.DefaultClient.Do(req.WithContext(ctx))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    // Server should default to application/octet-stream and accept it
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("expected 200 got %d", resp.StatusCode)
    }
    return nil
}

// Test POST with invalid Content-Type
func testPostInvalidContentType(baseURL string) error {
    postURL := baseURL + "/data"
    req, err := http.NewRequest("POST", postURL, bytes.NewReader([]byte("test")))
    if err != nil {
        return err
    }
    // Invalid Content-Type - no slash will cause mime.ParseMediaType to return mediatype without "/"
    req.Header.Set("Content-Type", "invalid")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    resp, err := http.DefaultClient.Do(req.WithContext(ctx))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusBadRequest {
        return fmt.Errorf("expected 400 got %d", resp.StatusCode)
    }
    return nil
}

// Test GET with malformed address
func testGetMalformedAddress(baseURL string) error {
    getURL := baseURL + "/data/abc"
    resp, err := http.Get(getURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusBadRequest {
        return fmt.Errorf("expected 400 got %d", resp.StatusCode)
    }
    return nil
}

// Test DELETE non-existent address
func testDeleteNonexistent(baseURL string) error {
    addr := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
    delURL := baseURL + "/data/" + addr
    req, err := http.NewRequest("DELETE", delURL, nil)
    if err != nil {
        return err
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    // Should return 204 No Content or 200 OK for non-existent resource
    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        return fmt.Errorf("expected 204 or 200 got %d", resp.StatusCode)
    }
    return nil
}

// Test HEAD on non-existent address
func testHeadNonexistent(baseURL string) error {
    addr := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
    headURL := baseURL + "/data/" + addr
    req, err := http.NewRequest("HEAD", headURL, nil)
    if err != nil {
        return err
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusNotFound {
        return fmt.Errorf("expected 404 got %d", resp.StatusCode)
    }
    return nil
}

// Test POST with empty body
func testPostEmptyBody(baseURL string) error {
    postURL := baseURL + "/data"
    req, err := http.NewRequest("POST", postURL, bytes.NewReader([]byte("")))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/octet-stream")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    resp, err := http.DefaultClient.Do(req.WithContext(ctx))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("expected 200 got %d", resp.StatusCode)
    }
    return nil
}

// Test GET after DELETE returns 404
func testGetAfterDelete(baseURL string) error {
    // Store data
    addr, err := postDataURL(baseURL, []byte("test"), "text/plain")
    if err != nil {
        return err
    }
    // Delete
    delURL := baseURL + "/data/" + addr
    req, err := http.NewRequest("DELETE", delURL, nil)
    if err != nil {
        return err
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    resp.Body.Close()
    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        return fmt.Errorf("delete expected 204/200 got %d", resp.StatusCode)
    }
    // Get
    getURL := baseURL + "/data/" + addr
    resp, err = http.Get(getURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusNotFound {
        return fmt.Errorf("expected 404 got %d", resp.StatusCode)
    }
    return nil
}

// Test concurrency: multiple POST same data
func testConcurrentPostSameData(baseURL string) error {
    payload := []byte("concurrent data")
    var wg sync.WaitGroup
    results := make(chan string, 5)
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            addr, err := postDataURL(baseURL, payload, "text/plain")
            if err != nil {
                return
            }
            results <- addr
        }()
    }
    wg.Wait()
    close(results)
    if len(results) == 0 {
        return fmt.Errorf("no responses")
    }
    // Verify all addresses are the same (deduplication)
    first := <-results
    for a := range results {
        if a != first {
            return fmt.Errorf("different addresses: %s vs %s", first, a)
        }
    }
    return nil
}

// Test POST with very large payload
func testLargePayload(baseURL string) error {
    payload := make([]byte, 5*1024*1024) // 5MB
    rand.Read(payload)
    addr, err := postDataURL(baseURL, payload, "application/octet-stream")
    if err != nil {
        return err
    }
    if len(addr) != 64 {
        return fmt.Errorf("invalid addr length %d", len(addr))
    }
    return nil
}

// Helper function to POST data and return address
func postDataURL(baseURL string, data []byte, contentType string) (string, error) {
    postURL := baseURL + "/data"
    req, err := http.NewRequest("POST", postURL, bytes.NewReader(data))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", contentType)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    resp, err := http.DefaultClient.Do(req.WithContext(ctx))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("POST status %d", resp.StatusCode)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}
