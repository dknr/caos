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
//	Phase 0: Auth             → verifies API key protection (before any data)
//	Phase 1: Data upload      → establishes known addresses
//	Phase 2: Address resolve  → uses addresses from phase 1
//	Phase 3: Tags             → uses addresses from phase 1
//	Phase 4: Names            → uses addresses from phase 1
//	Phase 5: Path             → uses data+tags from phases 1-3
//	Phase 6: PushPull         → uses data+tags from phases 1-3
func RunSuite(t *testing.T, client *SuiteClient) {
	t.Run("Phase0_Auth", func(t *testing.T) { testAuthCases(t, client) })
	t.Run("Phase1_Data", func(t *testing.T) { testDataCases(t, client) })
	t.Run("Phase2_Addr", func(t *testing.T) { testAddrCases(t, client) })
	t.Run("Phase3_Tags", func(t *testing.T) { testTagCases(t, client) })
	t.Run("Phase4_Names", func(t *testing.T) { testNameCases(t, client) })
	t.Run("Phase5_Path", func(t *testing.T) { testPathCases(t, client) })
	t.Run("Phase6_PushPull", func(t *testing.T) { testPushPullCases(t, client) })
}

// testAuthCases tests API key authentication for write endpoints.
func testAuthCases(t *testing.T, c *SuiteClient) {
	t.Helper()

	// First upload data (with auth) to have addresses for read/write tests
	content := "auth test object"
	addr, err := c.AddData(strings.NewReader(content))
	if err != nil {
		t.Fatalf("Setup upload failed: %v", err)
	}

	t.Run("ReadOpsWithoutAuth", func(t *testing.T) {
		// GET /data/{addr} should work without auth
		status, err := c.ReadRequest("GET", "/data/"+addr)
		if err != nil {
			t.Fatalf("GET /data failed: %v", err)
		}
		if status != 200 {
			t.Fatalf("Expected 200 for GET /data without auth, got %d", status)
		}

		// GET /addr/{addr} should work without auth
		status, err = c.ReadRequest("GET", "/addr/"+addr)
		if err != nil {
			t.Fatalf("GET /addr failed: %v", err)
		}
		if status != 200 {
			t.Fatalf("Expected 200 for GET /addr without auth, got %d", status)
		}

		// GET /tags/{addr} should work without auth
		status, err = c.ReadRequest("GET", "/tags/"+addr)
		if err != nil {
			t.Fatalf("GET /tags failed: %v", err)
		}
		if status != 200 {
			t.Fatalf("Expected 200 for GET /tags without auth, got %d", status)
		}

		// GET /name/{name} should work without auth (name not set yet, expect 404 not 401)
		status, err = c.ReadRequest("GET", "/name/nonexistent")
		if err != nil {
			t.Fatalf("GET /name failed: %v", err)
		}
		if status == 401 {
			t.Fatalf("Expected non-401 for GET /name without auth, got 401")
		}
	})

	t.Run("PostDataWithoutAuth", func(t *testing.T) {
		status, err := c.AddDataNoAuth(strings.NewReader("no auth data"))
		if err != nil {
			t.Fatalf("POST /data no auth failed: %v", err)
		}
		if status != 401 {
			t.Fatalf("Expected 401 for POST /data without auth, got %d", status)
		}
	})

	t.Run("PostDataWrongKey", func(t *testing.T) {
		status, err := c.AddDataWithWrongKey(strings.NewReader("wrong key data"))
		if err != nil {
			t.Fatalf("POST /data wrong key failed: %v", err)
		}
		if status != 401 {
			t.Fatalf("Expected 401 for POST /data with wrong key, got %d", status)
		}
	})

	t.Run("PostDataWithCorrectKey", func(t *testing.T) {
		// Verify the SuiteClient (which has the correct key) can still write
		newAddr, err := c.AddData(strings.NewReader("correct key data"))
		if err != nil {
			t.Fatalf("POST /data with correct key failed: %v", err)
		}
		if len(newAddr) != 64 {
			t.Fatalf("Expected 64-char hash, got %q", newAddr)
		}
	})

	t.Run("PutTagWithoutAuth", func(t *testing.T) {
		status, err := c.SetTagNoAuth(addr, "testtag", "testval")
		if err != nil {
			t.Fatalf("PUT /tags no auth failed: %v", err)
		}
		if status != 401 {
			t.Fatalf("Expected 401 for PUT /tags without auth, got %d", status)
		}
	})

	t.Run("PutTagWrongKey", func(t *testing.T) {
		status, err := c.WriteRequestWrongKey("PUT", "/tags/"+addr+"/testtag", strings.NewReader("testval"))
		if err != nil {
			t.Fatalf("PUT /tags wrong key failed: %v", err)
		}
		if status != 401 {
			t.Fatalf("Expected 401 for PUT /tags with wrong key, got %d", status)
		}
	})

	t.Run("PutTagWithCorrectKey", func(t *testing.T) {
		if err := c.SetTag(addr, "type", "text/plain"); err != nil {
			t.Fatalf("PUT /tags with correct key failed: %v", err)
		}
	})

	t.Run("DeleteTagWithoutAuth", func(t *testing.T) {
		status, err := c.DelTagNoAuth(addr, "testtag")
		if err != nil {
			t.Fatalf("DELETE /tags no auth failed: %v", err)
		}
		if status != 401 {
			t.Fatalf("Expected 401 for DELETE /tags without auth, got %d", status)
		}
	})

	t.Run("DeleteTagWrongKey", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", c.base+"/tags/"+addr+"/testtag", nil)
		if err != nil {
			t.Fatalf("DELETE request failed: %v", err)
		}
		req.Header.Set("X-API-Key", "wrong-key")
		resp, err := c.hc.Do(req)
		if err != nil {
			t.Fatalf("DELETE request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != 401 {
			t.Fatalf("Expected 401 for DELETE /tags with wrong key, got %d", resp.StatusCode)
		}
	})

	t.Run("DeleteTagWithCorrectKey", func(t *testing.T) {
		if err := c.DelTag(addr, "testtag"); err != nil {
			t.Fatalf("DELETE /tags with correct key failed: %v", err)
		}
	})

	t.Run("PostNameWithoutAuth", func(t *testing.T) {
		status, err := c.SetNameNoAuth("myname", addr)
		if err != nil {
			t.Fatalf("POST /name no auth failed: %v", err)
		}
		if status != 401 {
			t.Fatalf("Expected 401 for POST /name without auth, got %d", status)
		}
	})

	t.Run("PostNameWrongKey", func(t *testing.T) {
		status, err := c.WriteRequestWrongKey("POST", "/name/myname", strings.NewReader(addr))
		if err != nil {
			t.Fatalf("POST /name wrong key failed: %v", err)
		}
		if status != 401 {
			t.Fatalf("Expected 401 for POST /name with wrong key, got %d", status)
		}
	})

	t.Run("PostNameWithCorrectKey", func(t *testing.T) {
		if err := c.SetName("myname", addr); err != nil {
			t.Fatalf("POST /name with correct key failed: %v", err)
		}
	})
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

		// Verify size tag is stored
		tags, err := c.GetTags(addr)
		if err != nil {
			t.Fatalf("GetTags failed: %v", err)
		}
		if tags["size"] == "" {
			t.Errorf("Expected size tag to be set, got tags=%v", tags)
		}

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

		// Verify size tag is stored
		tags, err := c.GetTags(addr)
		if err != nil {
			t.Fatalf("GetTags failed: %v", err)
		}
		if tags["size"] == "" {
			t.Errorf("Expected size tag to be set, got tags=%v", tags)
		}

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
		// First set a custom tag (not the protected 'type' tag)
		if err := c.SetTag(addr, "customtag", "somevalue"); err != nil {
			t.Fatalf("SetTag failed: %v", err)
		}
		if err := c.DelTag(addr, "customtag"); err != nil {
			t.Fatalf("DelTag failed: %v", err)
		}
		// Verify it's gone
		_, err := c.GetTag(addr, "customtag")
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
		if err := c.DelTag("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "customtag"); err != nil {
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

	t.Run("PullAbsolutePath", func(t *testing.T) {
		// Create a path object with an absolute path entry
		fileAddr, err := c.AddData(strings.NewReader("file content"))
		if err != nil {
			t.Fatalf("AddData failed: %v", err)
		}
		pathContent := fileAddr + " /etc/evil\n"
		pathAddr, err := c.AddData(strings.NewReader(pathContent))
		if err != nil {
			t.Fatalf("AddData failed: %v", err)
		}
		if err := c.SetTag(pathAddr, "type", "caos/path"); err != nil {
			t.Fatalf("SetTag failed: %v", err)
		}
		err = c.PullAddr(pathAddr, "/tmp")
		if err == nil || !strings.Contains(err.Error(), "absolute path") {
			t.Fatalf("Expected error rejecting absolute path, got %v", err)
		}
	})

	t.Run("PullPathTraversal", func(t *testing.T) {
		// Create a path object with a relative path traversal entry
		fileAddr, err := c.AddData(strings.NewReader("file content"))
		if err != nil {
			t.Fatalf("AddData failed: %v", err)
		}
		pathContent := fileAddr + " ../../etc/evil\n"
		pathAddr, err := c.AddData(strings.NewReader(pathContent))
		if err != nil {
			t.Fatalf("AddData failed: %v", err)
		}
		if err := c.SetTag(pathAddr, "type", "caos/path"); err != nil {
			t.Fatalf("SetTag failed: %v", err)
		}
		err = c.PullAddr(pathAddr, "/tmp")
		if err == nil || !strings.Contains(err.Error(), "traversal") {
			t.Fatalf("Expected error rejecting path traversal, got %v", err)
		}
	})
}
