# Manual Test Plan – CAOS Level 0 Compliance

**Purpose**  
Verify that the CAOS server implementation satisfies the Level 0 API specification (`CAOS‑API‑LEVEL0.md`). The plan covers the required endpoints, address format, content‑type handling, error codes, tagging, binary data, and idempotency.

**Prerequisites**  
- Go 1.22+ installed (to build the server).  
- `curl` or any HTTP client (e.g., HTTPie, Postman).  
- A temporary directory for server storage (will be created automatically).  

**Setup**  
1. Build the server binary (if not already built):  
   ```bash
   cd /path/to/caos
   go build -o caos
   ```
2. Start the server on a known port (default 31923):  
   ```bash
   ./caos serve :31923 &
   ```
   Note the PID; the server logs “Starting CAOS server on :31923”.  
3. Wait a second for the server to bind.  
4. Define a base URL variable for convenience:  
   ```bash
   BASE=http://localhost:31923
   ```
5. (Optional) Clean storage between test runs:  
   ```bash
   rm -rf ./caos-store
   ```

**Test Categories**  
Each test case includes:  
- **Request** (method, URL, headers, body)  
- **Expected Status Code**  
- **Expected Response Body/Headers**  
- **Pass/Fail Criteria**  

---

## 1. Basic Compliance (Happy Path)

| # | Description | Request | Expected |
|---|-------------|---------|----------|
| 1.1 | Store simple text | `POST $BASE/data`<br>Header: `Content-Type: text/plain`<br>Body: `Hello, CAOS!` | `200 OK`<br>Body: 64‑char lower‑case hex address<br>Header: `Content-Type: text/plain` |
| 1.2 | Retrieve stored data | `GET $BASE/data/<addr_from_1.1>` | `200 OK`<br>Header: `Content-Type: text/plain`<br>Body: `Hello, CAOS!` |
| 1.3 | HEAD metadata | `HEAD $BASE/data/<addr_from_1.1>` | `200 OK`<br>Headers: `Content-Type: text/plain`, `Content-Length: 13` (byte length of body) |
| 1.4 | Delete data | `DELETE $BASE/data/<addr_from_1.1>` | `204 No Content` (or `200 OK` – both acceptable) |
| 1.5 | GET after delete → 404 | `GET $BASE/data/<addr_from_1.1>` | `404 Not Found` |
| 1.6 | Store binary data | `POST $BASE/data`<br>Header: `Content-Type: application/octet-stream`<br>Body: raw bytes `\x00\x01\x02\x03` | `200 OK`<br>Body: 64‑char address |
| 1.7 | Retrieve binary data | `GET $BASE/data/<addr_from_1.6>` | `200 OK`<br>Header: `Content-Type: application/octet-stream`<br>Body: `\x00\x01\x02\x03` |

**Pass** if all requests return the expected status codes and bodies/headers match.

---

## 2. Address Format Validation

| # | Description | Request | Expected |
|---|-------------|---------|----------|
| 2.1 | GET with too‑short address | `GET $BASE/data/abc` | `400 Bad Request` |
| 2.2 | GET with non‑hex characters | `GET $BASE/data/zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz` | `400 Bad Request` |
| 2.3 | GET with uppercase hex (valid) | `GET $BASE/data/ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF` | `404 Not Found` (address format valid but not stored) |
| 2.4 | DELETE with malformed address | `DELETE $BASE/data/123` | `400 Bad Request` |
| 2.5 | HEAD with malformed address | `HEAD $BASE/data/123` | `400 Bad Request` |

**Pass** if each malformed request yields `400 Bad Request`. Valid‑format but unknown address yields `404`.

---

## 3. Content‑Type Handling

| # | Description | Request | Expected |
|---|-------------|---------|----------|
| 3.1 | POST without Content-Type (should default) | `POST $BASE/data`<br>Body: `test` | `200 OK`<br>Body: address |
| 3.2 | Verify default type on GET | `GET $BASE/data/<addr_from_3.1>` | `200 OK`<br>Header: `Content-Type: application/octet-stream` |
| 3.3 | POST with explicit text/plain | `POST $BASE/data`<br>Header: `Content-Type: text/plain`<br>Body: `utf8 text` | `200 OK` |
| 3.4 | Retrieve with correct type | `GET $BASE/data/<addr_from_3.3>` | `200 OK`<br>Header: `Content-Type: text/plain` |
| 3.5 | POST with invalid MIME (no slash) | `POST $BASE/data`<br>Header: `Content-Type: invalid`<br>Body: `x` | `400 Bad Request` |
| 3.6 | POST with incomplete MIME (missing subtype) | `POST $BASE/data`<br>Header: `Content-Type: text/`<br>Body: `x` | **Current behavior:** `200 OK` (accepts). **Spec discussion point** – note if this should be rejected. |
| 3.7 | POST with text/* and non‑UTF‑8 payload (invalid after first 512 bytes) | Create a 600‑byte payload: first 512 bytes valid UTF‑8, remaining bytes `0xC0 0x80` (invalid UTF‑8).<br>`POST $BASE/data`<br>Header: `Content-Type: text/plain`<br>Body: <payload> | **Current behavior:** `200 OK` (only first 512 bytes checked). **Spec discussion point** – note if full validation expected. |
| 3.8 | POST with binary type and arbitrary data (no validation) | `POST $BASE/data`<br>Header: `Content-Type: image/png`<br>Body: `not a png` | `200 OK`<br>Later GET returns same body and `Content-Type: image/png`. (Server does not verify actual PNG validity.) |

**Pass** for cases 3.1‑3.4 and 3.5 (400). Cases 3.6‑3.8 document current behavior; QA should record whether the observed behavior matches the team’s interpretation of the spec.

---

## 4. Error Codes

| # | Description | Request | Expected |
|---|-------------|---------|----------|
| 4.1 | Internal error (simulate by causing store failure) – not easily triggered without mocking; rely on existing coverage. | N/A | N/A |
| 4.2 | DELETE on non‑existent address | `DELETE $BASE/data/<random‑64‑hex>` | `204 No Content` (or `200 OK`) |
| 4.3 | HEAD on non‑existent address | `HEAD $BASE/data/<random‑64‑hex>` | `404 Not Found` |
| 4.4 | GET on non‑existent address | `GET $BASE/data/<random‑64‑hex>` | `404 Not Found` |
| 4.5 | POST that causes store to return error (e.g., disk full) – out of scope for manual test. | N/A | N/A |

**Pass** if status codes match expectations.

---

## 5. Tagging (Size & Type)

| # | Description | Request | Expected |
|---|-------------|---------|----------|
| 5.1 | Store data, check size tag via HEAD | `POST $BASE/data`<br>Body: `12345` (5 bytes)<br>`HEAD $BASE/data/<addr>` | `200 OK`<br>Headers: `Content-Type: application/octet-stream` (default)<br>`Content-Length: 5` |
| 5.2 | Store with explicit type, verify via GET | `POST $BASE/data`<br>Header: `Content-Type: application/json`<br>Body: `{"foo":1}`<br>`GET $BASE/data/<addr>` | `200 OK`<br>Header: `Content-Type: application/json`<br>Body: `{"foo":1}` |
| 5.3 | Update type not allowed (no endpoint) – N/A for Level 0. | N/A | N/A |

**Pass** if headers reflect the provided size and type.

---

## 6. Binary Data & Large Payloads

| # | Description | Request | Expected |
|---|-------------|---------|----------|
| 6.1 | Store 1 MiB random binary | `POST $BASE/data`<br>Header: `Content-Type: application/octet-stream`<br>Body: `<1 MiB random>` | `200 OK`<br>Body: address |
| 6.2 | Retrieve and compare | `GET $BASE/data/<addr>` → save to file, compare byte‑wise with original | Identical |
| 6.3 | Store 5 MiB payload (as used in automated tests) | Same as 6.1 with 5 MiB | `200 OK` |
| 6.4 | Verify address length | Length of returned body = 64 chars | Pass |

**Pass** if data round‑trips correctly and server does not crash or truncate.

---

## 7. Idempotency of POST

| # | Description | Request | Expected |
|---|-------------|---------|----------|
| 7.1 | POST same payload twice | `POST $BASE/data`<br>Body: `repeat me` → capture addr1<br>Repeat same POST → capture addr2 | Both responses `200 OK`<br>addr1 == addr2 (same hex string) |
| 7.2 | Verify metadata not changed on second POST | After first POST, optionally HEAD to see size/type; after second POST, HEAD should return same values. | No change. |

**Pass** if both calls succeed with identical address and metadata unchanged.

---

## 8. Concurrency (Basic)

| # | Description | Request | Expected |
|---|-------------|---------|----------|
| 8.1 | Launch 5 concurrent POSTs with identical payload | Use a script or tool (e.g., `ab`, `wrk`, or simple bash loop with `&`) to POST same body simultaneously. | All responses `200 OK` and all return the **same** address. |
| 8.2 | Verify stored data is correct | GET that address → compare to payload. | Matches. |

**Pass** if all succeed and address is unique across calls.

---

## 9. Cleanup

After testing, stop the server:  
```bash
kill <server‑pid>
```
Optionally remove storage directory:  
```bash
rm -rf ./caos-store
```

---

**Pass/Fail Summary**  
- A test **passes** if the actual response matches the expected status code, headers, and body as defined.  
- Any deviation is a **fail** and should be logged with the request details and observed response.  
- For the discussion points (3.6, 3.7, 3.8, 7.1‑7.2 regarding MetaStore behavior), record the observed behavior; the team will decide if it conforms to the spec or requires adjustment.

**Notes for QA**  
- Use a fresh storage directory for each test run to avoid cross‑test interference.  
- When testing large payloads, ensure the client does not buffer the entire payload in memory if using shell utilities; `curl --data-binary @file` works fine.  
- The server logs to stdout; watch for unexpected panic messages (none should appear).  
- All address strings returned by the server must be lower‑case hex; the acceptance regex allows upper‑case, but the implementation produces lower‑case.  

---  

*End of Test Plan.*