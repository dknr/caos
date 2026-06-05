# Caos Integration Test Plan

## Goal

Provide a reusable integration test suite that validates any Caos server
implementation against the canonical OpenAPI spec (`openapi.yaml`). The suite
is designed to be run against the existing Deno implementation during
development, then reused verbatim for the new Go implementation to prove
equivalence.

---

## Test Infrastructure

### Server Lifecycle

Each test run manages its own Caos server process:

```
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│  test runner       │     │  caos server       │     │  temporary root    │
│  (Go test binary)  │ ──► │  (any impl)        │ ──► │  /tmp/caos-{pid}   │
│                    │     │                     │     │  (ephemeral)       │
└────────────────────┘     └────────────────────┘     └────────────────────┘
```

- Start the server with `--in-memory` or a fresh temp directory for isolation.
- Use `exec.Command` to spawn the server binary; capture stdout for log
  monitoring; kill on test cleanup.
- Wait for the "listening" log line before sending requests.
- Allocate a unique port per test run to avoid conflicts.

### Client

A lightweight Go HTTP client wrapping `net/http.Client`, driven by test cases
that reference the OpenAPI spec paths/methods.

### OpenAPI-Driven Validation

Each endpoint test:

1. Reads the corresponding OpenAPI path + method definition.
2. Builds a request matching the spec's `requestBody` schema.
3. Asserts the response status code matches the spec's `responses`.
4. Validates the response body against the spec's response `content` schema.
5. Checks that no undocumented status codes are returned.

A helper `validateAgainstSpec(method, path, statusCode, body)` performs step 4
using structural validation (JSON schema for JSON responses, type/pattern checks
for plain text responses).

---

## Test Cases

### 1. Root Redirect

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| R1 | GET | `/` | 302 | Redirects to configured home path |
| R2 | GET | `/` | — | `Location` header is non-empty |

### 2. Address Resolution — `/addr/{addr}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| A1 | GET | `/addr/{full_addr}` | 200 | Exact full address → single-element JSON array |
| A2 | GET | `/addr/{partial_6plus}` | 200 | Partial (≥6 chars) resolving to exactly one |
| A3 | GET | `/addr/{partial_ambiguous}` | 300 | Partial resolving to multiple addresses |
| A4 | GET | `/addr/{nonexistent}` | 404 | No match |
| A5 | GET | `/addr/{short}` | 404 | <6 chars (not enforced by addr handler, but worth documenting) |

**Setup**: POST data to create known objects with overlapping partial hashes.

### 3. Data Upload — `POST /data`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| D1 | POST | `/data` | 200 | Plain text → returns SHA-256 hash |
| D2 | POST | `/data` | 200 | Binary data → returns SHA-256 hash |
| D3 | POST | `/data` | 200 | Empty body → returns hash of empty content |

**Verify**: The returned hash matches `sha256(content)`. Re-fetch via GET
`/data/{hash}` and compare content.

### 4. Data Retrieval — `GET /data/{addr}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| DR1 | GET | `/data/{full_addr}` | 200 | Full address → exact content |
| DR2 | GET | `/data/{partial_unique}` | 200 | Partial (≥6 chars, unique) → exact content |
| DR3 | GET | `/data/{partial_ambiguous}` | 300 | Partial → JSON array of addresses |
| DR4 | GET | `/data/{nonexistent}` | 404 | Unknown address |
| DR5 | GET | `/data/{short}` | 404 | <6 chars |
| DR6 | GET | `/data/{addr}` | — | Content-Type matches the `type` tag if set |

**Setup**: POST data, then set a `type` tag before retrieving.

### 5. Tags — Get All `GET /tags/{addr}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| T1 | GET | `/tags/{addr}` | 200 | Returns JSON object of all tags |
| T2 | GET | `/tags/{addr_without_tags}` | 200 | Returns empty JSON object `{}` |

### 6. Tags — Get Single `GET /tags/{addr}/{tag}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| TS1 | GET | `/tags/{addr}/{existing_tag}` | 200 | Returns tag value as plain text |
| TS2 | GET | `/tags/{addr}/{missing_tag}` | 404 | Tag key not found |

### 7. Tags — Set `PUT /tags/{addr}/{tag}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| TP1 | PUT | `/tags/{addr}/{tag}` | 204 | Sets tag, body is plain text |
| TP2 | PUT | `/tags/{addr}/{tag}` | 204 | Overwrites existing tag |
| TP3 | PUT | `/tags/{missing_addr}/{tag}` | 404 | Address does not exist |
| TP4 | PUT | `/tags/{addr}/{tag}` | 400 | Empty body |

**Verify**: After TP1, GET the same tag returns the set value. After TP2, GET
returns the new value.

### 8. Tags — Delete `DELETE /tags/{addr}/{tag}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| TD1 | DELETE | `/tags/{addr}/{existing_tag}` | 204 | Deletes tag |
| TD2 | DELETE | `/tags/{addr}/{missing_tag}` | 204 | No-op for non-existent tag |
| TD3 | DELETE | `/tags/{missing_addr}/{tag}` | 204 | No-op for non-existent addr |

**Verify**: After TD1, GET the same tag returns 404.

### 9. Names — Resolve `GET /name/{name}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| N1 | GET | `/name/{existing_name}` | 200 | Returns address as text |
| N2 | GET | `/name/{existing_name}` | 302 | Also sets Location → /data/{addr} |
| N3 | GET | `/name/{missing_name}` | 404 | Name not set |

### 10. Names — Set `POST /name/{name}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| NS1 | POST | `/name/{name}` | 200 | Body is address text, response `"{name} {addr}"` |
| NS2 | POST | `/name/{name}` | 200 | Overwrites existing name |

**Verify**: After NS1, GET the name returns the address.

### 11. Path — Autoindex `GET /path/{addr}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| P1 | GET | `/path/{addr}` | 301 | No trailing slash → redirect with `/` |
| P2 | GET | `/path/{addr}/` | 200 | Returns autoindex HTML listing |
| P3 | GET | `/path/{addr}/` | 200 | If index.html exists with text/html type → serves it |
| P4 | GET | `/path/{missing_addr}` | 404 | Unknown address |
| P5 | GET | `/path/{non_path_addr}` | 404 | Object exists but type ≠ `caos/path` |

### 12. Path — File Serving `GET /path/{addr}/{name}`

| ID | Method | Path | Expected Status | Notes |
|---|---|---|---|---|
| PF1 | GET | `/path/{addr}/{existing_name}` | 200 | Serves file with correct Content-Type |
| PF2 | GET | `/path/{addr}/{missing_name}` | 404 | Name not in path index |
| PF3 | GET | `/path/{addr}/{name_ambiguous}` | 300 | File's address resolves to multiple |

---

## Go Test Structure

```
caos/
├── openapi.yaml                          ← canonical spec
├── internal/
│   └── test/
│       ├── main_test.go                  ← TestMain: server lifecycle
│       ├── suite.go                      ← test suite runner
│       ├── client.go                     ← typed HTTP client
│       ├── spec.go                       ← OpenAPI spec loader
│       ├── cases_addr_test.go            ← /addr/ tests
│       ├── cases_data_test.go            ← /data tests
│       ├── cases_tags_test.go            ← /tags tests
│       ├── cases_name_test.go            ← /name tests
│       └── cases_path_test.go            ← /path tests
└── cmd/
    └── caos/
        └── main.go                       ← server entry point
```

### TestMain Pattern

```go
func TestMain(m *testing.M) {
    port := pickFreePort()
    root := filepath.Join(os.TempDir(), "caos-test-"+strconv.Itoa(os.Getpid()))
    cmd := exec.Command("../caos", "serve", "--port", strconv.Itoa(port), "--root", root)
    // start, wait for ready, run tests, cleanup
    os.Exit(m.Run())
}
```

### Client

```go
type Client struct {
    base string  // e.g. "http://localhost:31923"
    hc   *http.Client
}

func (c *Client) AddData(r io.Reader) (addr string, err error)
func (c *Client) GetData(addr string) (io.ReadCloser, error)
func (c *Client) ResolveAddr(addr string) ([]string, error)
func (c *Client) GetTags(addr string) (map[string]string, error)
func (c *Client) GetTag(addr, tag string) (string, error)
func (c *Client) SetTag(addr, tag, value string) error
func (c *Client) DelTag(addr, tag string) error
func (c *Client) GetName(name string) (string, error)
func (c *Client) SetName(name, addr string) error
func (c *Client) GetPath(addr string) (string, error)
func (c *Client) GetPathFile(addr, name string) (string, error)
```

### Spec Validation Helper

```go
func ValidateAgainstSpec(t *testing.T, method, path, contentType string, statusCode int, body []byte)
```

This helper:
- Parses `openapi.yaml` once at test startup
- Looks up the expected responses for `method` + `path`
- Asserts the status code is listed in the spec
- For JSON responses: unmarshals and validates structure
- For plain text: checks pattern from spec's schema

### Test Suite Execution

Tests run in dependency order to build on each other's state:

```
Phase 1: Data upload      → establishes known addresses
Phase 2: Address resolve  → uses addresses from phase 1
Phase 3: Tags             → uses addresses from phase 1
Phase 4: Names            → uses addresses from phase 1
Phase 5: Path             → uses data+tags from phases 1-3
```

Each phase uses `t.Run(name, ...)` for isolation. Test functions are
`TestCaos_*` and share a `*Client` created in `TestMain`.

---

## CI Integration

The test suite runs against the Deno implementation during CI. When the Go
implementation is ready, CI runs the same suite against both:

```yaml
test-deno:  deno run --allow-all server/mod.ts & go test ./internal/test/...
test-go:    go build ./cmd/caos && go test ./internal/test/... -target=go
```

A `-target` flag in the test binary selects which server binary to start,
allowing side-by-side validation during migration.

---

## Edge Cases & Notes

1. **Concurrent access**: Tests run against a single server instance; the
   suite is not concurrent. Parallel execution is out of scope.
2. **Partial address minimum length**: The `addr` and `data` endpoints accept
   partial addresses ≥6 chars. Shorter inputs return 404 from the data
   handler. The addr handler doesn't enforce this — document but don't rely on.
3. **Tag deletion on missing addr**: The tags delete handler silently succeeds
   (204) even if the address doesn't exist. This is intentional but worth
   noting in a Go reimplementation as a potential footgun.
4. **Path index.html edge case**: If index.html exists but its type tag is not
   `text/html`, the server returns 404 instead of falling back to autoindex.
5. **Empty POST body for tags**: Returns 400. The data endpoint accepts empty
   bodies (returns a hash). These are distinct behaviors.
