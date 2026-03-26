package main

import (
    "bytes"
    "context"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
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

    // Define a small payload for POST.
    payload := []byte("Hello, CAOS!")
    contentType := "text/plain"

    // Perform POST /data.
    postURL := base.ResolveReference(&url.URL{Path: "/data"})
    addr, err := postData(postURL, payload, contentType)
    if err != nil {
        log.Fatalf("POST failed: %v", err)
    }

    // GET /data/:addr
    getURL := base.ResolveReference(&url.URL{Path: "/data/" + addr})
    if err := getData(getURL, payload); err != nil {
        log.Fatalf("GET failed: %v", err)
    }

    // HEAD /data/:addr
    if err := headData(getURL); err != nil {
        log.Fatalf("HEAD failed: %v", err)
    }

    // DELETE /data/:addr
    if err := deleteData(getURL); err != nil {
        log.Fatalf("DELETE failed: %v", err)
    }

    // GET after delete should return 404.
    if err := getDataNotFound(getURL); err != nil {
        log.Fatalf("GET after DELETE should 404: %v", err)
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
