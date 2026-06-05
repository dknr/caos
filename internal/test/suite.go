package test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// RunSuite runs the full integration test suite.
// Tests run in dependency order:
//
//	Phase 1: Data upload      → establishes known addresses
//	Phase 2: Address resolve  → uses addresses from phase 1
//	Phase 3: Tags             → uses addresses from phase 1
//	Phase 4: Names            → uses addresses from phase 1
//	Phase 5: Path             → uses data+tags from phases 1-3
func RunSuite(t *testing.T, client *SuiteClient) {
	t.Run("Phase1_Data", func(t *testing.T) { testDataCases(t, client) })
	t.Run("Phase2_Addr", func(t *testing.T) { testAddrCases(t, client) })
	t.Run("Phase3_Tags", func(t *testing.T) { testTagCases(t, client) })
	t.Run("Phase4_Names", func(t *testing.T) { testNameCases(t, client) })
	t.Run("Phase5_Path", func(t *testing.T) { testPathCases(t, client) })
	t.Run("Phase6_PushPull", func(t *testing.T) { testPushPullCases(t, client) })
}

// testDataCases tests POST /data and GET /data/{addr}.
func testDataCases(t *testing.T, c *SuiteClient) {
	t.Helper()

	t.Run("UploadPlainText", func(t *testing.T) {
		content := "hello world"
		r := strings.NewReader(content)
		addr, err := c.AddData(r)
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}
		if len(addr) != 64 {
			t.Fatalf("Expected 64-char hash, got %q", addr)
		}
		// Verify hash
		h := sha256.Sum256([]byte(content))
		expected := hex.EncodeToString(h[:])
		if addr != expected {
			t.Fatalf("Hash mismatch: got %s, want %s", addr, expected)
		}
		t.Logf("Uploaded: %s", addr)

		// Retrieve by full address
		rc, _, status, err := c.GetDataWithStatus(addr)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		defer rc.Close()
		if status != 200 {
			t.Fatalf("Expected 200, got %d", status)
		}
		data, _ := io.ReadAll(rc)
		if string(data) != content {
			t.Fatalf("Content mismatch: got %q, want %q", string(data), content)
		}
	})

	t.Run("UploadBinary", func(t *testing.T) {
		binary := []byte{0xFF, 0xFE, 0x00, 0x01, 0x02, 0x03}
		addr, err := c.AddData(bytes.NewReader(binary))
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}
		if len(addr) != 64 {
			t.Fatalf("Expected 64-char hash, got %q", addr)
		}
		t.Logf("Uploaded binary: %s", addr)

		// Verify content
		rc, _, _, err := c.GetDataWithStatus(addr)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		defer rc.Close()
		data, _ := io.ReadAll(rc)
		if !bytes.Equal(data, binary) {
			t.Fatalf("Binary content mismatch")
		}
	})

	t.Run("UploadEmpty", func(t *testing.T) {
		addr, err := c.AddData(strings.NewReader(""))
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}
		if len(addr) != 64 {
			t.Fatalf("Expected 64-char hash, got %q", addr)
		}
		t.Logf("Uploaded empty: %s", addr)
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		_, _, status, err := c.GetDataWithStatus("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if status != 404 {
			t.Fatalf("Expected 404, got %d", status)
		}
	})

	t.Run("GetShortAddr", func(t *testing.T) {
		_, _, status, err := c.GetDataWithStatus("abc")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if status != 404 {
			t.Fatalf("Expected 404, got %d", status)
		}
	})
}

// testAddrCases tests GET /addr/{addr}.
func testAddrCases(t *testing.T, c *SuiteClient) {
	t.Helper()

	// First upload data to have known addresses
	addr1, err := c.AddData(strings.NewReader("first object"))
	if err != nil {
		t.Fatalf("Setup upload failed: %v", err)
	}
	addr2, err := c.AddData(strings.NewReader("second object"))
	if err != nil {
		t.Fatalf("Setup upload failed: %v", err)
	}

	t.Run("ExactFullAddress", func(t *testing.T) {
		addrs, err := c.ResolveAddr(addr1)
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if len(addrs) != 1 || addrs[0] != addr1 {
			t.Fatalf("Expected [%s], got %v", addr1, addrs)
		}
	})

	t.Run("PartialAddress", func(t *testing.T) {
		addrs, err := c.ResolveAddr(addr1[:10])
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if len(addrs) != 1 || addrs[0] != addr1 {
			t.Fatalf("Expected [%s], got %v", addr1, addrs)
		}
	})

	t.Run("NonExistent", func(t *testing.T) {
		_, status, err := c.ResolveAddrWithStatus("abcdef1234")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if status != 404 {
			t.Fatalf("Expected 404, got %d", status)
		}
	})

	t.Run("ShortInput", func(t *testing.T) {
		_, status, err := c.ResolveAddrWithStatus("abc")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if status != 404 {
			t.Fatalf("Expected 404, got %d", status)
		}
	})

	_ = addr2
}

// testTagCases tests tag endpoints.
func testTagCases(t *testing.T, c *SuiteClient) {
	t.Helper()

	addr, err := c.AddData(strings.NewReader("tag test object"))
	if err != nil {
		t.Fatalf("Setup upload failed: %v", err)
	}

	t.Run("SetTag", func(t *testing.T) {
		if err := c.SetTag(addr, "type", "text/plain"); err != nil {
			t.Fatalf("SetTag failed: %v", err)
		}
	})

	t.Run("GetTag", func(t *testing.T) {
		val, err := c.GetTag(addr, "type")
		if err != nil {
			t.Fatalf("GetTag failed: %v", err)
		}
		if val != "text/plain" {
			t.Fatalf("Expected 'text/plain', got %q", val)
		}
	})

	t.Run("GetAllTags", func(t *testing.T) {
		tags, err := c.GetTags(addr)
		if err != nil {
			t.Fatalf("GetTags failed: %v", err)
		}
		if tags["type"] != "text/plain" {
			t.Fatalf("Expected type=text/plain in tags, got %v", tags)
		}
	})

	t.Run("GetMissingTag", func(t *testing.T) {
		_, err := c.GetTag(addr, "nonexistent")
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Fatalf("Expected 404 error, got %v", err)
		}
	})

	t.Run("SetTagNonExistentAddr", func(t *testing.T) {
		err := c.SetTag("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "type", "text/plain")
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Fatalf("Expected 404 error, got %v", err)
		}
	})

	t.Run("SetTagEmptyBody", func(t *testing.T) {
		err := c.SetTag(addr, "empty", "")
		if err == nil || !strings.Contains(err.Error(), "400") {
			t.Fatalf("Expected 400 error for empty body, got %v", err)
		}
	})

	t.Run("DeleteTag", func(t *testing.T) {
		if err := c.DelTag(addr, "type"); err != nil {
			t.Fatalf("DelTag failed: %v", err)
		}
		// Verify it's gone
		_, err := c.GetTag(addr, "type")
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Fatalf("Expected 404 after delete, got %v", err)
		}
	})

	t.Run("DeleteMissingTag", func(t *testing.T) {
		if err := c.DelTag(addr, "nonexistent"); err != nil {
			t.Fatalf("DelTag non-existent failed: %v", err)
		}
	})

	t.Run("DeleteTagNonExistentAddr", func(t *testing.T) {
		if err := c.DelTag("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "type"); err != nil {
			t.Fatalf("DelTag on missing addr failed: %v", err)
		}
	})
}

// testNameCases tests name endpoints.
func testNameCases(t *testing.T, c *SuiteClient) {
	t.Helper()

	addr, err := c.AddData(strings.NewReader("name test object"))
	if err != nil {
		t.Fatalf("Setup upload failed: %v", err)
	}

	t.Run("SetName", func(t *testing.T) {
		if err := c.SetName("myfile", addr); err != nil {
			t.Fatalf("SetName failed: %v", err)
		}
	})

	t.Run("GetName", func(t *testing.T) {
		resolved, err := c.GetName("myfile")
		if err != nil {
			t.Fatalf("GetName failed: %v", err)
		}
		if resolved != addr {
			t.Fatalf("Expected %s, got %s", addr, resolved)
		}
	})

	t.Run("LocationHeader", func(t *testing.T) {
		resp, err := http.Get(c.base + "/name/" + PathEscape("myfile"))
		if err != nil {
			t.Fatalf("GET /name/myfile failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		loc := resp.Header.Get("Location")
		if loc == "" {
			t.Errorf("Expected Location header to be set")
		} else if !strings.HasPrefix(loc, "/data/") {
			t.Errorf("Expected Location to start with /data/, got %q", loc)
		}
	})

	t.Run("MissingName", func(t *testing.T) {
		_, err := c.GetName("nonexistent")
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Fatalf("Expected 404 error, got %v", err)
		}
	})

	t.Run("OverwriteName", func(t *testing.T) {
		newAddr, err := c.AddData(strings.NewReader("overwrite test"))
		if err != nil {
			t.Fatalf("Setup upload failed: %v", err)
		}
		if err := c.SetName("myfile", newAddr); err != nil {
			t.Fatalf("SetName overwrite failed: %v", err)
		}
		resolved, err := c.GetName("myfile")
		if err != nil {
			t.Fatalf("GetName after overwrite failed: %v", err)
		}
		if resolved != newAddr {
			t.Fatalf("Expected %s after overwrite, got %s", newAddr, resolved)
		}
	})
}

// testPathCases tests path endpoints.
func testPathCases(t *testing.T, c *SuiteClient) {
	t.Helper()

	// Set up real content for path entries
	indexContent := []byte("<html><body>hello</body></html>")
	indexAddr, err := c.AddData(bytes.NewReader(indexContent))
	if err != nil {
		t.Fatalf("Upload index.html failed: %v", err)
	}
	// Tag as text/html
	if err := c.SetTag(indexAddr, "type", "text/html"); err != nil {
		t.Fatalf("Set index type tag failed: %v", err)
	}

	readmeContent := []byte("readme content")
	readmeAddr, err := c.AddData(bytes.NewReader(readmeContent))
	if err != nil {
		t.Fatalf("Upload readme.txt failed: %v", err)
	}

	// Path object WITH index.html (for index.html and file serving tests)
	pathWithIndex := indexAddr + " index.html\n" + readmeAddr + " readme.txt\n"
	pathWithIndexAddr, err := c.AddData(strings.NewReader(pathWithIndex))
	if err != nil {
		t.Fatalf("Upload path with index failed: %v", err)
	}
	if err := c.SetTag(pathWithIndexAddr, "type", "caos/path"); err != nil {
		t.Fatalf("Set path type tag failed: %v", err)
	}

	// Path object WITHOUT index.html (for autoindex test)
	pathNoIndex := readmeAddr + " readme.txt\n"
	pathNoIndexAddr, err := c.AddData(strings.NewReader(pathNoIndex))
	if err != nil {
		t.Fatalf("Upload path without index failed: %v", err)
	}
	if err := c.SetTag(pathNoIndexAddr, "type", "caos/path"); err != nil {
		t.Fatalf("Set path type tag failed: %v", err)
	}

	t.Run("TrailingSlashRedirect", func(t *testing.T) {
		// Use a client that doesn't follow redirects
		hc := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		resp, err := hc.Get(c.base + "/path/" + pathNoIndexAddr)
		if err != nil {
			t.Fatalf("GET /path/{addr} failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 301 {
			t.Fatalf("Expected 301 redirect, got %d", resp.StatusCode)
		}
		loc := resp.Header.Get("Location")
		if loc == "" {
			t.Errorf("Expected Location header")
		}
	})

	t.Run("UnknownAddr", func(t *testing.T) {
		_, status, err := c.GetPathWithStatus("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		if err != nil {
			t.Fatalf("GetPath failed: %v", err)
		}
		if status != 404 {
			t.Fatalf("Expected 404, got %d", status)
		}
	})

	t.Run("NonPathObject", func(t *testing.T) {
		plainAddr, err := c.AddData(strings.NewReader("plain text"))
		if err != nil {
			t.Fatalf("Upload plain data failed: %v", err)
		}
		// Tag with a non-path type
		if err := c.SetTag(plainAddr, "type", "text/plain"); err != nil {
			t.Fatalf("Set type tag failed: %v", err)
		}
		_, status, err := c.GetPathWithStatus(plainAddr)
		if err != nil {
			t.Fatalf("GetPath failed: %v", err)
		}
		if status != 404 {
			t.Fatalf("Expected 404 for non-path object, got %d", status)
		}
	})

	t.Run("Autoindex", func(t *testing.T) {
		body, status, err := c.GetPathWithStatus(pathNoIndexAddr)
		if err != nil {
			t.Fatalf("GetPath failed: %v", err)
		}
		if status != 200 {
			t.Fatalf("Expected 200, got %d", status)
		}
		if !strings.Contains(body, "readme.txt") {
			t.Fatalf("Autoindex missing readme.txt entry: %s", body)
		}
	})

	t.Run("IndexHTML", func(t *testing.T) {
		resp, err := http.Get(c.base + "/path/" + pathWithIndexAddr + "/")
		if err != nil {
			t.Fatalf("GET /path/{addr}/ failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("Expected 200 serving index.html, got %d", resp.StatusCode)
		}
		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("Expected text/html content type, got %q", ct)
		}
		body, _ := io.ReadAll(resp.Body)
		if !bytes.Equal(body, indexContent) {
			t.Fatalf("Expected index.html content, got %q", string(body))
		}
	})

	t.Run("FileServing", func(t *testing.T) {
		resp, err := http.Get(c.base + "/path/" + pathWithIndexAddr + "/readme.txt")
		if err != nil {
			t.Fatalf("GET /path/{addr}/readme.txt failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if !bytes.Equal(body, readmeContent) {
			t.Fatalf("Expected readme content, got %q", string(body))
		}
	})

	t.Run("MissingFile", func(t *testing.T) {
		resp, err := http.Get(c.base + "/path/" + pathWithIndexAddr + "/nonexistent.txt")
		if err != nil {
			t.Fatalf("GET /path/{addr}/nonexistent.txt failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 404 {
			t.Fatalf("Expected 404, got %d", resp.StatusCode)
		}
	})
}

// testPushPullCases tests the push and pull operations.
func testPushPullCases(t *testing.T, c *SuiteClient) {
	t.Helper()

	// Create a temp directory with some test files
	tmpDir, err := os.MkdirTemp("", "caos-push-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a few files in a directory structure
	if err := os.MkdirAll(tmpDir+"/sub", 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := os.WriteFile(tmpDir+"/hello.txt", []byte("hello world"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(tmpDir+"/sub/data.bin", []byte{0x00, 0x01, 0x02}, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	t.Run("PushDirectory", func(t *testing.T) {
		pathAddr, count, err := c.PushDir(tmpDir)
		if err != nil {
			t.Fatalf("PushDir failed: %v", err)
		}
		if count != 2 {
			t.Fatalf("Expected 2 files, got %d", count)
		}
		if len(pathAddr) != 64 {
			t.Fatalf("Expected 64-char path address, got %q", pathAddr)
		}
		t.Logf("PushDir pathAddr=%s count=%d", pathAddr[:8], count)

		// Verify the path object is tagged correctly
		typeVal, err := c.GetTag(pathAddr, "type")
		if err != nil {
			t.Fatalf("GetTag type failed: %v", err)
		}
		if typeVal != "caos/path" {
			t.Fatalf("Expected type=caos/path, got %q", typeVal)
		}
	})

	t.Run("PullDirectory", func(t *testing.T) {
		// First push
		pathAddr, _, err := c.PushDir(tmpDir)
		if err != nil {
			t.Fatalf("PushDir failed: %v", err)
		}

		// Pull to a different directory
		pullDir, err := os.MkdirTemp("", "caos-pull-test-*")
		if err != nil {
			t.Fatalf("Failed to create pull dir: %v", err)
		}
		defer os.RemoveAll(pullDir)

		if err := c.PullAddr(pathAddr, pullDir); err != nil {
			t.Fatalf("PullAddr failed: %v", err)
		}

		// Verify files exist and content matches
		helloContent, err := os.ReadFile(pullDir + "/hello.txt")
		if err != nil {
			t.Fatalf("ReadFile hello.txt failed: %v", err)
		}
		if string(helloContent) != "hello world" {
			t.Fatalf("Expected 'hello world', got %q", string(helloContent))
		}

		binContent, err := os.ReadFile(pullDir + "/sub/data.bin")
		if err != nil {
			t.Fatalf("ReadFile sub/data.bin failed: %v", err)
		}
		if len(binContent) != 3 || binContent[0] != 0x00 {
			t.Fatalf("Expected 3 bytes starting with 0x00, got %v", binContent)
		}
	})

	t.Run("PullInvalidAddr", func(t *testing.T) {
		err := c.PullAddr("abcdef1234", "/tmp")
		if err == nil {
			t.Fatalf("Expected error for invalid addr, got nil")
		}
	})

	t.Run("PullNonPathObject", func(t *testing.T) {
		// Upload a plain file and try to pull it
		addr, err := c.AddData(strings.NewReader("not a path"))
		if err != nil {
			t.Fatalf("AddData failed: %v", err)
		}
		err = c.PullAddr(addr, "/tmp")
		if err == nil {
			t.Fatalf("Expected error for non-path object, got nil")
		}
	})
}
